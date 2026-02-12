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

	Theme           string `json:"theme"`
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
