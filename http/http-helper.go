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
	sshClient, err := ts6.GetPersistentClient(cfg, cfg.Teamspeak6.ServerID)
	if err != nil {
		log.Printf("[HTTP] Failed to get SSH client: %v\n", err)
		return view.VMTS6Viewer{}, err
	}

	channels, err := ts6.GetChannelList(cfg, sshClient)
	if err != nil {
		log.Printf("[HTTP] Failed to get channels: %v\n", err)
		return view.VMTS6Viewer{}, err
	}

	clients, err := ts6.GetClientList(cfg, sshClient)
	if err != nil {
		log.Printf("[HTTP] Failed to get clients: %v\n", err)
		return view.VMTS6Viewer{}, err
	}

	// The server's client database also covers clients that were only
	// briefly online between polls (they never show up in the live list).
	// Fetched only while the activity log is on; a failure here degrades
	// to the old behaviour instead of breaking the viewer.
	var clientDB []ts6.ClientDB
	if statsActivityEnabled {
		clientDB, err = ts6.GetClientDB(cfg, sshClient)
		if err != nil {
			log.Printf("[HTTP] Failed to get client database (brief connections not tracked): %v\n", err)
			clientDB = nil
		}
	}

	groups, err := ts6.GetServerGroups(cfg, sshClient)
	if err != nil {
		log.Printf("[HTTP] Failed to get servergroups: %v\n", err)
		return view.VMTS6Viewer{}, err
	}

	info, err := ts6.GetServerInfo(cfg, sshClient)
	if err != nil {
		log.Printf("[HTTP] Failed to get server info: %v\n", err)
		return view.VMTS6Viewer{}, err
	}

	maxWidth := cfg.MaxWidth
	if maxWidth == "" {
		maxWidth = "800px"
	}

	groupBadges := view.GroupBadgeMap(groups)

	now := time.Now().Unix()
	sampleHistory(len(clients), now)

	// Fold the client database in before the log is reconciled and
	// rendered, so brief connections that the live list never saw are
	// recorded as offline entries with the server's last-seen time.
	mergeClientDB(clientDB, now)

	vmTS6Viewer := view.VMTS6Viewer{
		VMServer:        view.BuildVMServer(cfg, info, clients),
		VMChannels:      view.BuildVMChannels(channels, clients, groupBadges),
		VMActivity:      updateActivity(clients, groupBadges, now),
		VMHistory:       onlineHistoryFor(now),
		VMStats:         statsViewModel(),
		RefreshInterval: cfg.RefreshInterval,
		MaxWidth:        maxWidth,
	}

	cacheData = vmTS6Viewer
	cacheTimestamp = time.Now()
	// Persist the statistics state so it survives a server restart.
	// mu is held for the whole fetch, so the snapshot is consistent.
	saveState()
	log.Println("[HTTP] Viewer data updated and cached")

	return vmTS6Viewer, nil
}
