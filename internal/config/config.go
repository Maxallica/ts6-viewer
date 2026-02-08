package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort string `json:"server_port"`
	Theme      string `json:"theme"`

	Teamspeak6 struct {
		BaseURL                  string `json:"base_url"`
		ApiKey                   string `json:"api_key"`
		Host                     string `json:"host"`
		Port                     string `json:"port"`
		User                     string `json:"user"`
		Password                 string `json:"password"`
		Mode                     string `json:"mode"`
		EnableDetailedClientInfo string `json:"enable_detailed_client_info"`
		ServerID                 string `json:"server_id"`
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
