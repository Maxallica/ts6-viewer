package ts6

import "fmt"

// Client represents an online TeamSpeak client
type Client struct {
	CID        string `json:"cid"`
	CLID       string `json:"clid"`
	DatabaseID string `json:"client_database_id"`
	Nickname   string `json:"client_nickname"`
	ClientType string `json:"client_type"`
}

// ClientListResponse represents the TS6 API response
type ClientListResponse struct {
	Body   []Client `json:"body"`
	Status Status   `json:"status"`
}

// GetClientList returns all connected clients for a virtual server
func GetClientList(baseURL, apiKey string, serverID string) ([]Client, error) {
	var resp ClientListResponse

	err := doGET(
		baseURL,
		apiKey,
		fmt.Sprintf("/%s/clientlist", serverID),
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

	filtered := make([]Client, 0, len(resp.Body))
	for _, c := range resp.Body {
		if c.DatabaseID != "1" {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}
