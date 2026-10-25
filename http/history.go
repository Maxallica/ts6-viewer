package http

import (
	"time"
	"ts6-viewer/internal/view"
)

// historyWindow is how long the online-count graph keeps its samples.
const historyWindow = 24 * time.Hour

// historySampleGap deduplicates samples from the same refresh burst (two
// tabs polling within seconds must not create two identical points).
const historySampleGap = 10 * time.Second

// historySample is one observation of the online client count.
type historySample struct {
	t int64 // unix seconds
	n int   // online clients
}

// onlineHistory is the rolling 24h online-count series, oldest first.
// Guarded by mu: getViewerData holds it for the whole fetch, and
// sampleHistory/onlineHistoryFor are only ever called from that path.
var onlineHistory []historySample

// sampleHistory records the current online count once per data fetch and
// prunes samples that fell out of the 24h window.
func sampleHistory(onlineCount int, now int64) {
	if len(onlineHistory) > 0 {
		last := onlineHistory[len(onlineHistory)-1]
		if now-last.t < int64(historySampleGap/time.Second) {
			// Same observation burst: refresh the last point instead of
			// appending a duplicate.
			onlineHistory[len(onlineHistory)-1].n = onlineCount
			return
		}
	}

	onlineHistory = append(onlineHistory, historySample{t: now, n: onlineCount})
	pruneHistory(now)
}

// onlineHistoryFor returns the rolling series as view model points, oldest
// first, for the 24h online-count graph.
func onlineHistoryFor(now int64) []*view.VMHistoryPoint {
	pruneHistory(now)

	points := make([]*view.VMHistoryPoint, 0, len(onlineHistory))
	for _, s := range onlineHistory {
		points = append(points, &view.VMHistoryPoint{
			T: s.t,
			N: s.n,
		})
	}
	return points
}

// pruneHistory drops samples older than the 24h window.
func pruneHistory(now int64) {
	cutoff := now - int64(historyWindow/time.Second)
	i := 0
	for ; i < len(onlineHistory); i++ {
		if onlineHistory[i].t >= cutoff {
			break
		}
	}
	onlineHistory = onlineHistory[i:]
}
