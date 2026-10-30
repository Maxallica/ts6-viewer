package http

import (
	"testing"
	"time"
)

// TestHistorySampleDedupSameCount checks that a repeat of the current count
// inside the sample gap is dropped (two tabs polling within seconds must not
// create two identical points), and that it does NOT overwrite the existing
// point.
func TestHistorySampleDedupSameCount(t *testing.T) {
	saved := onlineHistory
	onlineHistory = nil
	defer func() { onlineHistory = saved }()

	now := time.Now().Unix()

	sampleHistory(3, now-100)
	sampleHistory(3, now-95) // 5 s gap, same count → drop the repeat

	if len(onlineHistory) != 1 {
		t.Fatalf("expected 1 sample after same-count dedup, got %d: %+v", len(onlineHistory), onlineHistory)
	}
	if onlineHistory[0].t != now-100 || onlineHistory[0].n != 3 {
		t.Errorf("sample = %+v, want t=%d n=3 (unchanged)", onlineHistory[0], now-100)
	}
}

// TestHistorySampleChangeWithinGap checks the regression where a changed
// count inside the sample gap overwrote the last point's value at its old
// timestamp: a changed count must append a fresh point at the current time
// and leave the previous value intact, so a brief spike is captured at the
// right moment and never destroys history.
func TestHistorySampleChangeWithinGap(t *testing.T) {
	saved := onlineHistory
	onlineHistory = nil
	defer func() { onlineHistory = saved }()

	now := time.Now().Unix()

	sampleHistory(3, now-100)
	sampleHistory(4, now-95) // 5 s gap, count changed → append, do not overwrite

	if len(onlineHistory) != 2 {
		t.Fatalf("expected 2 samples after a count change, got %d: %+v", len(onlineHistory), onlineHistory)
	}
	if onlineHistory[0].t != now-100 || onlineHistory[0].n != 3 {
		t.Errorf("first sample = %+v, want t=%d n=3 (previous value preserved)", onlineHistory[0], now-100)
	}
	if onlineHistory[1].t != now-95 || onlineHistory[1].n != 4 {
		t.Errorf("second sample = %+v, want t=%d n=4 (new point at the change time)", onlineHistory[1], now-95)
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
