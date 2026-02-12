package ts6

import (
	"fmt"
	"strings"
	"ts6-viewer/internal/config"
)

type Channel struct {
	CID                           string
	PID                           string
	ChannelOrder                  string
	Name                          string
	Topic                         string
	FlagPermanent                 string
	FlagSemiPermanent             string
	FlagDefault                   string
	FlagPassword                  string
	FlagMaxClientsUnlimited       string
	FlagMaxFamilyClientsUnlimited string
	MaxClients                    string
	MaxFamilyClients              string
	NeededTalkPower               string
	Codec                         string
	CodecQuality                  string
	TotalClients                  string
	IconID                        string
	SecondsEmpty                  string
}

// GetChannelList retrieves all channels using ServerQuery (SSH)
func GetChannelList(cfg *config.Config, ssh *SSHClient) ([]Channel, error) {

	raw, err := ssh.exec("channellist -topic -flags -limits -voice -icon -secondsempty")
	if err != nil {
		return nil, fmt.Errorf("failed to execute channellist: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(raw), "|")
	channels := make([]Channel, 0, len(parts))

	for _, p := range parts {

		fields := strings.Fields(p)
		ch := Channel{}

		for _, f := range fields {
			if !strings.Contains(f, "=") {
				continue
			}

			kv := strings.SplitN(f, "=", 2)
			key := kv[0]
			val := UnescapeTS6(kv[1])

			switch key {

			case "cid":
				ch.CID = val
			case "pid":
				ch.PID = val
			case "channel_order":
				ch.ChannelOrder = val
			case "channel_name":
				ch.Name = val
			case "channel_topic":
				ch.Topic = val

			case "channel_flag_permanent":
				ch.FlagPermanent = val
			case "channel_flag_semi_permanent":
				ch.FlagSemiPermanent = val
			case "channel_flag_default":
				ch.FlagDefault = val
			case "channel_flag_password":
				ch.FlagPassword = val
			case "channel_flag_maxclients_unlimited":
				ch.FlagMaxClientsUnlimited = val
			case "channel_flag_maxfamilyclients_unlimited":
				ch.FlagMaxFamilyClientsUnlimited = val

			case "channel_maxclients":
				ch.MaxClients = val
			case "channel_maxfamilyclients":
				ch.MaxFamilyClients = val

			case "channel_needed_talk_power":
				ch.NeededTalkPower = val

			case "channel_codec":
				ch.Codec = val
			case "channel_codec_quality":
				ch.CodecQuality = val
			case "total_clients":
				ch.TotalClients = val

			case "channel_icon_id":
				ch.IconID = val

			case "seconds_empty":
				ch.SecondsEmpty = val
			}
		}

		channels = append(channels, ch)
	}

	return channels, nil
}
