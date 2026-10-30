package http

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
	"time"

	"ts6-viewer/internal/config"
	"ts6-viewer/internal/ts6"
)

// testStatsCfg returns a config pointing at a refused local port: every SSH
// dial fails fast (connection refused), so tests that start the background
// sampler exercise the "fetch failed, state unchanged" path without touching
// a real server.
func testStatsCfg() *config.Config {
	cfg := &config.Config{}
	cfg.Teamspeak6.Host = "127.0.0.1"
	cfg.Teamspeak6.Port = "1" // closed port: dial is refused immediately
	cfg.Teamspeak6.ServerID = "1"
	return cfg
}

// TestRecordStatisticsFoldsAndPersists pins the core contract shared by the
// background sampler and the viewer path: one observation is folded into the
// graph (sampleHistory), then the brief-visit safety net (mergeClientDB),
// then the activity log (updateActivity), and finally persisted to the state
// file — and the reconciled list is returned so the caller can render it
// without a second pass.
func TestRecordStatisticsFoldsAndPersists(t *testing.T) {
	saveStateGlobals(t)
	now := time.Now().Unix()

	snap := &statisticsSnapshot{
		clients: []ts6.Client{
			activityClientAt("Live", "uid-live", now-120),
		},
		// Brief visitor: only in the client database, never seen live.
		clientDB: []ts6.ClientDB{
			clientDBEntry("uid-flash", "Flash", "Flash", "de", now-10),
		},
	}

	activity := recordStatistics(snap, now)

	// Graph: exactly one point at the observation time with the live count.
	if len(onlineHistory) != 1 || onlineHistory[0].t != now || onlineHistory[0].n != 1 {
		t.Fatalf("history = %+v, want one point {t:%d n:1}", onlineHistory, now)
	}

	// Activity: the live client is online, the brief visitor is offline.
	live := activityLog["uid-live"]
	if live == nil || live.offlineAt != 0 || live.onlineAt != now {
		t.Errorf("uid-live entry wrong: %+v", live)
	}
	flash := activityLog["uid-flash"]
	if flash == nil || flash.offlineAt != now-10 || flash.onlineAt != now-10 {
		t.Errorf("uid-flash entry wrong: %+v", flash)
	}

	// The reconciled list must contain both rows, brief visitor first
	// (newest event: now-10 beats the live client's connection time now-120).
	if len(activity) != 2 || activity[0].Nickname != "Flash" || activity[1].Nickname != "Live" {
		t.Fatalf("activity = %+v, want [Flash Live]", activity)
	}
	if activity[0].Online || !activity[1].Online {
		t.Errorf("online flags wrong: %+v", activity)
	}

	// Persisted: the state file now holds both entries and the graph point.
	raw, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("state file was not written: %v", err)
	}
	var s persistedState
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("state file is not valid JSON: %v", err)
	}
	if len(s.Activity) != 2 {
		t.Errorf("persisted activity = %d entries, want 2: %+v", len(s.Activity), s.Activity)
	}
	if len(s.History) != 1 || s.History[0].T != now || s.History[0].N != 1 {
		t.Errorf("persisted history = %+v, want one point at %d with n=1", s.History, now)
	}
}

// TestRecordStatisticsRepeatWithinGap pins the sampler tick behaviour for a
// steady state: a second observation inside the 10 s dedupe gap with the
// same count must not add a duplicate graph point, must not duplicate the
// activity entry (one row per UID), and must still persist.
func TestRecordStatisticsRepeatWithinGap(t *testing.T) {
	saveStateGlobals(t)
	now := time.Now().Unix()

	snap := &statisticsSnapshot{
		clients: []ts6.Client{
			activityClientAt("Live", "uid-live", now-120),
		},
	}

	recordStatistics(snap, now)
	activity := recordStatistics(snap, now+5) // same count, 5 s later

	if len(onlineHistory) != 1 || onlineHistory[0].t != now {
		t.Fatalf("history = %+v, want the single original point at %d (no duplicate)", onlineHistory, now)
	}
	if len(activityLog) != 1 {
		t.Fatalf("activityLog = %+v, want exactly 1 entry (no duplicate row)", activityLog)
	}
	if e := activityLog["uid-live"]; e.onlineAt != now+5 {
		t.Errorf("onlineAt = %d, want %d (observation time must advance)", e.onlineAt, now+5)
	}
	if len(activity) != 1 || activity[0].Nickname != "Live" {
		t.Fatalf("activity = %+v, want [Live]", activity)
	}

	// And the repeat was persisted as well.
	if _, err := os.Stat(stateFile); err != nil {
		t.Errorf("state file missing after repeat observation: %v", err)
	}
}

// TestRecordStatisticsNoClientDB pins the degraded observation (client
// database fetch failed, or activity disabled): only the live clients are
// folded in, nothing is invented for absent UIDs.
func TestRecordStatisticsNoClientDB(t *testing.T) {
	saveStateGlobals(t)
	now := time.Now().Unix()

	snap := &statisticsSnapshot{
		clients: []ts6.Client{
			activityClientAt("Live", "uid-live", now-60),
		},
		// clientDB is nil: the fetch failed on this tick.
	}

	activity := recordStatistics(snap, now)

	if len(activityLog) != 1 {
		t.Fatalf("activityLog = %+v, want only the live client", activityLog)
	}
	if len(activity) != 1 || activity[0].Nickname != "Live" || !activity[0].Online {
		t.Fatalf("activity = %+v, want [Live online]", activity)
	}
}

// TestSamplerStartStopIdempotent pins the lifecycle contract: starting twice
// spawns exactly one sampler goroutine, stopping without a start (or twice)
// is a no-op, and the goroutine is guaranteed to have exited before the
// second stop returns.
func TestSamplerStartStopIdempotent(t *testing.T) {
	saveStateGlobals(t) // in case any tick manages to record: file goes to the temp dir
	cfg := testStatsCfg()

	base := runtime.NumGoroutine()

	StartStatisticsSampler(cfg)
	// Give the first recordOne a moment to attempt (and fail) its fetch.
	time.Sleep(300 * time.Millisecond)
	if got := runtime.NumGoroutine(); got != base+1 {
		t.Fatalf("after start: %d goroutines, want %d (exactly one sampler)", got, base+1)
	}

	StartStatisticsSampler(cfg) // idempotent: no second goroutine
	time.Sleep(300 * time.Millisecond)
	if got := runtime.NumGoroutine(); got != base+1 {
		t.Fatalf("after double start: %d goroutines, want %d (still one sampler)", got, base+1)
	}

	StopStatisticsSampler()
	if got := runtime.NumGoroutine(); got != base {
		t.Fatalf("after stop: %d goroutines, want %d (sampler exited)", got, base)
	}

	StopStatisticsSampler() // no-op, must not hang or panic

	// State must be untouched: the failed fetches never reached recordStatistics.
	if len(onlineHistory) != 0 || len(activityLog) != 0 {
		t.Errorf("state mutated despite failed fetches: history=%+v activity=%+v",
			onlineHistory, activityLog)
	}
}
