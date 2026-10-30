package http

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// saveStateGlobals snapshots the package-level state that the persistence
// tests replace, points the state file at a per-test temp file, and
// restores everything afterwards.
func saveStateGlobals(t *testing.T) {
	t.Helper()
	savedLog := activityLog
	savedHistory := onlineHistory
	savedFile := stateFile
	t.Cleanup(func() {
		activityLog = savedLog
		onlineHistory = savedHistory
		stateFile = savedFile
	})
	activityLog = make(map[string]*activityEntry)
	onlineHistory = nil
	stateFile = filepath.Join(t.TempDir(), "state.json")
}

// seedState fills the in-memory statistics state with one online and one
// offline entry plus two history samples.
func seedState(now int64) {
	activityLog["uid1"] = &activityEntry{
		nickname:        "alice",
		badge:           "admin",
		country:         "DE",
		connectedAt:     "14:32",
		connectedAtUnix: now - 120,
		onlineAt:        now,
	}
	activityLog["uid2"] = &activityEntry{
		nickname:  "bob",
		country:   "AT",
		onlineAt:  now - 3600,
		offlineAt: now - 1800,
	}
	onlineHistory = []historySample{{t: now - 3600, n: 2}, {t: now, n: 1}}
}

// TestStateRoundTrip pins the persistence contract: save, "restart"
// (in-memory state wiped), load, and everything must come back intact.
func TestStateRoundTrip(t *testing.T) {
	saveStateGlobals(t)
	now := time.Now().Unix()
	seedState(now)

	saveState()

	// Simulate a process restart: in-memory state is gone.
	activityLog = make(map[string]*activityEntry)
	onlineHistory = nil

	loadState()

	if len(activityLog) != 2 {
		t.Fatalf("after load, want 2 activity entries, got %d", len(activityLog))
	}
	e := activityLog["uid1"]
	if e == nil || e.nickname != "alice" || e.badge != "admin" ||
		e.connectedAt != "14:32" || e.connectedAtUnix != now-120 || e.onlineAt != now || e.offlineAt != 0 {
		t.Errorf("uid1 not restored correctly: %+v", e)
	}
	e = activityLog["uid2"]
	if e == nil || e.nickname != "bob" || e.onlineAt != now-3600 || e.offlineAt != now-1800 {
		t.Errorf("uid2 not restored correctly: %+v", e)
	}
	if len(onlineHistory) != 2 || onlineHistory[0].t != now-3600 || onlineHistory[0].n != 2 ||
		onlineHistory[1].t != now || onlineHistory[1].n != 1 {
		t.Errorf("history not restored: %+v", onlineHistory)
	}
}

// TestLoadStatePrunesStale pins that entries and samples already outside
// their windows are dropped while loading. Both windows are set explicitly
// (24h) so the test does not depend on the defaults.
func TestLoadStatePrunesStale(t *testing.T) {
	saveStateGlobals(t)
	savedHistory := historyWindow
	savedActivity := activityWindow
	historyWindow = 24 * time.Hour
	activityWindow = 24 * time.Hour
	t.Cleanup(func() {
		historyWindow = savedHistory
		activityWindow = savedActivity
	})
	now := time.Now().Unix()
	stale := now - int64(25*time.Hour/time.Second)

	raw := `{
		"activity": {
			"fresh": {"nickname": "new", "badge": "", "country": "", "connectedAt": "", "connectedAtUnix": 0, "onlineAt": ` + strconv.FormatInt(now-60, 10) + `, "offlineAt": 0},
			"stale": {"nickname": "old", "badge": "", "country": "", "connectedAt": "", "connectedAtUnix": 0, "onlineAt": ` + strconv.FormatInt(stale, 10) + `, "offlineAt": ` + strconv.FormatInt(stale, 10) + `}
		},
		"history": [{"t": ` + strconv.FormatInt(stale, 10) + `, "n": 9}, {"t": ` + strconv.FormatInt(now-120, 10) + `, "n": 3}]
	}`
	if err := os.WriteFile(stateFile, []byte(raw), 0o600); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	loadState()

	if len(activityLog) != 1 {
		t.Fatalf("after load, want 1 activity entry (stale pruned), got %d", len(activityLog))
	}
	if activityLog["fresh"] == nil || activityLog["fresh"].nickname != "new" {
		t.Errorf("fresh entry not restored: %+v", activityLog)
	}
	if len(onlineHistory) != 1 || onlineHistory[0].n != 3 {
		t.Errorf("history not pruned correctly: %+v", onlineHistory)
	}
}

// TestLoadStateMissingFile pins that a missing state file is a normal
// first start: no error, state simply stays empty.
func TestLoadStateMissingFile(t *testing.T) {
	saveStateGlobals(t)

	loadState()

	if len(activityLog) != 0 || len(onlineHistory) != 0 {
		t.Errorf("expected empty state after first start, got %d entries / %d samples",
			len(activityLog), len(onlineHistory))
	}
}

// TestLoadStateCorruptFile pins that an unreadable state file is ignored
// instead of crashing startup.
func TestLoadStateCorruptFile(t *testing.T) {
	saveStateGlobals(t)

	if err := os.WriteFile(stateFile, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write corrupt state file: %v", err)
	}

	loadState()

	if len(activityLog) != 0 || len(onlineHistory) != 0 {
		t.Errorf("expected empty state after corrupt file, got %d entries / %d samples",
			len(activityLog), len(onlineHistory))
	}
}
