package http

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

func activityClient(nickname, uid string) ts6.Client {
	return ts6.Client{Nickname: nickname, UniqueIdentifier: uid, ConnectionTime: "1700000000"}
}

func activityClientAt(nickname, uid string, connectedAtUnix int64) ts6.Client {
	return ts6.Client{
		Nickname:       nickname,
		UniqueIdentifier: uid,
		ConnectionTime: fmt.Sprint(connectedAtUnix),
	}
}

// TestConnectionUnix pins the raw-timestamp parsing: valid values pass
// through, missing or invalid values map to 0.
func TestConnectionUnix(t *testing.T) {
	if got := connectionUnix("1700000000"); got != 1700000000 {
		t.Errorf("connectionUnix(1700000000) = %d, want 1700000000", got)
	}
	for _, raw := range []string{"", "abc", "-5", "1.5"} {
		if got := connectionUnix(raw); got != 0 {
			t.Errorf("connectionUnix(%q) = %d, want 0", raw, got)
		}
	}
}

// TestUpdateActivityOnlineDescending checks that the online group is sorted
// by connection time, most recently connected first (nickname tie-break).
func TestUpdateActivityOnlineDescending(t *testing.T) {
	saved := activityLog
	activityLog = make(map[string]*activityEntry)
	defer func() { activityLog = saved }()

	now := time.Now().Unix()
	clients := []ts6.Client{
		activityClientAt("Early", "uid-early", now-7200),
		activityClientAt("Late", "uid-late", now-60),
		activityClientAt("Mid", "uid-mid", now-1800),
	}

	result := updateActivity(clients, nil, now)

	want := []string{"Late", "Mid", "Early"} // most recent connection first
	if len(result) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(result), len(want), result)
	}
	for i, w := range want {
		if result[i].Nickname != w {
			t.Errorf("position %d: nickname = %q, want %q", i, result[i].Nickname, w)
		}
		if !result[i].Online {
			t.Errorf("position %d: expected online", i)
		}
		// Online entries must carry an absolute date + clock time.
		if result[i].ConnectedAt == "" || strings.Contains(result[i].ConnectedAt, ":") == false {
			t.Errorf("position %d: expected absolute date+time, got %q", i, result[i].ConnectedAt)
		}
	}
}

// TestUpdateActivityOrderAndAbsoluteTime verifies that updateActivity returns
// recently offline entries first (most recent offline first, nickname
// tie-break), then online entries (most recently connected first, nickname
// tie-break), and that every "last seen" value is an absolute time with no
// relative phrasing.
func TestUpdateActivityOrderAndAbsoluteTime(t *testing.T) {
	saved := activityLog
	activityLog = make(map[string]*activityEntry)
	defer func() { activityLog = saved }()

	now := time.Now().Unix()
	t2 := now
	t1 := now - 3600
	t0 := now - 7200

	allFour := []ts6.Client{
		activityClient("Alice", "uid-alice"),
		activityClient("Bob", "uid-bob"),
		activityClient("Carol", "uid-carol"),
		activityClient("Dave", "uid-dave"),
	}
	aliceBobDave := []ts6.Client{
		activityClient("Alice", "uid-alice"),
		activityClient("Bob", "uid-bob"),
		activityClient("Dave", "uid-dave"),
	}
	aliceBob := []ts6.Client{
		activityClient("Alice", "uid-alice"),
		activityClient("Bob", "uid-bob"),
	}

	// Everyone online at t0; Carol drops at t1; Dave drops at t2.
	updateActivity(allFour, nil, t0)
	updateActivity(aliceBobDave, nil, t1)
	result := updateActivity(aliceBob, nil, t2)

	want := []struct {
		name   string
		online bool
	}{
		{"Dave", false},  // most recently offline → top
		{"Carol", false}, // offline before Dave
		{"Alice", true},  // online → bottom
		{"Bob", true},
	}

	if len(result) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(result), len(want), result)
	}

	for i, w := range want {
		if result[i].Nickname != w.name {
			t.Errorf("position %d: nickname = %q, want %q", i, result[i].Nickname, w.name)
		}
		if result[i].Online != w.online {
			t.Errorf("position %d: online = %v, want %v", i, result[i].Online, w.online)
		}
		if result[i].LastSeen == "" {
			t.Errorf("position %d: expected non-empty LastSeen", i)
		}
		if strings.Contains(result[i].LastSeen, "ago") || strings.Contains(result[i].LastSeen, "just now") {
			t.Errorf("position %d: LastSeen %q must be absolute, not relative", i, result[i].LastSeen)
		}
	}

	// Offline last-seen values must come from the last online observation.
	if got, wantLast := result[0].LastSeen, view.MakeLastSeen(t1); got != wantLast {
		t.Errorf("Dave LastSeen = %q, want %q", got, wantLast)
	}
	if got, wantLast := result[1].LastSeen, view.MakeLastSeen(t0); got != wantLast {
		t.Errorf("Carol LastSeen = %q, want %q", got, wantLast)
	}
}

// TestUpdateActivityRejoin covers the regression where a client that went
// offline and then came back online would keep its offline marker forever:
// the offline flag must be cleared on the next online observation.
func TestUpdateActivityRejoin(t *testing.T) {
	saved := activityLog
	activityLog = make(map[string]*activityEntry)
	defer func() { activityLog = saved }()

	now := time.Now().Unix()
	t0 := now - 3600
	t1 := now - 1800

	solo := []ts6.Client{activityClient("Mia", "uid-mia")}

	updateActivity(solo, nil, t0) // Mia online
	updateActivity(nil, nil, t1)  // Mia offline (first missing observation)
	result := updateActivity(solo, nil, now) // Mia back online

	if len(result) != 1 {
		t.Fatalf("got %d entries, want 1: %+v", len(result), result)
	}
	if !result[0].Online {
		t.Errorf("Mia must be online after rejoining, got %+v", result[0])
	}
}

// TestUpdateActivityOfflineTieBreak checks the nickname tie-break when two
// clients go offline in the same observation.
func TestUpdateActivityOfflineTieBreak(t *testing.T) {
	saved := activityLog
	activityLog = make(map[string]*activityEntry)
	defer func() { activityLog = saved }()

	now := time.Now().Unix()
	t0 := now - 3600

	both := []ts6.Client{
		activityClient("Zed", "uid-zed"),
		activityClient("Amy", "uid-amy"),
	}

	updateActivity(both, nil, t0)
	result := updateActivity(nil, nil, now)

	// Same offlineAt → alphabetical: Amy before Zed.
	if len(result) != 2 || result[0].Nickname != "Amy" || result[1].Nickname != "Zed" {
		t.Fatalf("unexpected tie-break order: %+v", result)
	}
}
