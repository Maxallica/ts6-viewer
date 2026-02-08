package webquery

import (
	"encoding/json"
	"fmt"
)

type ClientInfo struct {
	CID        string `json:"cid"`
	CLID       string `json:"clid"`
	DatabaseID string `json:"client_database_id"`
	Nickname   string `json:"client_nickname"`
	ClientType string `json:"client_type"`

	// Voice / Mute
	InputMuted      string `json:"client_input_muted"`
	InputHardware   string `json:"client_input_hardware"`
	OutputMuted     string `json:"client_output_muted"`
	OutputOnlyMuted string `json:"client_outputonly_muted"`
	IsTalker        string `json:"client_is_talker"`

	// Extra fields you posted
	IsRecording      string `json:"client_is_recording"`
	IsCommander      string `json:"client_is_channel_commander"`
	NicknamePhonetic string `json:"client_nickname_phonetic"`
	TalkPower        string `json:"client_talk_power"`
	UniqueID         string `json:"client_unique_identifier"`
	Base64HashUID    string `json:"client_base64HashClientUID"`
	ChannelGroupID   string `json:"client_channel_group_id"`
	IdleTime         string `json:"client_idle_time"`

	// Connection
	ConnectedTime           string `json:"connection_connected_time"`
	BandwidthSentLastMinute string `json:"connection_bandwidth_sent_last_minute_total"`
}

type ClientInfoResponse struct {
	Body   []ClientInfo `json:"body"`
	Status Status       `json:"status"`
}

func GetClientInfo(baseURL, apiKey, serverID, clid string) (*ClientInfo, error) {
	var raw map[string]any

	err := doGET(
		baseURL,
		apiKey,
		fmt.Sprintf("/%s/clientinfo?clid=%s", serverID, clid),
		&raw,
	)
	if err != nil {
		return nil, err
	}

	var resp ClientInfoResponse
	b, _ := json.Marshal(raw)
	json.Unmarshal(b, &resp)

	if resp.Status.Code != 0 {
		return nil, fmt.Errorf(
			"ts6 error %d: %s",
			resp.Status.Code,
			resp.Status.Message,
		)
	}

	return &resp.Body[0], nil
}
