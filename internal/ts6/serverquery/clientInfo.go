package serverquery

import (
	"fmt"
	"strings"
)

type ClientInfo struct {
	CID             string
	CLID            string
	DatabaseID      string
	Nickname        string
	ClientType      string
	InputMuted      string
	InputHardware   string
	OutputMuted     string
	OutputOnlyMuted string
	IsTalker        string
}

func GetClientInfo(ssh *SSHClient, clid string) (*ClientInfo, error) {
	raw, err := ssh.exec(fmt.Sprintf("clientinfo clid=%s", clid))
	if err != nil {
		return nil, fmt.Errorf("failed to execute clientinfo: %w", err)
	}

	// erste Zeile: die eigentlichen Daten
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response from clientinfo")
	}
	dataLine := strings.TrimSpace(lines[0])
	if dataLine == "" {
		return nil, fmt.Errorf("no data line in clientinfo response")
	}

	fields := strings.Fields(dataLine)
	info := &ClientInfo{}

	for _, f := range fields {
		if !strings.Contains(f, "=") {
			continue
		}

		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		val := UnescapeTS6(kv[1])

		switch key {
		case "cid":
			info.CID = val
		case "clid":
			info.CLID = val
		case "client_database_id":
			info.DatabaseID = val
		case "client_nickname":
			info.Nickname = val
		case "client_type":
			info.ClientType = val
		case "client_input_muted":
			info.InputMuted = val
		case "client_input_hardware":
			info.InputHardware = val
		case "client_output_muted":
			info.OutputMuted = val
		case "client_outputonly_muted":
			info.OutputOnlyMuted = val
		case "client_is_talker":
			info.IsTalker = val
		}
	}

	return info, nil
}
