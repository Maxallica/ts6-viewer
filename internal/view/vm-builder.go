package view

import (
	"sort"
	"strconv"
	"ts6-viewer/internal/config"
	"ts6-viewer/internal/ts6"
)

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

func BuildVMChannels(channels []ts6.Channel, clients []ts6.Client) []*VMChannel {
	// Channels
	viewMap := make(map[string]*VMChannel)
	for _, ch := range channels {
		viewMap[ch.CID] = BuildVMChannel(ch)
	}

	// Clients
	for _, c := range clients {
		if vch, ok := viewMap[c.CID]; ok {
			vch.Clients = append(vch.Clients, BuildVMClient(c))
		}
	}

	// Sort clients in each channel alphabetically
	for _, vch := range viewMap {
		sort.Slice(vch.Clients, func(i, j int) bool {
			return vch.Clients[i].Nickname < vch.Clients[j].Nickname
		})
	}

	// Build tree
	var roots []*VMChannel
	for _, ch := range channels {
		vch := viewMap[ch.CID]
		if ch.PID == "0" {
			roots = append(roots, vch)
		} else if parent, ok := viewMap[ch.PID]; ok {
			parent.Children = append(parent.Children, vch)
		}
	}

	return roots
}

func BuildVMServer(cfg *config.Config, info *ts6.ServerInfo, clients []ts6.Client) *VMServer {
	return &VMServer{
		Name:               info.Name,
		ClientsOnline:      strconv.Itoa(len(clients)),
		MaxClients:         info.MaxClients,
		UptimePretty:       MakeUptimePretty(info.Uptime),
		ChannelsOnline:     info.ChannelsOnline,
		HostBannerURL:      info.HostBannerURL,
		HostConnectionLink: cfg.HostConnectionLink,
		ClientConnections:  info.ClientConnections,
	}
}

func BuildVMChannel(ch ts6.Channel) *VMChannel {
	chType, align, repeat, cleanName := ParseChannelName(ch.Name)

	return &VMChannel{
		Name:   cleanName,
		Type:   chType,
		Align:  align,
		Repeat: repeat,
	}
}

func BuildVMClient(c ts6.Client) *VMClient {
	return &VMClient{
		Nickname:    c.Nickname,
		MicMuted:    c.InputMuted == "1" || c.InputHardware == "0",
		OutputMuted: c.OutputMuted == "1",
		IsTalking:   c.IsTalking == "1",
	}
}
