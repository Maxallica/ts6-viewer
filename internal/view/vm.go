package view

type VMTS6Viewer struct {
	VMServer        *VMServer
	VMChannels      []*VMChannel
	Theme           string
	RefreshInterval string
}

type VMServer struct {
	Name               string
	ClientsOnline      string
	MaxClients         string
	UptimePretty       string
	ChannelsOnline     string
	HostBannerURL      string
	HostConnectionLink string
	ClientConnections  string
}

type VMClient struct {
	Nickname    string
	Platform    string
	Version     string
	MicMuted    bool
	OutputMuted bool
	IsTalking   bool
}

type VMChannel struct {
	Name     string
	Type     ChannelType
	Align    Aligned
	Repeat   bool
	Clients  []*VMClient
	Children []*VMChannel
}
