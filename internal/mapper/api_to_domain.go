package mapper

import (
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/ts6"
)

func MapAPIServer(server *ts6.ServerInfo) *domain.Server {
	uptimePretty := domain.MakeUptimePretty(server.Uptime)

	return &domain.Server{
		Name:              server.Name,
		ClientsOnline:     server.ClientsOnline,
		MaxClients:        server.MaxClients,
		UptimePretty:      uptimePretty,
		ChannelsOnline:    server.ChannelsOnline,
		HostBannerURL:     server.HostBannerURL,
		ClientConnections: server.ClientConnections,
	}
}

func MapAPIChannel(api ts6.Channel) *domain.Channel {
	chType, align, repeat, cleanName := domain.ParseChannelName(api.Name)

	return &domain.Channel{
		ID:       api.CID,
		ParentID: api.PID,
		Name:     cleanName,
		Topic:    api.Topic,
		Type:     chType,
		Align:    align,
		Repeat:   repeat,
	}
}

func MapAPIClient(c ts6.Client) *domain.Client {
	return &domain.Client{
		ID:        c.CLID,
		Nickname:  c.Nickname,
		ChannelID: c.CID,
	}
}

func MapAPIClientInfo(info *ts6.ClientInfo) *domain.ClientInfo {
	if info == nil {
		return &domain.ClientInfo{
			MicMuted:    false,
			OutputMuted: false,
			IsTalking:   false,
		}
	}

	return &domain.ClientInfo{
		MicMuted:    info.InputMuted == "1" || info.InputHardware == "1",
		OutputMuted: info.OutputMuted == "1",
		IsTalking:   info.IsTalker == "1",
	}
}
