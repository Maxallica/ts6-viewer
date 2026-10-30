package http

import (
	"time"
	"ts6-viewer/internal/view"
)

// historyWindow is how long the online-count graph keeps its samples. It is
// a var because the config's graph_max_range ("24h" | "7d" | "14d" | "1m")
// overrides it at startup; the default is 1m (31 days).
var historyWindow = 31 * 24 * time.Hour

// historyFineWindow: samples newer than this are stored at full resolution.
const historyFineWindow = 24 * time.Hour

// historyBucketSize: how finely samples older than historyFineWindow are
// aggregated (average online count per bucket). 15 minutes keeps even a
// month of history small (about 2900 buckets) while the 24h view stays
// sample-exact.
const historyBucketSize = 15 * time.Minute

// historySampleGap deduplicates samples from the same refresh burst (two
// tabs polling within seconds must not create two identical points). It only
// suppresses a *repeat* of the current count; a changed count always appends
// a fresh point, so a brief spike is never lost and the previous value is
// never overwritten at an old timestamp.
const historySampleGap = 10 * time.Second

// historySample is one observation of the online client count. For samples
// older than historyFineWindow, t is the bucket start and n the average
// online count inside that bucket.
type historySample struct {
	t int64 // unix seconds
	n int   // online clients
}

// onlineHistory is the rolling online-count series, oldest first.
// Guarded by mu: both the viewer path (getViewerData) and the background
// sampler (recordStatistics) hold it for the whole observation, so
// sampleHistory/onlineHistoryFor never run concurrently with each other.
var onlineHistory []historySample

// sampleHistory records the current online count once per data fetch and
// compacts + prunes the series against the configured window. A changed
// count always appends a new point at the current time, so a brief spike is
// captured at the right moment and the previous value is never destroyed.
// A repeat of the current count inside the sample gap is dropped (two tabs
// polling within seconds must not create two identical points); a repeat
// outside the gap is kept, so the line still has a point at that time.
func sampleHistory(onlineCount int, now int64) {
	if len(onlineHistory) > 0 {
		last := onlineHistory[len(onlineHistory)-1]
		if now-last.t < int64(historySampleGap/time.Second) && last.n == onlineCount {
			// Same observation burst with the same count: drop the repeat.
			return
		}
	}

	onlineHistory = append(onlineHistory, historySample{t: now, n: onlineCount})
	compactHistory(now)
	pruneHistory(now)
}

// compactHistory aggregates samples older than historyFineWindow into
// historyBucketSize buckets (average online count per bucket) so long
// retention (7d/14d/1m) does not bloat state.json or the poll payload.
func compactHistory(now int64) {
	bucketStart := now - int64(historyFineWindow/time.Second)
	bucketSecs := int64(historyBucketSize / time.Second)

	i := 0
	for i < len(onlineHistory) && onlineHistory[i].t < bucketStart {
		i++
	}
	if i == 0 || i == len(onlineHistory) {
		return // nothing to compact (or nothing new to keep after it)
	}

	var compacted []historySample
	for j := 0; j < i; {
		bucket := onlineHistory[j].t / bucketSecs * bucketSecs
		sum, cnt := 0, 0
		for j < i && onlineHistory[j].t/bucketSecs*bucketSecs == bucket {
			sum += onlineHistory[j].n
			cnt++
			j++
		}
		compacted = append(compacted, historySample{t: bucket, n: sum / cnt})
	}

	onlineHistory = append(compacted, onlineHistory[i:]...)
}

// onlineHistoryFor returns the rolling series as view model points, oldest
// first, for the online-count graph.
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

// pruneHistory drops samples older than the configured window.
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
