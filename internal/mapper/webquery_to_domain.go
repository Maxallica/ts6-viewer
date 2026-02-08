package mapper

import (
	"strconv"
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/ts6/webquery"
)

func MapServerByWebQuery(server *webquery.ServerInfo, clients *[]webquery.Client) *domain.Server {
	uptimePretty := domain.MakeUptimePretty(server.Uptime)

	return &domain.Server{
		Name:              server.Name,
		ClientsOnline:     strconv.Itoa(len(*clients)),
		MaxClients:        server.MaxClients,
		UptimePretty:      uptimePretty,
		ChannelsOnline:    server.ChannelsOnline,
		HostBannerURL:     server.HostBannerURL,
		ClientConnections: server.ClientConnections,
	}
}

func MapChannelByWebQuery(api webquery.Channel) *domain.Channel {
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

func MapClientByWebQuery(c webquery.Client) *domain.Client {
	return &domain.Client{
		ID:        c.CLID,
		Nickname:  c.Nickname,
		ChannelID: c.CID,
	}
}

func MapClientInfoByWebQuery(info *webquery.ClientInfo) *domain.ClientInfo {
	if info == nil {
		return &domain.ClientInfo{
			MicMuted:    false,
			OutputMuted: false,
			IsTalking:   false,
		}
	}

	return &domain.ClientInfo{
		MicMuted:    info.InputMuted == "1",
		OutputMuted: info.OutputMuted == "1",
		IsTalking:   info.IsTalker == "1",
	}
}
