package http

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"ts6-viewer/internal/config"
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/mapper"
	"ts6-viewer/internal/ts6/serverquery"
	"ts6-viewer/internal/ts6/webquery"
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

	var domainServer *domain.Server
	var domainChannels []*domain.Channel
	var fullClients []*domain.FullClient

	switch cfg.Teamspeak6.Mode {
	case "webquery":
		// Channels
		webChannels, err := webquery.GetChannelList(baseURL, apiKey, serverID)
		if err != nil {
			return view.ViewerData{}, err
		}

		domainChannels = make([]*domain.Channel, 0, len(webChannels))
		for _, ch := range webChannels {
			domainChannels = append(domainChannels, mapper.MapChannelByWebQuery(ch))
		}

		// Clients
		webClients, err := webquery.GetClientList(baseURL, apiKey, serverID)
		if err != nil {
			return view.ViewerData{}, err
		}

		fullClients = make([]*domain.FullClient, 0, len(webClients))

		for _, c := range webClients {
			var info *webquery.ClientInfo
			if cfg.Teamspeak6.EnableDetailedClientInfo == "true" {
				info, err = webquery.GetClientInfo(baseURL, apiKey, serverID, c.CLID)
				if err != nil {
					fmt.Println("clientinfo error:", err)
				}
			}

			domainClient := mapper.MapClientByWebQuery(c)
			domainInfo := mapper.MapClientInfoByWebQuery(info)

			fullClients = append(fullClients, &domain.FullClient{
				Client: *domainClient,
				Info:   domainInfo,
			})
		}

		// ServerInfo
		webServerInfo, err := webquery.GetServerInfo(baseURL, apiKey, serverID)
		if err != nil {
			return view.ViewerData{}, err
		}
		domainServer = mapper.MapServerByWebQuery(webServerInfo, &webClients)

	case "serverquery":
		ssh, err := serverquery.NewSSHClient(cfg)
		if err != nil {
			return view.ViewerData{}, err
		}

		err = ssh.Use(cfg.Teamspeak6.ServerID)
		if err != nil {
			return view.ViewerData{}, err
		}

		// Channels
		serverChannels, err := serverquery.GetChannelList(cfg, ssh)
		if err != nil {
			return view.ViewerData{}, err
		}

		domainChannels = make([]*domain.Channel, 0, len(serverChannels))
		for _, ch := range serverChannels {
			domainChannels = append(domainChannels, mapper.MapChannelByServerQuery(ch))
		}

		// Clients
		serverClients, err := serverquery.GetClientList(ssh)
		if err != nil {
			return view.ViewerData{}, err
		}

		fullClients = make([]*domain.FullClient, 0, len(serverClients))
		for _, c := range serverClients {
			var info *serverquery.ClientInfo
			if cfg.Teamspeak6.EnableDetailedClientInfo == "true" {
				info, err = serverquery.GetClientInfo(ssh, c.CLID)
				if err != nil {
					fmt.Println("clientinfo error:", err)
				}
			}

			domainClient := mapper.MapClientByServerQuery(c)
			domainInfo := mapper.MapClientInfoByServerQuery(info)

			fullClients = append(fullClients, &domain.FullClient{
				Client: *domainClient,
				Info:   domainInfo,
			})
		}

		// ServerInfo
		sshServerInfo, err := serverquery.GetServerInfo(cfg, ssh)
		if err != nil {
			return view.ViewerData{}, err
		}

		domainServer = mapper.MapServerByServerQuery(sshServerInfo, &serverClients)

		ssh.Close()
	default:
		return view.ViewerData{}, fmt.Errorf("unsupported mode: %s", cfg.Teamspeak6.Mode)
	}

	channelTree := domain.BuildChannelTree(domainChannels, fullClients)

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
