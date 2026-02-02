package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort string `json:"server_port"`
	Theme      string `json:"theme"`

	Teamspeak6 struct {
		BaseURL  string `json:"base_url"`
		ApiKey   string `json:"api_key"`
		ServerID string `json:"server_id"`
	} `json:"teamspeak6"`

	RefreshInterval string `json:"refresh_interval"`
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
