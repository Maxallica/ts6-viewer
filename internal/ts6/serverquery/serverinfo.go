package serverquery

import (
	"fmt"
	"strconv"
	"strings"
	"ts6-viewer/internal/config"
)

type ServerInfo struct {
	ServerID               string
	Name                   string
	Uptime                 string
	ClientsOnline          string
	MaxClients             string
	ChannelsOnline         string
	HostBannerURL          string
	HostBannerGfxURL       string
	NeededIdentitySecurity string
	QueryClientConnections string
	ClientConnections      string
}

func GetServerInfo(cfg *config.Config, c *SSHClient) (*ServerInfo, error) {
	if err := c.Use(cfg.Teamspeak6.ServerID); err != nil {
		return nil, err
	}

	raw, err := c.exec("serverinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to execute serverinfo: %w", err)
	}

	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty serverinfo response")
	}

	data := strings.TrimSpace(lines[0])
	parts := strings.Split(data, " ")

	info := &ServerInfo{}

	for _, p := range parts {
		if !strings.Contains(p, "=") {
			continue
		}

		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := kv[0]
		val := UnescapeTS6(kv[1])

		switch key {
		case "virtualserver_id":
			info.ServerID = val
		case "virtualserver_name":
			info.Name = val
		case "virtualserver_uptime":
			info.Uptime = val
		case "virtualserver_clientsonline":
			info.ClientsOnline = val
		case "virtualserver_maxclients":
			info.MaxClients = val
		case "virtualserver_channelsonline":
			info.ChannelsOnline = val
		case "virtualserver_hostbanner_url":
			info.HostBannerURL = val
		case "virtualserver_hostbanner_gfx_url":
			info.HostBannerGfxURL = val
		case "virtualserver_needed_identity_security_level":
			info.NeededIdentitySecurity = val
		case "virtualserver_query_client_connections":
			info.QueryClientConnections = val
		case "virtualserver_client_connections":
			info.ClientConnections = val
		}
	}

	if n, err := strconv.Atoi(info.ClientsOnline); err == nil {
		n--
		if n < 0 {
			n = 0
		}
		info.ClientsOnline = strconv.Itoa(n)
	}

	return info, nil
}
