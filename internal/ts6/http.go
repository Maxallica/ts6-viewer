package ts6

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// doGET performs an authenticated HTTP GET request against TS6
func doGET(baseURL, apiKey, path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
