package view

type VMTS6Viewer struct {
	VMServer        *VMServer
	VMChannels      []*VMChannel
	VMActivity      []*VMActivity
	RefreshInterval string
	MaxWidth        string
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
	Nickname         string
	UID              string
	Badge            string
	Country          string
	ConnectedAt      string
	MicMuted         bool
	OutputMuted      bool
	IsTalking        bool
	OnlineTimePretty string
}

// VMActivity is one entry of the recent-activity list: the client's latest
// known state, whether they are online right now, and (if offline) the last
// time they were confirmed online.
type VMActivity struct {
	Nickname    string
	Badge       string
	Country     string
	ConnectedAt string
	Online      bool
	LastSeenAt  int64
	LastSeen    string
}

type VMChannel struct {
	Name     string
	Type     ChannelType
	Align    Aligned
	Repeat   bool
	Clients  []*VMClient
	Children []*VMChannel
}
