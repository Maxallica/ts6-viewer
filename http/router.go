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
	"ts6-viewer/internal/view"
)

var (
	cacheData      view.ViewerData
	cacheTimestamp time.Time
	cacheTTL       = 5 * time.Second

	lastRequestTime = make(map[string]time.Time)
	rateLimitWindow = 1 * time.Second

	mu sync.Mutex
)

func NewRouter(cfg config.Config) http.Handler {
	if ttl, err := time.ParseDuration(cfg.RefreshInterval + "s"); err == nil {
		cacheTTL = ttl
	}

	baseURL := cfg.Teamspeak6.BaseURL
	apiKey := cfg.Teamspeak6.ApiKey
	serverID := cfg.Teamspeak6.ServerID
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

		var data view.ViewerData
		var err error

		if allowRequest(ip) {
			data, err = getViewerData(&cfg, baseURL, apiKey, serverID)
		} else {
			data = cacheData
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}

	mux.HandleFunc("/ts6viewer/data", dataHandler)
	mux.HandleFunc("/ts6viewer/data/", dataHandler)

	// -----------------------------
	// HTML view (/ts6viewer + /ts6viewer/)
	// -----------------------------
	viewHandler := func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		var data view.ViewerData
		var err error

		if allowRequest(ip) {
			data, err = getViewerData(&cfg, baseURL, apiKey, serverID)
		} else {
			data = cacheData
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Println("Template error:", err)
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
