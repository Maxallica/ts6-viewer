package http

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"testing"
	"time"

	"ts6-viewer/internal/view"
)

// TestStatsPanelOfflineRender ensures the stats panel template renders both
// online and offline activity entries without a template execution error
// (e.g. a field that only exists on one struct).
func TestStatsPanelOfflineRender(t *testing.T) {
	tmpl := template.Must(template.ParseFiles("../internal/web/templates/ts6viewer.html"))

	// A fixed timestamp from an earlier day of the current year, formatted
	// through the same absolute helper the app uses ("DD.MM. HH:MM").
	lastSeenAt := time.Now().Add(-3 * 24 * time.Hour).Unix()
	lastSeen := view.MakeLastSeen(lastSeenAt)

	data := view.VMTS6Viewer{
		VMServer: &view.VMServer{
			Name:           "Test",
			ClientsOnline:  "2",
			MaxClients:     "32",
			UptimePretty:   "1 h",
			ChannelsOnline: "10",
		},
		RefreshInterval: "60",
		MaxWidth:        "1000px",
		VMStats: &view.VMStats{
			GraphEnabled:    true,
			ActivityEnabled: true,
			MaxRange:        "7d",
			Ranges:          []string{"24h", "7d"},
		},
		VMActivity: []*view.VMActivity{
			{Nickname: "Live", Badge: "admin", Country: "DE", ConnectedAt: "10:00", Online: true},
			{Nickname: "Ghost", Badge: "vip", Country: "FR", ConnectedAt: "09:00", Online: false, LastSeenAt: lastSeenAt, LastSeen: lastSeen},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Live", "Ghost",
		"ADMIN", "VIP",
		"since", "last seen", lastSeen,
		// Graph signboard below the server info with its range buttons.
		"graphPanel", "graphRanges", "data-ranges=\"24h,7d\"",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}

	// The status color is on the name (log style); the timestamp stays neutral.
	if !strings.Contains(out, `class="stat-name online"`) {
		t.Error("expected online name marker class in output")
	}
	if !strings.Contains(out, `class="stat-name offline"`) {
		t.Error("expected offline name marker class in output")
	}
}

// TestStatsSectionsDisabled ensures that when the operator disables both
// statistics sections, neither the graph signboard, the stats panel, nor the
// toggle button is rendered at all.
func TestStatsSectionsDisabled(t *testing.T) {
	tmpl := template.Must(template.ParseFiles("../internal/web/templates/ts6viewer.html"))

	data := view.VMTS6Viewer{
		VMServer: &view.VMServer{Name: "Test"},
		VMStats: &view.VMStats{
			GraphEnabled:    false,
			ActivityEnabled: false,
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	out := buf.String()
	// The style block still references the IDs, so check the elements.
	for _, absent := range []string{`id="graphPanel"`, `id="statsPanel"`, `id="statsToggle"`} {
		if strings.Contains(out, absent) {
			t.Errorf("expected %q to be absent when both sections are disabled", absent)
		}
	}
}

// TestMakeLastSeenAbsolute pins the absolute "last seen" formats: HH:MM for
// the current local day, DD.MM. HH:MM for an earlier day of the same year,
// DD.MM.YYYY HH:MM for an earlier year. No relative phrasing allowed.
func TestMakeLastSeenAbsolute(t *testing.T) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 9, 7, 0, 0, now.Location())
	earlierDayWant := "09:07"
	if now.Day() != 1 {
		earlierDayWant = fmt.Sprintf("01.%02d. 09:07", int(now.Month()))
	}

	cases := []struct {
		name string
		ts   time.Time
		want string
	}{
		{"today", dayStart.Add(2*time.Hour + 5*time.Minute), dayStart.Add(2*time.Hour + 5*time.Minute).Format("15:04")},
		{"earlier day same year", firstOfMonth, earlierDayWant},
		{"earlier year", time.Date(now.Year()-1, now.Month(), now.Day(), 8, 30, 0, 0, now.Location()), fmt.Sprintf("%02d.%02d.%d 08:30", now.Day(), int(now.Month()), now.Year()-1)},
		{"invalid", time.Unix(0, 0), ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := view.MakeLastSeen(tc.ts.Unix())
			if got != tc.want {
				t.Fatalf("MakeLastSeen(%d) = %q, want %q", tc.ts.Unix(), got, tc.want)
			}
			if strings.Contains(got, "ago") || strings.Contains(got, "just now") {
				t.Fatalf("MakeLastSeen(%d) = %q, must not contain relative phrasing", tc.ts.Unix(), got)
			}
		})
	}
}
