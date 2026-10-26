package http

import (
	"testing"
	"time"
)

// TestHistorySampleDedup checks that two observations within the sample gap
// refresh the last point instead of appending a duplicate.
func TestHistorySampleDedup(t *testing.T) {
	saved := onlineHistory
	onlineHistory = nil
	defer func() { onlineHistory = saved }()

	now := time.Now().Unix()

	sampleHistory(3, now-100)
	sampleHistory(4, now-95) // 5 s gap → refresh the last point

	if len(onlineHistory) != 1 {
		t.Fatalf("expected 1 sample after dedup, got %d: %+v", len(onlineHistory), onlineHistory)
	}
	if onlineHistory[0].t != now-100 || onlineHistory[0].n != 4 {
		t.Errorf("last sample = %+v, want t=%d n=4", onlineHistory[0], now-100)
	}

	sampleHistory(2, now) // 100 s gap → new point
	if len(onlineHistory) != 2 {
		t.Fatalf("expected 2 samples after new observation, got %d: %+v", len(onlineHistory), onlineHistory)
	}
}

// TestHistoryPurge checks that samples older than the configured window are
// dropped and that the returned view model series is ordered oldest first.
// The window is set explicitly so the test does not depend on the default.
func TestHistoryPurge(t *testing.T) {
	savedHist := onlineHistory
	savedWindow := historyWindow
	onlineHistory = nil
	historyWindow = 24 * time.Hour
	defer func() {
		onlineHistory = savedHist
		historyWindow = savedWindow
	}()

	now := time.Now().Unix()

	sampleHistory(1, now-25*3600) // 25 h ago → outside the window
	sampleHistory(2, now-600)
	sampleHistory(5, now-10)
	sampleHistory(4, now)

	points := onlineHistoryFor(now)
	if len(points) != 3 {
		t.Fatalf("expected 3 points after purge, got %d: %+v", len(points), points)
	}
	if points[0].T != now-600 || points[1].T != now-10 || points[2].T != now {
		t.Errorf("series order broken: %+v", points)
	}
	if points[2].N != 4 {
		t.Errorf("last point n = %d, want 4", points[2].N)
	}
}

// TestHistoryEmptySeries pins the empty state: no samples in, no points out.
func TestHistoryEmptySeries(t *testing.T) {
	saved := onlineHistory
	onlineHistory = nil
	defer func() { onlineHistory = saved }()

	points := onlineHistoryFor(time.Now().Unix())
	if points == nil || len(points) != 0 {
		t.Errorf("expected an empty (non-nil) slice, got %+v", points)
	}
}
