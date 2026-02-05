package domain

type ChannelType int

const (
	NormalChannel    ChannelType = iota // Regular TS6 channel
	SolidSpacer                         // ___
	DashSpacer                          // ---
	DotSpacer                           // ...
	DashDotSpacer                       // -.-
	DashDotDotSpacer                    // -..
	AlignedSpacer                       // [cSpacer#], [lSpacer#], [rSpacer#]
	RepeatingSpacer                     // [*spacer#]X
	BlankSpacer                         // Empty spacer like [cSpacer0]
)

type Aligned int

const (
	AlignLeft Aligned = iota
	AlignCenter
	AlignRight
)

type Channel struct {
	ID       string
	ParentID string
	Name     string
	Topic    string

	Type   ChannelType
	Align  Aligned
	Repeat bool

	Children []*Channel
	Clients  []*FullClient
}

type Client struct {
	ID        string
	Nickname  string
	ChannelID string
}

type ClientInfo struct {
	MicMuted    bool
	OutputMuted bool
	IsTalking   bool
}

type FullClient struct {
	Client
	Info *ClientInfo
}

type Server struct {
	Name              string
	ClientsOnline     string
	MaxClients        string
	UptimePretty      string
	ChannelsOnline    string
	HostBannerURL     string
	ClientConnections string
}

type ViewerData struct {
	Server      *Server
	ChannelTree []*Channel
}
