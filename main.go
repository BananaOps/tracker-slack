package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {

	// Initialiser le cache des projets
	InitProjectCache()

	c := cron.New()

	// Add task for daily messages
	_, err = c.AddFunc(os.Getenv("TRACKER_SLACK_CRON_MESSAGE"), listEventToday)
	if err != nil {
		log.Fatalf("Error adding scheduled task : %v", err)
	}

	// Add task for cache refresh (every hour)
	_, err = c.AddFunc("0 * * * *", func() {
		log.Println("Refreshing project cache...")
		if err := RefreshProjectCache(); err != nil {
			log.Printf("Error refreshing project cache: %v", err)
		} else {
			log.Println("Project cache refreshed successfully")
		}
	})
	if err != nil {
		log.Printf("Warning: Could not schedule cache refresh task: %v", err)
	}

	// Start Task
	c.Start()
	log.Println("task planner started")

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		fmt.Println("server start on localhost:8080")
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Register handlers.
	mux.HandleFunc("/slack/command", handleCommand)
	mux.HandleFunc("/slack/interactive_api_endpoint", handleInteractiveAPIEndpoint)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/cache/status", handleCacheStatus)

	handler := http.Handler(mux)
	return handler

}

// handleHealth endpoint de santé
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleCacheStatus endpoint pour vérifier le statut du cache
func handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	stats := GetCacheStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, _ := json.Marshal(stats)
	w.Write(response)
}
