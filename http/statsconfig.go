package http

import (
	"strconv"
	"strings"
	"time"

	"ts6-viewer/internal/config"
	"ts6-viewer/internal/view"
)

// graphRanges are the canonical graph time ranges in ascending order. The
// operator picks the largest one (graph_max_range); all smaller ranges stay
// available as buttons.
var graphRanges = []string{"24h", "7d", "14d", "1m"}

// rangeWindows maps a range key to the history retention it requires.
var rangeWindows = map[string]time.Duration{
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"14d": 14 * 24 * time.Hour,
	"1m":  31 * 24 * time.Hour,
}

// Statistics display settings (config with defaults: graph on, activity on,
// 10 entries [activityLimit, 0 = unlimited], 1m max range → all range
// buttons and a month of retention, graph NOT mirrored onto the main page).
var (
	statsGraphEnabled    = true
	statsActivityEnabled = true
	statsMaxRange        = "1m"
	statsShowOnMainPage  = false
)

// parseStatsConfig applies the operator's statistics settings from the
// config file. Missing values keep the defaults, so old config files keep
// working unchanged. graph_max_range controls both the available buttons and
// the history retention (longer retention only when requested).
func parseStatsConfig(cfg config.Config) {
	if s := strings.ToLower(strings.TrimSpace(cfg.Statistics.Graph)); s == "false" {
		statsGraphEnabled = false
	}
	if s := strings.ToLower(strings.TrimSpace(cfg.Statistics.Activity)); s == "false" {
		statsActivityEnabled = false
	}

	if s := strings.TrimSpace(cfg.Statistics.ActivityLimit); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			activityLimit = n // 0 = unlimited
		}
	}

	if r := strings.ToLower(strings.TrimSpace(cfg.Statistics.MaxRange)); r != "" {
		if window, ok := rangeWindows[r]; ok {
			statsMaxRange = r
			historyWindow = window
		}
	}

	// show_on_main_page defaults to false; only an explicit "true" enables it.
	if s := strings.ToLower(strings.TrimSpace(cfg.Statistics.ShowOnMainPage)); s == "true" {
		statsShowOnMainPage = true
	}
}

// statsViewModel builds the client-facing statistics settings (which sections
// are enabled, which graph time ranges are available, and whether the Online
// Clients graph is mirrored onto the main page).
func statsViewModel() *view.VMStats {
	stats := &view.VMStats{
		GraphEnabled:     statsGraphEnabled,
		ActivityEnabled:  statsActivityEnabled,
		MaxRange:         statsMaxRange,
		ShowOnMainPage:   statsShowOnMainPage,
	}
	for _, r := range graphRanges {
		if rangeWindows[r] <= historyWindow {
			stats.Ranges = append(stats.Ranges, r)
		}
	}
	return stats
}
