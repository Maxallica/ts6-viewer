package http

import "ts6-viewer/internal/ts6"

type ViewerData struct {
	Server          *ServerView
	ChannelTree     []*ChannelView
	Theme           string
	RefreshInterval string
}

type ServerView struct {
	Name              string
	ClientsOnline     string
	MaxClients        string
	UptimePretty      string
	ChannelsOnline    string
	HostBannerURL     string
	ClientConnections string
}

type ClientView struct {
	Nickname string
}

type ChannelView struct {
	Name     string
	Type     ts6.ChannelType
	Align    ts6.Aligned
	Repeat   bool
	Clients  []*ClientView
	Children []*ChannelView
}
