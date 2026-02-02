package ts6

import (
	"fmt"
	"strconv"
)

// ServerInfo represents basic virtual server information
type ServerInfo struct {
	ServerID               string `json:"virtualserver_id"`
	Name                   string `json:"virtualserver_name"`
	Uptime                 string `json:"virtualserver_uptime"`
	ClientsOnline          string `json:"virtualserver_clientsonline"`
	MaxClients             string `json:"virtualserver_maxclients"`
	ChannelsOnline         string `json:"virtualserver_channelsonline"`
	HostBannerURL          string `json:"virtualserver_hostbanner_url"`
	HostBannerGfxURL       string `json:"virtualserver_hostbanner_gfx_url"`
	NeededIdentitySecurity string `json:"virtualserver_needed_identity_security_level"`
	QueryClientConnections string `json:"virtualserver_query_client_connections"`
	ClientConnections      string `json:"virtualserver_client_connections"`

	UptimePretty string `json:"-"`
}

// ServerInfoResponse represents the TS6 API response
type ServerInfoResponse struct {
	Body   []ServerInfo `json:"body"`
	Status Status       `json:"status"`
}

// GetServerInfo retrieves information about a virtual server
func GetServerInfo(baseURL, apiKey string, serverID string) (*ServerInfo, error) {
	var resp ServerInfoResponse

	err := doGET(
		baseURL,
		apiKey,
		fmt.Sprintf("/%s/serverinfo", serverID),
		&resp,
	)
	if err != nil {
		return nil, err
	}

	if resp.Status.Code != 0 {
		return nil, fmt.Errorf(
			"ts6 error %d: %s",
			resp.Status.Code,
			resp.Status.Message,
		)
	}

	if len(resp.Body) != 1 {
		return nil, fmt.Errorf(
			"unexpected serverinfo result count: %d",
			len(resp.Body),
		)
	}

	info := resp.Body[0]

	if n, err := strconv.Atoi(info.ClientsOnline); err == nil {
		n--
		if n < 0 {
			n = 0
		}
		info.ClientsOnline = strconv.Itoa(n)
	}

	modifiedInfo := &info
	modifiedInfo.UptimePretty = formatUptime(modifiedInfo.Uptime)

	return modifiedInfo, nil
}

func formatUptime(secondsStr string) string {
	sec, err := strconv.Atoi(secondsStr)
	if err != nil || sec < 0 {
		return secondsStr
	}

	days := sec / 86400
	sec %= 86400
	hours := sec / 3600
	sec %= 3600
	minutes := sec / 60
	seconds := sec % 60

	return fmt.Sprintf("%dD %02d:%02d:%02d", days, hours, minutes, seconds)
}
