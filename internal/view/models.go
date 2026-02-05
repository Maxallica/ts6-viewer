package view

import (
	"ts6-viewer/internal/domain"
)

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
	Nickname    string
	Platform    string
	Version     string
	MicMuted    bool
	OutputMuted bool
	IsTalking   bool
}

type ChannelView struct {
	Name     string
	Type     domain.ChannelType
	Align    domain.Aligned
	Repeat   bool
	Clients  []*ClientView
	Children []*ChannelView
}
