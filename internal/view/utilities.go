package view

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reCmd = regexp.MustCompile(`(?i)\[([clr]|\*)?spacer([^\]]*?)\]`)

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

// FlagEmoji converts an ISO 3166-1 alpha-2 country code into its flag emoji
// (regional indicator symbols). Returns an empty string for missing or
// invalid codes.
func FlagEmoji(countryCode string) string {
	code := strings.ToUpper(strings.TrimSpace(countryCode))
	if len(code) != 2 {
		return ""
	}
	for i := 0; i < len(code); i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return ""
		}
	}

	const offset = rune(0x1F1E6) - 'A'
	return string([]rune{offset + rune(code[0]), offset + rune(code[1])})
}

// MakeConnectedAt formats the unix connection timestamp as a local "HH:MM"
// clock time. Returns an empty string if the timestamp is missing or invalid.
func MakeConnectedAt(unixSeconds string) string {
	ts, err := strconv.ParseInt(unixSeconds, 10, 64)
	if err != nil || ts <= 0 {
		return ""
	}
	return formatClock(ts)
}

// MakeLastSeen formats a unix timestamp as an absolute "last seen" time:
// "HH:MM" for the current local day, "DD.MM. HH:MM" for an earlier day,
// "DD.MM.YYYY HH:MM" for an earlier year. Returns an empty string if the
// timestamp is missing or invalid.
func MakeLastSeen(unixSeconds int64) string {
	if unixSeconds <= 0 {
		return ""
	}

	t := time.Unix(unixSeconds, 0)
	now := time.Now()

	clock := t.Format("15:04")
	if t.Year() == now.Year() && t.Day() == now.Day() && t.Month() == now.Month() {
		return clock
	}
	if t.Year() == now.Year() {
		return t.Format("02.01.") + " " + clock
	}
	return t.Format("02.01.2006") + " " + clock
}

func formatClock(unixSeconds int64) string {
	return time.Unix(unixSeconds, 0).Format("15:04")
}

// MakeDateTime formats a unix timestamp as an absolute date + clock time:
// "DD.MM. HH:MM" for the current year, "DD.MM.YYYY HH:MM" for an earlier
// year. Unlike MakeLastSeen the date is always shown, so entries from
// different days stay unambiguous in a single sorted list.
// Returns an empty string if the timestamp is missing or invalid.
func MakeDateTime(unixSeconds int64) string {
	if unixSeconds <= 0 {
		return ""
	}

	t := time.Unix(unixSeconds, 0)
	now := time.Now()
	clock := t.Format("15:04")
	if t.Year() == now.Year() {
		return t.Format("02.01.") + " " + clock
	}
	return t.Format("02.01.2006") + " " + clock
}

// MakeOnlineTimePretty calculates how long a client has been online based on
// the unix connection timestamp and formats it as "Xd HH:MM:SS".
// Returns an empty string if the timestamp is missing or invalid.
func MakeOnlineTimePretty(unixSeconds string) string {
	ts, err := strconv.ParseInt(unixSeconds, 10, 64)
	if err != nil || ts <= 0 {
		return ""
	}

	delta := time.Now().Unix() - ts
	if delta < 0 { // guard against clock skew
		delta = 0
	}

	days := delta / 86400
	delta %= 86400
	hours := delta / 3600
	delta %= 3600
	minutes := delta / 60
	seconds := delta % 60

	return fmt.Sprintf("%dd %02d:%02d:%02d", days, hours, minutes, seconds)
}
