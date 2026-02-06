package serverquery

import (
	"fmt"
	"strings"
)

type Client struct {
	CID        string
	CLID       string
	DatabaseID string
	Nickname   string
	ClientType string
}

// GetClientList returns all connected clients for a virtual server (SSH)
func GetClientList(c *SSHClient, serverID string) ([]Client, error) {
	// Select the virtual server
	if err := c.Use(serverID); err != nil {
		return nil, err
	}

	raw, err := c.exec("clientlist -uid -away -voice -groups -info")
	if err != nil {
		return nil, fmt.Errorf("failed to execute clientlist: %w", err)
	}

	parts := strings.Split(raw, "|")
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
			case "cid":
				cl.CID = val
			case "clid":
				cl.CLID = val
			case "client_database_id":
				cl.DatabaseID = val
			case "client_nickname":
				cl.Nickname = val
			case "client_type":
				cl.ClientType = val
			}
		}

		// Filter out Query Client (database_id = 1)
		if cl.DatabaseID != "1" {
			clients = append(clients, cl)
		}
	}

	return clients, nil
}
