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
	cacheData      view.VMTS6Viewer
	cacheTimestamp time.Time
	cacheTTL       = 5 * time.Second

	lastRequestTime = make(map[string]time.Time)
	rateLimitWindow = 1 * time.Second

	mu sync.Mutex
)

// NewRouter sets up all HTTP routes and returns the router.
func NewRouter(cfg config.Config) http.Handler {
	// parse refresh interval
	if ttl, err := time.ParseDuration(cfg.RefreshInterval + "s"); err == nil {
		cacheTTL = ttl
		log.Printf("[HTTP] Cache TTL set to %v\n", cacheTTL)
	}

	refreshIntervalStr := cfg.RefreshInterval
	refreshInterval, err := strconv.Atoi(refreshIntervalStr)
	if err != nil || refreshInterval <= 0 {
		refreshIntervalStr = "60"
		refreshInterval = 60
	}

	log.Printf("[HTTP] Refresh interval: %d seconds\n", refreshInterval)

	mux := http.NewServeMux()

	// Load templates
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("[HTTP] Cannot get working directory:", err)
	}

	tmplPath := filepath.Join(wd, "..", "..", "internal", "web", "templates", "ts6viewer.html")
	tmpl := template.Must(template.ParseFiles(tmplPath))
	log.Printf("[HTTP] Loaded template: %s\n", tmplPath)

	// Static assets
	staticPath := filepath.Join(wd, "..", "..", "internal", "web", "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	log.Printf("[HTTP] Static files served from: %s\n", staticPath)

	// -----------------------------
	// JSON data endpoint
	// -----------------------------
	dataHandler := func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		log.Printf("[HTTP] /ts6viewer/data requested from IP: %s\n", ip)

		var data view.VMTS6Viewer
		var err error
		force := r.URL.Query().Get("force") == "1"
		if force {
			log.Printf("[HTTP] Force refresh requested by IP: %s\n", ip)
		}

		if allowRequest(ip) {
			data, err = getViewerData(&cfg, force)
		} else {
			log.Printf("[HTTP] Rate limit hit for IP: %s\n", ip)
			data = cacheData
		}

		if err != nil {
			log.Printf("[HTTP] Error getting viewer data: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("[HTTP] Error encoding JSON response: %v\n", err)
		}
	}

	mux.HandleFunc("/ts6viewer/data", dataHandler)
	mux.HandleFunc("/ts6viewer/data/", dataHandler)

	// -----------------------------
	// HTML view endpoint
	// -----------------------------
	viewHandler := func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		log.Printf("[HTTP] /ts6viewer requested from IP: %s\n", ip)

		var data view.VMTS6Viewer
		var err error
		if allowRequest(ip) {
			data, err = getViewerData(&cfg, true)
		} else {
			log.Printf("[HTTP] Rate limit hit for IP: %s\n", ip)
			data = cacheData
		}

		if err != nil {
			log.Printf("[HTTP] Error getting viewer data: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("[HTTP] Template execution error: %v\n", err)
		}
	}

	mux.HandleFunc("/ts6viewer", viewHandler)
	mux.HandleFunc("/ts6viewer/", viewHandler)

	// -----------------------------
	// Health check
	// -----------------------------
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] /health check from IP: %s\n", getIP(r))
		w.WriteHeader(http.StatusOK)
	})

	// -----------------------------
	// Root endpoint
	// -----------------------------
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] / root requested from IP: %s\n", getIP(r))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TS6Viewer is running!"))
	})

	return mux
}
