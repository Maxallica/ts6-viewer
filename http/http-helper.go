package http

import (
	"net"
	"net/http"
	"time"
	"ts6-viewer/internal/config"
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/mapper"
	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

func getIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func allowRequest(ip string) bool {
	mu.Lock()
	defer mu.Unlock()
	last, exists := lastRequestTime[ip]

	if exists && time.Since(last) < rateLimitWindow {
		return false
	}

	lastRequestTime[ip] = time.Now()
	return true
}

func getViewerData(cfg *config.Config, baseURL, apiKey, serverID string) (view.ViewerData, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(cacheTimestamp) < cacheTTL {
		return cacheData, nil
	}

	serverInfo, err := ts6.GetServerInfo(baseURL, apiKey, serverID)
	if err != nil {
		return view.ViewerData{}, err
	}

	apiClients, err := ts6.GetClientList(baseURL, apiKey, serverID)
	if err != nil {
		return view.ViewerData{}, err
	}

	apiChannels, err := ts6.GetChannelList(baseURL, apiKey, serverID)
	if err != nil {
		return view.ViewerData{}, err
	}

	// API → Domain
	domainChannels := make([]*domain.Channel, 0, len(apiChannels))
	for _, ch := range apiChannels {
		domainChannels = append(domainChannels, mapper.MapAPIChannel(ch))
	}

	fullClients := make([]*domain.FullClient, 0, len(apiClients))
	for _, c := range apiClients {
		// info, err := ts6.GetClientInfo(baseURL, apiKey, serverID, c.CLID)
		// if err != nil {
		// 	fmt.Println("clientinfo error:", err)
		// }

		domainInfo := &domain.ClientInfo{
			MicMuted:    false,
			OutputMuted: false,
			IsTalking:   false,
		}

		domainClient := mapper.MapAPIClient(c)
		// domainInfo := mapper.MapAPIClientInfo(info)

		fullClients = append(fullClients, &domain.FullClient{
			Client: *domainClient,
			Info:   domainInfo,
		})
	}

	channelTree := domain.BuildChannelTree(domainChannels, fullClients)

	// ServerInfo → Domain
	domainServer := mapper.MapAPIServer(serverInfo)

	// Domain → View
	viewData := view.ViewerData{
		Server:          mapper.MapServerToView(domainServer),
		ChannelTree:     mapper.MapChannelTreeToView(channelTree),
		Theme:           cfg.Theme,
		RefreshInterval: cfg.RefreshInterval,
	}

	// Set cache
	cacheData = viewData
	cacheTimestamp = time.Now()
	return viewData, nil
}
