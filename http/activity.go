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

// activityLimit caps the displayed list (set from config); 0 = unlimited.
var activityLimit = 10

// eventTime returns the timestamp that orders an entry in the merged,
// descending activity list: the connection start for online clients, the
// last confirmed-online time for offline clients (both what is displayed).
func (e *activityEntry) eventTime() int64 {
	if e.offlineAt == 0 && e.connectedAtUnix > 0 {
		return e.connectedAtUnix
	}
	return e.onlineAt
}

// updateActivity reconciles the rolling log with the currently online clients
// and returns the merged list to display: online and offline entries in one
// list, most recent event first (online: connection time, offline: last
// seen), nickname as tie-break.
func updateActivity(clients []ts6.Client, groupBadges map[string]string, now int64) []*view.VMActivity {

	seen := make(map[string]bool, len(clients))
	for _, c := range clients {
		if c.UniqueIdentifier == "" {
			continue
		}

		seen[c.UniqueIdentifier] = true
		badge := view.BadgeForClient(c.ServerGroups, groupBadges)
		country := view.FlagEmoji(c.Country)

		connectedAtUnix := connectionUnix(c.ConnectionTime)
		// Absolute date + clock time so the merged list stays readable.
		connectedAt := view.MakeDateTime(connectedAtUnix)

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

	// Merge online + offline into one list, most recent event first.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].eventTime() != entries[j].eventTime() {
			return entries[i].eventTime() > entries[j].eventTime()
		}
		return entries[i].nickname < entries[j].nickname
	})

	result := make([]*view.VMActivity, 0, len(entries))
	for _, e := range entries {
		connectedAt := e.connectedAt
		// Online entry without a known connection time: fall back to the
		// last confirmed-online time so the row still carries a timestamp.
		if e.offlineAt == 0 && connectedAt == "" && e.onlineAt > 0 {
			connectedAt = view.MakeDateTime(e.onlineAt)
		}
		result = append(result, &view.VMActivity{
			Nickname:    e.nickname,
			Badge:       e.badge,
			Country:     e.country,
			ConnectedAt: connectedAt,
			Online:      e.offlineAt == 0,
			LastSeenAt:  e.onlineAt,
			LastSeen:    view.MakeLastSeen(e.onlineAt),
		})
	}

	// Cap the displayed list (0 = unlimited). The full log stays persisted,
	// so the cap only affects the display.
	if activityLimit > 0 && len(result) > activityLimit {
		result = result[:activityLimit]
	}

	return result
}
