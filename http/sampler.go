package http

import (
	"context"
	"log"
	"sync"
	"time"

	"ts6-viewer/internal/config"
	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

// samplerInterval is how often the background statistics sampler polls the
// TeamSpeak server. It is a var (not a const) so tests can shrink it.
// 30s keeps the graph and the recent-activity log current even when nobody
// is viewing the page, while a user online for two minutes or less is still
// caught (either live or via the client database safety net).
var samplerInterval = 30 * time.Second

// statisticsSampler holds the handle for the background sampler goroutine.
// samplerCancel/samplerDone are set by StartStatisticsSampler and cleared by
// StopStatisticsSampler, which is what makes both idempotent. Guarded by
// samplerMu.
var (
	samplerCancel context.CancelFunc
	samplerDone   chan struct{}
	samplerMu     sync.Mutex
)

// withSSH runs fn on the persistent SSH ServerQuery connection, creating it
// on first use. Both getViewerData and the background sampler acquire the
// same singleton this way, so their state updates never interleave (each
// holds mu for its whole fetch).
func withSSH(cfg *config.Config, fn func(ssh *ts6.SSHClient) error) error {
	sshClient, err := ts6.GetPersistentClient(cfg, cfg.Teamspeak6.ServerID)
	if err != nil {
		return err
	}
	return fn(sshClient)
}

// fetchStatisticsSnapshot pulls the live data the statistics state needs for
// one observation: the online client list, the client database (the
// brief-visit safety net) and the server-group badge map. It reuses the
// persistent SSH connection. The caller must hold mu (it shares the SSH
// client with getViewerData, which also holds mu, so this keeps the two
// paths from interleaving their state updates).
func fetchStatisticsSnapshot(cfg *config.Config) (*statisticsSnapshot, error) {
	var snap statisticsSnapshot
	err := withSSH(cfg, func(ssh *ts6.SSHClient) error {
		clients, err := ts6.GetClientList(cfg, ssh)
		if err != nil {
			return err
		}
		snap.clients = clients

		// The server's client database also covers clients that were only
		// briefly online between observations (they never show up in the
		// live list). Fetched only while the activity log is on; a failure
		// here degrades to the live list only instead of breaking the
		// observation.
		if statsActivityEnabled {
			clientDB, err := ts6.GetClientDB(cfg, ssh)
			if err != nil {
				log.Printf("[Sampler] Failed to get client database (brief connections not tracked this tick): %v\n", err)
			} else {
				snap.clientDB = clientDB
			}
		}

		groups, err := ts6.GetServerGroups(cfg, ssh)
		if err != nil {
			return err
		}
		snap.groupBadges = view.GroupBadgeMap(groups)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

// recordStatistics folds one observation into the rolling statistics state
// (online-count history + recent activity) and persists it. It returns the
// reconciled activity list so the caller can render it without a second
// pass. The caller must hold mu: it mutates onlineHistory and activityLog
// and rewrites the state file, exactly like the old getViewerData path did.
func recordStatistics(snap *statisticsSnapshot, now int64) []*view.VMActivity {
	// Sample the online count first so the graph reflects the live list,
	// then fold the client database in (brief visits the live list never
	// saw), then reconcile the activity log against the live list. The
	// order matters: the client-database entries must be present before the
	// log is reconciled, or a brief visit would be pruned as "not online".
	sampleHistory(len(snap.clients), now)
	mergeClientDB(snap.clientDB, now)
	activity := updateActivity(snap.clients, snap.groupBadges, now)
	saveState()
	return activity
}

// StartStatisticsSampler launches the background statistics sampler and
// returns nothing. It is idempotent: starting twice is a no-op (the second
// call returns without spawning another goroutine). The sampler polls the
// TeamSpeak server on a fixed interval and records + persists the statistics
// state, independently of whether any viewer is currently requesting data.
// This is what makes a brief (≤2 min) session land in both the graph and the
// recent-activity log even when no one is watching the page.
func StartStatisticsSampler(cfg *config.Config) {
	samplerMu.Lock()
	defer samplerMu.Unlock()

	if samplerCancel != nil {
		// Already running.
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	samplerCancel = cancel
	samplerDone = done

	log.Printf("[Sampler] Starting statistics sampler (interval %v)\n", samplerInterval)
	go runStatisticsSampler(cfg, ctx, done)
}

// StopStatisticsSampler cancels the background statistics sampler (if
// running) and waits for the goroutine to exit. Idempotent: calling it
// without a start, or twice, is a no-op. Safe to call from main on shutdown.
func StopStatisticsSampler() {
	samplerMu.Lock()
	cancel := samplerCancel
	done := samplerDone
	if cancel != nil {
		samplerCancel = nil
		samplerDone = nil
	}
	samplerMu.Unlock()

	if cancel == nil {
		return // never started
	}

	cancel()
	<-done
	log.Println("[Sampler] Statistics sampler stopped")
}

// runStatisticsSampler is the sampler loop body. It records an observation
// immediately on start (so a fresh server samples the current state right
// away) and then on every samplerInterval tick until ctx is cancelled.
// Each tick that fails to fetch (e.g. the SSH connection is down) is logged
// and skipped; the loop keeps going so a transient outage does not stop the
// sampler.
func runStatisticsSampler(cfg *config.Config, ctx context.Context, done chan struct{}) {
	defer close(done)

	ticker := time.NewTicker(samplerInterval)
	defer ticker.Stop()

	recordOne(cfg, "start")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			recordOne(cfg, "tick")
		}
	}
}

// recordOne performs a single statistics observation: fetch the live data,
// fold it into the state and persist. It holds mu for the whole fetch so the
// snapshot and the state update are consistent (the same guarantee
// getViewerData provides). A fetch failure is logged, not fatal.
func recordOne(cfg *config.Config, label string) {
	mu.Lock()
	defer mu.Unlock()

	snap, err := fetchStatisticsSnapshot(cfg)
	if err != nil {
		log.Printf("[Sampler] %s: fetch failed (state unchanged): %v\n", label, err)
		return
	}

	recordStatistics(snap, time.Now().Unix())
	log.Printf("[Sampler] %s: recorded %d online clients\n", label, len(snap.clients))
}

// statisticsSnapshot is one observation's worth of live data for the
// statistics state.
type statisticsSnapshot struct {
	clients     []ts6.Client
	clientDB    []ts6.ClientDB
	groupBadges map[string]string
}
