package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reCmd = regexp.MustCompile(`(?i)\[([clr]|\*)?spacer([^\]]*?)\]`)

// ParseChannelName parses TeamSpeak spacer syntax.
func ParseChannelName(name string) (ChannelType, Aligned, bool, string) {
	name = strings.TrimSpace(name)

	cmd := reCmd.FindStringSubmatch(name)
	if len(cmd) > 0 {

		// Alignment
		align := AlignLeft
		switch strings.ToLower(cmd[1]) {
		case "c":
			align = AlignCenter
		case "r":
			align = AlignRight
		}

		// Repeat by * in the command?
		repeat := strings.ToLower(cmd[1]) == "*"

		// Text hinter [xSpacer...] extrahieren
		clean := strings.TrimSpace(reCmd.ReplaceAllString(name, ""))

		// Blank spacer
		if clean == "" {
			return BlankSpacer, align, repeat, ""
		}

		// Solid spacer (___ --- ... -.- -..) → ALWAYS repeat = true
		if isSpacerPattern(clean) {
			return SolidSpacer, align, true, clean
		}

		// Aligned spacer with text
		return AlignedSpacer, align, repeat, clean
	}

	// Normal channel
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

func MakeUptimePretty(secondsStr string) string {
	sec, err := strconv.Atoi(secondsStr)
	if err != nil || sec < 0 {
		return secondsStr
	}

	days := sec / 86400
	sec %= 86400
	hours := sec / 3600
	sec %= 3600
	minutes := sec / 60
	seconds := sec % 60

	return fmt.Sprintf("%dD %02d:%02d:%02d", days, hours, minutes, seconds)
}

func BuildChannelTree(channels []*Channel, clients []*FullClient) []*Channel {
	lookup := make(map[string]*Channel)
	for _, ch := range channels {
		lookup[ch.ID] = ch
	}

	for _, cl := range clients {
		if ch, ok := lookup[cl.ChannelID]; ok {
			ch.Clients = append(ch.Clients, cl)
		}
	}

	var roots []*Channel
	for _, ch := range channels {
		if ch.ParentID == "0" {
			roots = append(roots, ch)
		} else if parent, ok := lookup[ch.ParentID]; ok {
			parent.Children = append(parent.Children, ch)
		}
	}

	return roots
}
