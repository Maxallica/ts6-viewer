package main

import (
	"log"
	"net"
	"net/http"

	router "ts6-viewer/http"
	"ts6-viewer/internal/config"
)

func main() {
	log.Println("Starting TS6 Viewer...")

	// Load config.json
	log.Println("Loading config.json")
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal(err)
	}
	serverPort := cfg.ServerPort
	log.Println("Successfully loaded config.json")

	// Create HTTP router
	r := router.NewRouter(*cfg)

	// Create listener first
	ln, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatal("Failed to bind port:", err)
	}

	// Now we are guaranteed the port is bound → safe callback
	log.Printf("HTTP server is now listening on port %s", serverPort)

	// Start the server (blocking)
	srv := &http.Server{
		Handler: r,
	}

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatal("HTTP server failed:", err)
	}
}
