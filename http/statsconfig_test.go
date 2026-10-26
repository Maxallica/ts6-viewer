package http

import (
	"testing"

	"ts6-viewer/internal/config"
)

// TestParseStatsConfigDefaults ensures that missing statistics settings keep
// the defaults: graph on, activity on, 1m max range, and the Online Clients
// graph NOT mirrored onto the main page.
func TestParseStatsConfigDefaults(t *testing.T) {
	savedGraph := statsGraphEnabled
	savedActivity := statsActivityEnabled
	savedMaxRange := statsMaxRange
	savedMain := statsShowOnMainPage
	savedLimit := activityLimit
	savedWindow := historyWindow
	defer func() {
		statsGraphEnabled = savedGraph
		statsActivityEnabled = savedActivity
		statsMaxRange = savedMaxRange
		statsShowOnMainPage = savedMain
		activityLimit = savedLimit
		historyWindow = savedWindow
	}()

	parseStatsConfig(config.Config{})

	if !statsGraphEnabled {
		t.Error("expected the graph to be enabled by default")
	}
	if !statsActivityEnabled {
		t.Error("expected recent activity to be enabled by default")
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
	savedWindow := historyWindow
	defer func() {
		statsGraphEnabled = savedGraph
		statsActivityEnabled = savedActivity
		statsMaxRange = savedMaxRange
		activityLimit = savedLimit
		historyWindow = savedWindow
	}()

	cfg := config.Config{}
	cfg.Statistics.Graph = "false"
	cfg.Statistics.Activity = "false"
	cfg.Statistics.MaxRange = "9d" // not a valid range
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
