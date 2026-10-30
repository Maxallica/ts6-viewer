package http

import (
	"testing"
	"time"

	"ts6-viewer/internal/config"
)

// TestParseStatsConfigDefaults ensures that missing statistics settings keep
// the defaults: graph on, activity on, 24h activity max range, 1m graph max
// range, and the Online Clients graph NOT mirrored onto the main page.
func TestParseStatsConfigDefaults(t *testing.T) {
	savedGraph := statsGraphEnabled
	savedActivity := statsActivityEnabled
	savedMaxRange := statsMaxRange
	savedMain := statsShowOnMainPage
	savedLimit := activityLimit
	savedActivityMaxRange := statsActivityMaxRange
	savedRetention := activityWindow
	savedWindow := historyWindow
	defer func() {
		statsGraphEnabled = savedGraph
		statsActivityEnabled = savedActivity
		statsMaxRange = savedMaxRange
		statsShowOnMainPage = savedMain
		activityLimit = savedLimit
		statsActivityMaxRange = savedActivityMaxRange
		activityWindow = savedRetention
		historyWindow = savedWindow
	}()

	parseStatsConfig(config.Config{})

	if !statsGraphEnabled {
		t.Error("expected the graph to be enabled by default")
	}
	if !statsActivityEnabled {
		t.Error("expected recent activity to be enabled by default")
	}
	if statsActivityMaxRange != "24h" {
		t.Errorf("expected default activity max range 24h, got %q", statsActivityMaxRange)
	}
	if activityWindow != 24*time.Hour {
		t.Errorf("expected default activity retention 24h, got %v", activityWindow)
	}
	if statsMaxRange != "1m" {
		t.Errorf("expected default max range 1m, got %q", statsMaxRange)
	}
	if statsShowOnMainPage {
		t.Error("expected show_on_main_page to be disabled by default")
	}
	if activityLimit != 10 {
		t.Errorf("expected default activity limit 10, got %d", activityLimit)
	}
}

// TestParseStatsConfigShowOnMainPage pins the show_on_main_page rule: it is
// disabled by default and only an explicit "true" enables the main-page
// mirror of the Online Clients graph (unlike the other toggles, whose
// default is on).
func TestParseStatsConfigShowOnMainPage(t *testing.T) {
	saved := statsShowOnMainPage
	defer func() { statsShowOnMainPage = saved }()

	statsShowOnMainPage = false
	parseStatsConfig(config.Config{})
	if statsShowOnMainPage {
		t.Fatal("expected show_on_main_page to stay disabled with an empty config")
	}

	statsShowOnMainPage = false
	cfg := config.Config{}
	cfg.Statistics.ShowOnMainPage = "true"
	parseStatsConfig(cfg)
	if !statsShowOnMainPage {
		t.Error(`expected explicit "true" to enable show_on_main_page`)
	}
}

// TestParseStatsConfigToggles pins the other statistics rules: explicit
// "false" disables graph and activity, an invalid max range keeps the
// default, and an invalid activity limit is ignored.
func TestParseStatsConfigToggles(t *testing.T) {
	savedGraph := statsGraphEnabled
	savedActivity := statsActivityEnabled
	savedMaxRange := statsMaxRange
	savedLimit := activityLimit
	savedActivityMaxRange := statsActivityMaxRange
	savedRetention := activityWindow
	savedWindow := historyWindow
	defer func() {
		statsGraphEnabled = savedGraph
		statsActivityEnabled = savedActivity
		statsMaxRange = savedMaxRange
		activityLimit = savedLimit
		statsActivityMaxRange = savedActivityMaxRange
		activityWindow = savedRetention
		historyWindow = savedWindow
	}()

	cfg := config.Config{}
	cfg.Statistics.Graph = "false"
	cfg.Statistics.Activity = "false"
	cfg.Statistics.MaxRange = "9d"       // not a valid range
	cfg.Statistics.ActivityLimit = "abc" // not a number

	parseStatsConfig(cfg)

	if statsGraphEnabled {
		t.Error(`expected the graph disabled by explicit "false"`)
	}
	if statsActivityEnabled {
		t.Error(`expected activity disabled by explicit "false"`)
	}
	if statsMaxRange != "1m" {
		t.Errorf("expected an invalid range to keep the default 1m, got %q", statsMaxRange)
	}
	if activityLimit != 10 {
		t.Errorf("expected an invalid limit to keep the default 10, got %d", activityLimit)
	}
}

// TestParseStatsConfigActivityMaxRange pins the recent_activity_max_range
// rule: each valid key maps to its retention duration (case-insensitive),
// and an invalid key keeps the 24h default.
func TestParseStatsConfigActivityMaxRange(t *testing.T) {
	savedKey := statsActivityMaxRange
	savedRetention := activityWindow
	defer func() {
		statsActivityMaxRange = savedKey
		activityWindow = savedRetention
	}()

	cases := []struct {
		key     string
		wantKey string
		wantDur time.Duration
	}{
		{"24h", "24h", 24 * time.Hour},
		{"7d", "7d", 7 * 24 * time.Hour},
		{"14d", "14d", 14 * 24 * time.Hour},
		{"1m", "1m", 31 * 24 * time.Hour},
		{" 1M ", "1m", 31 * 24 * time.Hour}, // case-insensitive, trimmed
		{"48h", "24h", 24 * time.Hour},      // not a graph range → default kept
		{"9d", "24h", 24 * time.Hour},       // invalid → default kept
		{"", "24h", 24 * time.Hour},         // missing → default kept
	}

	for _, c := range cases {
		statsActivityMaxRange = "24h"
		activityWindow = 24 * time.Hour
		cfg := config.Config{}
		cfg.Statistics.ActivityMaxRange = c.key
		parseStatsConfig(cfg)
		if statsActivityMaxRange != c.wantKey {
			t.Errorf("key %q: range key = %q, want %q", c.key, statsActivityMaxRange, c.wantKey)
		}
		if activityWindow != c.wantDur {
			t.Errorf("key %q: retention = %v, want %v", c.key, activityWindow, c.wantDur)
		}
	}
}

// TestStatsViewModelActivityRanges pins the button rule: only the ranges
// whose retention fits the configured max range are offered, in ascending
// order — so the default 24h shows a single button and 1m shows all four.
func TestStatsViewModelActivityRanges(t *testing.T) {
	savedKey := statsActivityMaxRange
	savedRetention := activityWindow
	defer func() {
		statsActivityMaxRange = savedKey
		activityWindow = savedRetention
	}()

	cases := []struct {
		key  string
		want []string
	}{
		{"24h", []string{"24h"}},
		{"7d", []string{"24h", "7d"}},
		{"14d", []string{"24h", "7d", "14d"}},
		{"1m", []string{"24h", "7d", "14d", "1m"}},
	}

	for _, c := range cases {
		statsActivityMaxRange = c.key
		activityWindow = rangeWindows[c.key]
		stats := statsViewModel()
		if len(stats.ActivityRanges) != len(c.want) {
			t.Errorf("max range %q: ranges = %v, want %v", c.key, stats.ActivityRanges, c.want)
			continue
		}
		for i := range c.want {
			if stats.ActivityRanges[i] != c.want[i] {
				t.Errorf("max range %q: range %d = %q, want %q", c.key, i, stats.ActivityRanges[i], c.want[i])
				break
			}
		}
		if stats.ActivityMaxRange != c.key {
			t.Errorf("max range %q: model = %q", c.key, stats.ActivityMaxRange)
		}
	}
}
