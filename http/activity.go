package http

import (
	"sort"
	"strconv"
	"time"
	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

// activityWindow is how long recent-activity entries are kept.
const activityWindow = 24 * time.Hour

// activityEntry is the latest known state of a client, keyed by UID.
type activityEntry struct {
	nickname        string
	badge           string
	country         string
	connectedAt     string
	connectedAtUnix int64 // raw connection timestamp (unix seconds, 0 = unknown)
	onlineAt        int64 // last time seen online (unix seconds)
	offlineAt       int64 // time first seen missing (0 = currently online)
}

// connectionUnix parses the unix connection timestamp returned by the
// TS6 client list. Returns 0 for missing or invalid values.
func connectionUnix(raw string) int64 {
	ts, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || ts <= 0 {
		return 0
	}
	return ts
}

// activityLog is the rolling 24h recent-activity list. Guarded by mu
// (getViewerData holds it for the whole fetch).
var activityLog = make(map[string]*activityEntry)

// updateActivity reconciles the rolling log with the currently online clients
// and returns the list to display: recently offline clients first (most recent
// offline first), then online clients (most recently connected first).
func updateActivity(clients []ts6.Client, groupBadges map[string]string, now int64) []*view.VMActivity {

	seen := make(map[string]bool, len(clients))
	for _, c := range clients {
		if c.UniqueIdentifier == "" {
			continue
		}

		seen[c.UniqueIdentifier] = true
		badge := view.BadgeForClient(c.ServerGroups, groupBadges)
		country := view.FlagEmoji(c.Country)
		connectedAt := view.MakeConnectedAt(c.ConnectionTime)

		connectedAtUnix := connectionUnix(c.ConnectionTime)

		e, ok := activityLog[c.UniqueIdentifier]
		if !ok {
			activityLog[c.UniqueIdentifier] = &activityEntry{
				nickname:        c.Nickname,
				badge:           badge,
				country:         country,
				connectedAt:     connectedAt,
				connectedAtUnix: connectedAtUnix,
				onlineAt:        now,
			}
			continue
		}

		e.nickname = c.Nickname
		e.badge = badge
		e.country = country
		e.connectedAt = connectedAt
		e.connectedAtUnix = connectedAtUnix
		e.onlineAt = now
		// Client is back online: clear a previous offline marker so it
		// leaves the offline section of the list.
		e.offlineAt = 0
	}

	// Mark newly offline clients (first missing observation).
	for key, e := range activityLog {
		if !seen[key] && e.offlineAt == 0 {
			e.offlineAt = now
		}
	}

	// Purge entries inactive for the whole window.
	cutoff := now - int64(activityWindow/time.Second)
	for key, e := range activityLog {
		last := e.onlineAt
		if e.offlineAt > last {
			last = e.offlineAt
		}
		if last < cutoff {
			delete(activityLog, key)
		}
	}

	entries := make([]*activityEntry, 0, len(activityLog))
	for _, e := range activityLog {
		entries = append(entries, e)
	}

	online := make([]*activityEntry, 0)
	offline := make([]*activityEntry, 0)
	for _, e := range entries {
		if e.offlineAt == 0 {
			online = append(online, e)
		} else {
			offline = append(offline, e)
		}
	}
	sort.Slice(online, func(i, j int) bool {
		if online[i].connectedAtUnix != online[j].connectedAtUnix {
			return online[i].connectedAtUnix > online[j].connectedAtUnix // most recent connection first
		}
		return online[i].nickname < online[j].nickname
	})
	sort.Slice(offline, func(i, j int) bool {
		if offline[i].offlineAt != offline[j].offlineAt {
			return offline[i].offlineAt > offline[j].offlineAt // most recent first
		}
		return offline[i].nickname < offline[j].nickname
	})

	result := make([]*view.VMActivity, 0, len(entries))
	for _, e := range offline {
		result = append(result, &view.VMActivity{
			Nickname:    e.nickname,
			Badge:       e.badge,
			Country:     e.country,
			ConnectedAt: e.connectedAt,
			Online:      false,
			LastSeenAt:  e.onlineAt,
			LastSeen:    view.MakeLastSeen(e.onlineAt),
		})
	}
	for _, e := range online {
		result = append(result, &view.VMActivity{
			Nickname:    e.nickname,
			Badge:       e.badge,
			Country:     e.country,
			ConnectedAt: e.connectedAt,
			Online:      true,
			LastSeenAt:  e.onlineAt,
			LastSeen:    view.MakeLastSeen(e.onlineAt),
		})
	}

	return result
}
