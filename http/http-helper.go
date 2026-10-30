package http

import (
	"log"
	"net/http"
	"time"
	"ts6-viewer/internal/config"
	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

// getIP extracts the IP address from the request.
func getIP(r *http.Request) string {
	ip := r.RemoteAddr
	if ipForwarded := r.Header.Get("X-Forwarded-For"); ipForwarded != "" {
		ip = ipForwarded
	}
	return ip
}

// allowRequest checks rate limiting per IP.
func allowRequest(ip string) bool {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	last, ok := lastRequestTime[ip]
	if ok && now.Sub(last) < rateLimitWindow {
		return false
	}

	lastRequestTime[ip] = now
	return true
}

// getViewerData fetches or returns cached viewer data.
func getViewerData(cfg *config.Config, force bool) (view.VMTS6Viewer, error) {
	mu.Lock()
	defer mu.Unlock()

	if !force && time.Since(cacheTimestamp) < cacheTTL {
		log.Println("[HTTP] Returning cached viewer data")
		return cacheData, nil
	}

	log.Println("[HTTP] Fetching new viewer data from TS6 server")

	var (
		channels   []ts6.Channel
		snap       statisticsSnapshot
		info       *ts6.ServerInfo
		fetchErr   error
	)
	// Fetch everything for the display in one pass over the persistent SSH
	// connection. The statistics fields (clients, clientDB, groupBadges) are
	// captured in snap so the same observation can be folded into the rolling
	// statistics state, exactly as the background sampler does.
	fetchErr = withSSH(cfg, func(ssh *ts6.SSHClient) error {
		var err error

		channels, err = ts6.GetChannelList(cfg, ssh)
		if err != nil {
			return err
		}

		snap.clients, err = ts6.GetClientList(cfg, ssh)
		if err != nil {
			return err
		}

		// The server's client database also covers clients that were only
		// briefly online between observations (they never show up in the
		// live list). Fetched only while the activity log is on; a failure
		// here degrades to the live list only instead of breaking the viewer.
		if statsActivityEnabled {
			if clientDB, err := ts6.GetClientDB(cfg, ssh); err != nil {
				log.Printf("[HTTP] Failed to get client database (brief connections not tracked): %v\n", err)
			} else {
				snap.clientDB = clientDB
			}
		}

		groups, err := ts6.GetServerGroups(cfg, ssh)
		if err != nil {
			return err
		}
		snap.groupBadges = view.GroupBadgeMap(groups)

		info, err = ts6.GetServerInfo(cfg, ssh)
		if err != nil {
			return err
		}
		return nil
	})
	if fetchErr != nil {
		log.Printf("[HTTP] Failed to get viewer data: %v\n", fetchErr)
		return view.VMTS6Viewer{}, fetchErr
	}

	now := time.Now().Unix()

	// Fold this observation into the rolling statistics state and persist,
	// on the same code path as the background sampler. A brief session is
	// therefore recorded identically no matter how the observation is
	// triggered (viewer request or sampler tick), and the reconciled activity
	// list is reused for the render below.
	activity := recordStatistics(&snap, now)

	maxWidth := cfg.MaxWidth
	if maxWidth == "" {
		maxWidth = "800px"
	}

	vmTS6Viewer := view.VMTS6Viewer{
		VMServer:        view.BuildVMServer(cfg, info, snap.clients),
		VMChannels:      view.BuildVMChannels(channels, snap.clients, snap.groupBadges),
		VMActivity:      activity,
		VMHistory:       onlineHistoryFor(now),
		VMStats:         statsViewModel(),
		RefreshInterval: cfg.RefreshInterval,
		MaxWidth:        maxWidth,
	}

	cacheData = vmTS6Viewer
	cacheTimestamp = time.Now()
	// The statistics state was already persisted by recordStatistics (mu is
	// held for the whole fetch, so the snapshot is consistent).
	log.Println("[HTTP] Viewer data updated and cached")

	return vmTS6Viewer, nil
}
