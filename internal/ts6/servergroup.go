package ts6

import (
	"fmt"
	"strings"
	"ts6-viewer/internal/config"
)

// ServerGroup is a servergroup as returned by "servergrouplist".
type ServerGroup struct {
	ID   string
	Name string
}

// GetServerGroups retrieves all servergroups using ServerQuery (SSH).
func GetServerGroups(cfg *config.Config, ssh *SSHClient) ([]ServerGroup, error) {

	raw, err := ssh.Exec("servergrouplist")
	if err != nil {
		return nil, fmt.Errorf("failed to execute servergrouplist: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(raw), "|")
	groups := make([]ServerGroup, 0, len(parts))

	for _, p := range parts {

		fields := strings.Fields(p)
		group := ServerGroup{}

		for _, f := range fields {
			if !strings.Contains(f, "=") {
				continue
			}

			kv := strings.SplitN(f, "=", 2)
			key := kv[0]
			val := UnescapeTS6(kv[1])

			switch key {
			case "sgid":
				group.ID = val
			case "name":
				group.Name = val
			}
		}

		if group.ID != "" {
			groups = append(groups, group)
		}
	}

	return groups, nil
}
