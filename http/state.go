package http

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// stateFile holds the persisted statistics state (recent-activity log and
// online-count history). By default it lives in the working directory,
// same convention as config.json, and is replaced atomically (write temp
// file, then rename) so a crash mid-write never corrupts it. The
// STATE_FILE environment variable can override the location (Docker
// mounts a volume there so the state survives container recreation).
var stateFile = "state.json"

func init() {
	if p := os.Getenv("STATE_FILE"); p != "" {
		stateFile = p
	}
}

// persistedEntry is the JSON form of one activityEntry.
type persistedEntry struct {
	Nickname        string `json:"nickname"`
	Badge           string `json:"badge"`
	Country         string `json:"country"`
	ConnectedAt     string `json:"connectedAt"`
	ConnectedAtUnix int64  `json:"connectedAtUnix"`
	OnlineAt        int64  `json:"onlineAt"`
	OfflineAt       int64  `json:"offlineAt"`
}

// persistedSample is the JSON form of one historySample.
type persistedSample struct {
	T int64 `json:"t"`
	N int   `json:"n"`
}

// persistedState is the on-disk statistics state.
type persistedState struct {
	Activity map[string]persistedEntry `json:"activity"`
	History  []persistedSample         `json:"history"`
}

// loadState restores the persisted statistics state at startup (before the
// first request). A missing file is a normal first start; an unreadable
// file is logged and ignored. Entries and samples that already fell out
// of their configured windows are dropped.
func loadState() {
	raw, err := os.ReadFile(stateFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[HTTP] Cannot read persisted state: %v\n", err)
		}
		return
	}

	var s persistedState
	if err := json.Unmarshal(raw, &s); err != nil {
		log.Printf("[HTTP] Persisted state is unreadable, starting fresh: %v\n", err)
		return
	}

	now := time.Now().Unix()
	activityCutoff := now - int64(activityWindow/time.Second)
	historyCutoff := now - int64(historyWindow/time.Second)

	activityLog = make(map[string]*activityEntry, len(s.Activity))
	for uid, p := range s.Activity {
		last := p.OnlineAt
		if p.OfflineAt > last {
			last = p.OfflineAt
		}
		if last < activityCutoff {
			continue
		}
		activityLog[uid] = &activityEntry{
			nickname:        p.Nickname,
			badge:           p.Badge,
			country:         p.Country,
			connectedAt:     p.ConnectedAt,
			connectedAtUnix: p.ConnectedAtUnix,
			onlineAt:        p.OnlineAt,
			offlineAt:       p.OfflineAt,
		}
	}

	onlineHistory = make([]historySample, 0, len(s.History))
	for _, h := range s.History {
		if h.T >= historyCutoff {
			onlineHistory = append(onlineHistory, historySample{t: h.T, n: h.N})
		}
	}

	log.Printf("[HTTP] Loaded persisted state: %d activity entries, %d history samples\n",
		len(activityLog), len(onlineHistory))
}

// saveState writes the current statistics state to disk. The caller must
// hold mu (getViewerData does, for the whole fetch).
func saveState() {
	s := persistedState{Activity: make(map[string]persistedEntry, len(activityLog))}
	for uid, e := range activityLog {
		s.Activity[uid] = persistedEntry{
			Nickname:        e.nickname,
			Badge:           e.badge,
			Country:         e.country,
			ConnectedAt:     e.connectedAt,
			ConnectedAtUnix: e.connectedAtUnix,
			OnlineAt:        e.onlineAt,
			OfflineAt:       e.offlineAt,
		}
	}

	s.History = make([]persistedSample, 0, len(onlineHistory))
	for _, h := range onlineHistory {
		s.History = append(s.History, persistedSample{T: h.t, N: h.n})
	}

	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Printf("[HTTP] Cannot encode state for persistence: %v\n", err)
		return
	}

	tmp := stateFile + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		log.Printf("[HTTP] Cannot write state file: %v\n", err)
		return
	}
	if err := os.Rename(tmp, stateFile); err != nil {
		log.Printf("[HTTP] Cannot move state file into place: %v\n", err)
		os.Remove(tmp)
	}
}
