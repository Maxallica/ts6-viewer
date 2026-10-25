package view

import (
	"sort"
	"strconv"
	"strings"
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

// GroupBadgeMap maps each servergroup ID to a badge label ("admin" or "vip")
// derived from the group name. Groups whose name matches neither are omitted.
func GroupBadgeMap(groups []ts6.ServerGroup) map[string]string {
	badges := make(map[string]string, len(groups))
	for _, g := range groups {
		if badge := BadgeForGroup(g.Name); badge != "" {
			badges[g.ID] = badge
		}
	}
	return badges
}

// BadgeForGroup derives a badge label from a servergroup name: "admin" if the
// name contains "admin", "vip" if it contains "vip", otherwise ""
// (case-insensitive).
func BadgeForGroup(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "admin") {
		return "admin"
	}
	if strings.Contains(lower, "vip") {
		return "vip"
	}
	return ""
}

func BuildVMChannels(channels []ts6.Channel, clients []ts6.Client, groupBadges map[string]string) []*VMChannel {
	// Channels
	viewMap := make(map[string]*VMChannel)
	for _, ch := range channels {
		viewMap[ch.CID] = BuildVMChannel(ch)
	}

	// Clients
	for _, c := range clients {
		if vch, ok := viewMap[c.CID]; ok {
			vch.Clients = append(vch.Clients, BuildVMClient(c, groupBadges))
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

// BadgeForClient derives the client's badge label from their servergroups:
// "admin" if any group is an admin group, otherwise "vip" if any group is a
// vip group, otherwise "" (priority: admin > vip).
func BadgeForClient(serverGroups string, groupBadges map[string]string) string {
	hasVIP := false
	for _, id := range strings.Split(serverGroups, ",") {
		switch groupBadges[strings.TrimSpace(id)] {
		case "admin":
			return "admin"
		case "vip":
			hasVIP = true
		}
	}
	if hasVIP {
		return "vip"
	}
	return ""
}

func BuildVMClient(c ts6.Client, groupBadges map[string]string) *VMClient {
	return &VMClient{
		Nickname:         c.Nickname,
		UID:              c.UniqueIdentifier,
		Badge:            BadgeForClient(c.ServerGroups, groupBadges),
		Country:          FlagEmoji(c.Country),
		ConnectedAt:      MakeConnectedAt(c.ConnectionTime),
		MicMuted:         c.InputMuted == "1" || c.InputHardware == "0",
		OutputMuted:      c.OutputMuted == "1",
		IsTalking:        c.IsTalking == "1",
		OnlineTimePretty: MakeOnlineTimePretty(c.ConnectionTime),
	}
}
