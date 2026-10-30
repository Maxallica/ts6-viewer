package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort         string `json:"server_port"`
	HostConnectionLink string `json:"host_connection_link"`

	Teamspeak6 struct {
		Host              string `json:"host"`
		Port              string `json:"port"`
		User              string `json:"user"`
		Password          string `json:"password"`
		EnableVoiceStatus string `json:"enable_voice_status"`
		ServerID          string `json:"server_id"`
	} `json:"teamspeak6"`

	RefreshInterval string `json:"refresh_interval"`
	MaxWidth        string `json:"max_width"`

	// Statistics panel settings. All values are optional; missing values fall
	// back to defaults (graph on, recent activity on, 10 entries, 1m graph
	// range, 24h activity range, graph NOT mirrored onto the main page).
	Statistics struct {
		Graph            string `json:"graph"`                     // "true" / "false" (default "true")
		MaxRange         string `json:"graph_max_range"`           // "24h" | "7d" | "14d" | "1m" (default "1m")
		Activity         string `json:"recent_activity"`           // "true" / "false" (default "true")
		ActivityMaxRange string `json:"recent_activity_max_range"` // "24h" | "7d" | "14d" | "1m" (default "24h")
		ActivityLimit    string `json:"recent_activity_limit"`     // entry count, "0" = unlimited (default "10")
		ShowOnMainPage   string `json:"show_on_main_page"`         // "true" / "false" — also show the Online Clients graph on the main page (default "false")
	} `json:"statistics"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
