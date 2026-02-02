package ts6

import (
	"fmt"
	"regexp"
	"strings"
)

// Channel represents a TeamSpeak channel
type Channel struct {
	CID               string `json:"cid"`
	PID               string `json:"pid"`
	ChannelOrder      string `json:"channel_order"`
	Name              string `json:"channel_name"`
	Topic             string `json:"channel_topic"`
	FlagPermanent     string `json:"channel_flag_permanent"`
	FlagSemiPermanent string `json:"channel_flag_semi_permanent"`
	FlagDefault       string `json:"channel_flag_default"`
	FlagPassword      string `json:"channel_flag_password"`
	MaxClients        string `json:"channel_maxclients"`
	MaxFamilyClients  string `json:"channel_maxfamilyclients"`
	NeededTalkPower   string `json:"channel_needed_talk_power"`

	// Additional informations, that wil added after response parsing
	Type      ChannelType `json:"-"`
	Align     Aligned     `json:"-"`
	FullWidth bool        `json:"-"`

	Children []*Channel `json:"-"`

	Clients []*Client `json:"-"`
}

// ChannelListResponse represents the TS6 API response
type ChannelListResponse struct {
	Body   []Channel `json:"body"`
	Status Status    `json:"status"`
}

// ChannelType defines the type of a channel or spacer
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

// Aligned defines the alignment for aligned spacers
type Aligned int

const (
	AlignLeft Aligned = iota
	AlignCenter
	AlignRight
)

// GetChannelList returns all channels for a virtual server
func GetChannelList(baseURL, apiKey string, serverID string) ([]Channel, error) {
	var resp ChannelListResponse

	err := doGET(
		baseURL,
		apiKey,
		fmt.Sprintf("/%s/channellist", serverID),
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if resp.Status.Code != 0 {
		return nil, fmt.Errorf("ts6 error %d: %s", resp.Status.Code, resp.Status.Message)
	}

	// Parse each channel name and set the additional fields
	for i := range resp.Body {
		ch := &resp.Body[i]

		chType, align, fullWidth, resolved := ParseChannelName(ch.Name)

		ch.Type = chType
		ch.Align = align
		ch.FullWidth = fullWidth
		ch.Name = resolved
	}

	return resp.Body, nil
}

// GetChannelTree returns all channels including the online clients in a tree structure for a virtual server
func GetChannelTree(baseURL, apiKey string, serverID string, clients []Client) ([]*Channel, error) {
	channels, err := GetChannelList(baseURL, apiKey, serverID)
	if err != nil {
		return nil, err
	}

	channelHierachy := buildChannelTree(clients, channels)
	return channelHierachy, nil
}

func buildChannelTree(clients []Client, channels []Channel) []*Channel {
	lookup := make(map[string]*Channel)
	for i := range channels {
		lookup[channels[i].CID] = &channels[i]
	}

	for i := range clients {
		cl := &clients[i]
		if ch, ok := lookup[cl.CID]; ok {
			ch.Clients = append(ch.Clients, cl)
		}
	}

	var roots []*Channel

	for i := range channels {
		ch := &channels[i]

		if ch.PID == "0" {
			roots = append(roots, ch)
		} else {
			parent, ok := lookup[ch.PID]
			if ok {
				parent.Children = append(parent.Children, ch)
			}
		}
	}

	return roots
}

// ParseChannelName parses TeamSpeak spacer syntax.
// It removes spacer commands and expands spacer patterns to a fixed width.
func ParseChannelName(name string) (ChannelType, Aligned, bool, string) {
	name = strings.TrimSpace(name)

	reCmd := regexp.MustCompile(`(?i)\[(c|l|r)?spacer\d*\]`)
	cmd := reCmd.FindStringSubmatch(name)

	if len(cmd) > 0 {
		align := AlignLeft
		if len(cmd) > 1 {
			switch strings.ToLower(cmd[1]) {
			case "c":
				align = AlignCenter
			case "r":
				align = AlignRight
			}
		}

		// Remove TS spacer command only
		clean := strings.TrimSpace(reCmd.ReplaceAllString(name, ""))

		// Blank spacer
		if clean == "" {
			return BlankSpacer, align, false, ""
		}

		// Line spacer (___ --- ...)
		if isSpacerPattern(clean) {
			return SolidSpacer, align, true, clean
		}

		// Aligned spacer with text
		return AlignedSpacer, align, false, clean
	}

	return NormalChannel, AlignLeft, false, name
}

func isSpacerPattern(s string) bool {
	switch s {
	case "___", "---", "...", "-.-", "-..":
		return true
	default:
		return false
	}
}

// repeatToWidth repeats a spacer pattern until a fixed width is filled.
// Width is approximated using a monospace character width.
func repeatToWidth(pattern string, targetPx int) string {
	const charWidth = 8 // px per character (approximation)

	if pattern == "" {
		return ""
	}

	patternWidth := len([]rune(pattern)) * charWidth
	if patternWidth == 0 {
		return pattern
	}

	repeats := targetPx / patternWidth
	if repeats < 1 {
		repeats = 1
	}

	return strings.Repeat(pattern, repeats)
}
