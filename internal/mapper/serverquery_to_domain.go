package mapper

import (
	"strconv"
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/ts6/serverquery"
)

func MapServerByServerQuery(info *serverquery.ServerInfo, clients *[]serverquery.Client) *domain.Server {
	if info == nil {
		return nil
	}

	uptimePretty := domain.MakeUptimePretty(info.Uptime)

	return &domain.Server{
		Name:              info.Name,
		ClientsOnline:     strconv.Itoa(len(*clients)),
		MaxClients:        info.MaxClients,
		UptimePretty:      uptimePretty,
		ChannelsOnline:    info.ChannelsOnline,
		HostBannerURL:     info.HostBannerURL,
		ClientConnections: info.ClientConnections,
	}
}

func MapChannelByServerQuery(api serverquery.Channel) *domain.Channel {
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

func MapClientByServerQuery(c serverquery.Client) *domain.Client {
	return &domain.Client{
		ID:        c.CLID,
		Nickname:  c.Nickname,
		ChannelID: c.CID,
	}
}

func MapClientInfoByServerQuery(info *serverquery.ClientInfo) *domain.ClientInfo {
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
