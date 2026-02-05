package mapper

import (
	"ts6-viewer/internal/domain"
	"ts6-viewer/internal/view"
)

func MapServerToView(s *domain.Server) *view.ServerView {
	return &view.ServerView{
		Name:              s.Name,
		ClientsOnline:     s.ClientsOnline,
		MaxClients:        s.MaxClients,
		UptimePretty:      s.UptimePretty,
		ChannelsOnline:    s.ChannelsOnline,
		HostBannerURL:     s.HostBannerURL,
		ClientConnections: s.ClientConnections,
	}
}

func MapClientToView(c *domain.FullClient) *view.ClientView {
	v := &view.ClientView{
		Nickname: c.Nickname,
	}

	if c.Info != nil {
		v.MicMuted = c.Info.MicMuted
		v.OutputMuted = c.Info.OutputMuted
		v.IsTalking = c.Info.IsTalking
	}

	return v
}

func MapChannelToView(ch *domain.Channel) *view.ChannelView {
	out := &view.ChannelView{
		Name:   ch.Name,
		Type:   ch.Type,
		Align:  ch.Align,
		Repeat: ch.Repeat,
	}

	for _, c := range ch.Clients {
		out.Clients = append(out.Clients, MapClientToView(c))
	}

	for _, child := range ch.Children {
		out.Children = append(out.Children, MapChannelToView(child))
	}

	return out
}

func MapChannelTreeToView(tree []*domain.Channel) []*view.ChannelView {
	out := make([]*view.ChannelView, 0, len(tree))
	for _, ch := range tree {
		out = append(out, MapChannelToView(ch))
	}
	return out
}
