package ts6

import (
	"fmt"
	"strings"
	"ts6-viewer/internal/config"
)

type Client struct {
	CLID             string
	CID              string
	DatabaseID       string
	Nickname         string
	Type             string
	UniqueIdentifier string

	Away        string
	AwayMessage string

	InputMuted      string
	OutputMuted     string
	OutputOnlyMuted string
	InputHardware   string
	OutputHardware  string
	TalkPower       string
	IsTalking       string

	ServerGroups   string
	ChannelGroupID string

	IdleTime       string
	ConnectionTime string

	Country string
	IconID  string

	Version  string
	Platform string
}

func GetClientList(cfg *config.Config, ssh *SSHClient) ([]Client, error) {

	voiceCmd := ""
	if cfg.Teamspeak6.EnableVoiceStatus == "true" {
		voiceCmd = "-voice"
	}

	raw, err := ssh.exec("clientlist -uid -away -groups -times -info -country -icon " + voiceCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute clientlist: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(raw), "|")
	clients := make([]Client, 0, len(parts))

	for _, p := range parts {

		fields := strings.Fields(p)
		cl := Client{}

		for _, f := range fields {
			if !strings.Contains(f, "=") {
				continue
			}

			kv := strings.SplitN(f, "=", 2)
			key := kv[0]
			val := UnescapeTS6(kv[1])

			switch key {

			case "clid":
				cl.CLID = val
			case "cid":
				cl.CID = val
			case "client_database_id":
				cl.DatabaseID = val
			case "client_nickname":
				cl.Nickname = val
			case "client_type":
				cl.Type = val
			case "client_unique_identifier":
				cl.UniqueIdentifier = val

			case "client_away":
				cl.Away = val
			case "client_away_message":
				cl.AwayMessage = val

			case "client_input_muted":
				cl.InputMuted = val
			case "client_output_muted":
				cl.OutputMuted = val
			case "client_outputonly_muted":
				cl.OutputOnlyMuted = val
			case "client_input_hardware":
				cl.InputHardware = val
			case "client_output_hardware":
				cl.OutputHardware = val
			case "client_talk_power":
				cl.TalkPower = val
			case "client_is_talking":
				cl.IsTalking = val

			case "client_servergroups":
				cl.ServerGroups = val
			case "client_channel_group_id":
				cl.ChannelGroupID = val

			case "client_idle_time":
				cl.IdleTime = val
			case "client_connection_connected_time":
				cl.ConnectionTime = val

			case "client_country":
				cl.Country = val
			case "client_icon_id":
				cl.IconID = val

			case "client_version":
				cl.Version = val
			case "client_platform":
				cl.Platform = val
			}
		}

		// Query Clients rausfiltern
		if cl.Type == "1" {
			continue
		}

		clients = append(clients, cl)
	}

	return clients, nil
}
