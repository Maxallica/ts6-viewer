package http

import (
	"strings"
	"testing"
	"time"

	"ts6-viewer/internal/ts6"
	"ts6-viewer/internal/view"
)

func activityClient(nickname, uid string) ts6.Client {
	return ts6.Client{Nickname: nickname, UniqueIdentifier: uid, ConnectionTime: "1700000000"}
}

// TestUpdateActivityOrderAndAbsoluteTime verifies that updateActivity returns
// recently offline entries first (most recent offline first, nickname
// tie-break), then online entries (by nickname), and that every "last seen"
// value is an absolute time with no relative phrasing.
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
