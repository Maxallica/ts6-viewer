package view

type VMTS6Viewer struct {
	VMServer        *VMServer
	VMChannels      []*VMChannel
	VMActivity      []*VMActivity
	VMHistory       []*VMHistoryPoint
	VMStats         *VMStats
	RefreshInterval string
	MaxWidth        string
}

// VMStats carries the operator's statistics display settings to the client:
// which sections are enabled, which graph time ranges are available, and
// whether the Online Clients graph is mirrored onto the main page.
type VMStats struct {
	GraphEnabled     bool
	ActivityEnabled  bool
	MaxRange         string   // one of "24h" | "7d" | "14d" | "1m"
	Ranges           []string // enabled ranges, ordered ascending (subset of the above)
	ShowOnMainPage   bool     // also render the Online Clients graph below the server info
}

// VMHistoryPoint is one observation of the rolling online-count graph:
// the unix time the count was taken and how many clients were online.
type VMHistoryPoint struct {
	T int64
	N int
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
