package http

import (
	"net"
	"net/http"
	"time"
	"ts6-viewer/internal/ts6"
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

func getViewerData(baseURL, apiKey string, serverID string) (ViewerData, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(cacheTimestamp) < cacheTTL {
		return cacheData, nil
	}

	serverInfo, err := ts6.GetServerInfo(baseURL, apiKey, serverID)
	if err != nil {
		return ViewerData{}, err
	}

	clients, err := ts6.GetClientList(baseURL, apiKey, serverID)
	if err != nil {
		return ViewerData{}, err
	}

	channelTree, err := ts6.GetChannelTree(baseURL, apiKey, serverID, clients)
	if err != nil {
		return ViewerData{}, err
	}

	view := ViewerData{
		Server:      mapServer(serverInfo),
		ChannelTree: mapChannelTree(channelTree),
	}

	cacheData = view
	cacheTimestamp = time.Now()

	return view, nil
}

func mapServer(s *ts6.ServerInfo) *ServerView {
	return &ServerView{
		Name:              s.Name,
		ClientsOnline:     s.ClientsOnline,
		MaxClients:        s.MaxClients,
		UptimePretty:      s.UptimePretty,
		ChannelsOnline:    s.ChannelsOnline,
		HostBannerURL:     s.HostBannerURL,
		ClientConnections: s.ClientConnections,
	}
}

func mapClient(c *ts6.Client) *ClientView {
	return &ClientView{
		Nickname: c.Nickname,
	}
}

func mapChannel(ch *ts6.Channel) *ChannelView {
	out := &ChannelView{
		Name:   ch.Name,
		Type:   ch.Type,
		Align:  ch.Align,
		Repeat: ch.Repeat,
	}

	for _, c := range ch.Clients {
		out.Clients = append(out.Clients, mapClient(c))
	}

	for _, child := range ch.Children {
		out.Children = append(out.Children, mapChannel(child))
	}

	return out
}

func mapChannelTree(tree []*ts6.Channel) []*ChannelView {
	out := make([]*ChannelView, 0, len(tree))
	for _, ch := range tree {
		out = append(out, mapChannel(ch))
	}
	return out
}
