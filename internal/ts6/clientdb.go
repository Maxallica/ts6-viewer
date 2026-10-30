package ts6

import (
	"fmt"
	"strconv"
	"strings"

	"ts6-viewer/internal/config"
)

// ClientDB is one entry from the server's client database (clientdblist).
// Unlike the live client list it also covers clients that are currently
// offline, with the time they last disconnected.
type ClientDB struct {
	UID            string
	Nickname       string
	LastNickname   string
	Country        string
	LastSeenUnix   int64 // unix seconds of the last disconnect (0 = unknown)
	NumSeen        int64 // total number of connections (0 = unknown)
}

// GetClientDB lists the server's client database via SSH ServerQuery.
// The server keeps an entry for every client that has ever connected, so
// this catches even one-second connections that fall between two polls of
// the live client list.
//
// Verified live against the production server (and the ServerQuery
// manual): the command is "clientdblist" (one word) and the last-seen
// timestamp is "client_lastconnected"; the connection count is
// "client_totalconnections". Entries are separated by "|".
func GetClientDB(cfg *config.Config, ssh *SSHClient) ([]ClientDB, error) {

	raw, err := ssh.Exec("clientdblist")
	if err != nil {
		return nil, fmt.Errorf("failed to execute clientdblist: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(raw), "|")
	entries := make([]ClientDB, 0, len(parts))

	for _, p := range parts {

		fields := strings.Fields(p)
		cl := ClientDB{}

		for _, f := range fields {
			if !strings.Contains(f, "=") {
				continue
			}

			kv := strings.SplitN(f, "=", 2)
			key := kv[0]
			val := UnescapeTS6(kv[1])

			switch key {
			case "client_unique_identifier":
				cl.UID = val
			case "client_nickname":
				cl.Nickname = val
			case "client_last_nickname":
				cl.LastNickname = val
			case "client_country":
				cl.Country = val
			case "client_totalconnections", "client_num_seen":
				cl.NumSeen, _ = strconv.ParseInt(val, 10, 64)
			case "client_lastconnected", "client_last_seen":
				cl.LastSeenUnix, _ = strconv.ParseInt(val, 10, 64)
			}
		}

		// Skip entries without a usable unique identifier or timestamp.
		if cl.UID == "" || cl.LastSeenUnix <= 0 {
			continue
		}

		entries = append(entries, cl)
	}

	return entries, nil
}
