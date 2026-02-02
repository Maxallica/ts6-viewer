package http

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"ts6-viewer/internal/config"
)

var (
	cacheData      ViewerData
	cacheTimestamp time.Time
	cacheTTL       = 3 * time.Second

	lastRequestTime = make(map[string]time.Time)
	rateLimitWindow = 1 * time.Second

	mu sync.Mutex
)

func NewRouter(cfg config.Config) http.Handler {
	baseURL := cfg.Teamspeak6.BaseURL
	apiKey := cfg.Teamspeak6.ApiKey
	serverID := cfg.Teamspeak6.ServerID
	theme := cfg.Theme
	refreshIntervalStr := cfg.RefreshInterval

	refreshInterval, err := strconv.Atoi(refreshIntervalStr)
	if err != nil || refreshInterval <= 0 {
		refreshIntervalStr = "60"
	}

	mux := http.NewServeMux()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("cannot get working directory:", err)
	}

	tmplPath := filepath.Join(wd, "..", "..", "internal", "web", "templates", "ts6viewer.html")
	tmpl := template.Must(template.ParseFiles(tmplPath))

	staticPath := filepath.Join(wd, "..", "..", "internal", "web", "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// -----------------------------
	// JSON endpoint (/ts6viewer/data + /ts6viewer/data/)
	// -----------------------------
	dataHandler := func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		var raw ViewerData
		var err error

		if allowRequest(ip) {
			raw, err = getViewerData(baseURL, apiKey, serverID)
		} else {
			raw = cacheData
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		raw.Theme = theme
		raw.RefreshInterval = refreshIntervalStr

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(raw)
	}

	mux.HandleFunc("/ts6viewer/data", dataHandler)
	mux.HandleFunc("/ts6viewer/data/", dataHandler)

	// -----------------------------
	// HTML view (/ts6viewer + /ts6viewer/)
	// -----------------------------
	viewHandler := func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		var data ViewerData
		var err error

		if allowRequest(ip) {
			data, err = getViewerData(baseURL, apiKey, serverID)
		} else {
			data = cacheData
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data.Theme = theme
		data.RefreshInterval = refreshIntervalStr

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Println("template error:", err)
		}
	}

	mux.HandleFunc("/ts6viewer", viewHandler)
	mux.HandleFunc("/ts6viewer/", viewHandler)

	// -----------------------------
	// Health endpoint
	// -----------------------------
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TS6Viewer is running!"))
	})

	return mux
}
