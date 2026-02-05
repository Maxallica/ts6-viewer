package ts6

import (
	"fmt"
)

// Channel represents a TeamSpeak channel
type Channel struct {
	CID               string `json:"cid"`
	PID               string `json:"pid"`
	ChannelOrder      string `json:"channel_order"`
	Name              string `json:"channel_name"`
	Topic             string `json:"channel_topic"`
	FlagPermanent     string `json:"channel_flag_permanent"`
	FlagSemiPermanent string `json:"channel_flag_semi_permanent"`
	FlagDefault       string `json:"channel_flag_default"`
	FlagPassword      string `json:"channel_flag_password"`
	MaxClients        string `json:"channel_maxclients"`
	MaxFamilyClients  string `json:"channel_maxfamilyclients"`
	NeededTalkPower   string `json:"channel_needed_talk_power"`
}

// ChannelListResponse represents the TS6 API response
type ChannelListResponse struct {
	Body   []Channel `json:"body"`
	Status Status    `json:"status"`
}

// GetChannelList returns all channels for a virtual server
func GetChannelList(baseURL, apiKey string, serverID string) ([]Channel, error) {
	var resp ChannelListResponse

	err := doGET(
		baseURL,
		apiKey,
		fmt.Sprintf("/%s/channellist", serverID),
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if resp.Status.Code != 0 {
		return nil, fmt.Errorf("ts6 error %d: %s", resp.Status.Code, resp.Status.Message)
	}

	return resp.Body, nil
}
