package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	// Initialiser le logger en premier
	InitLogger()

	if err := run(); err != nil {
		logger.Error("Application failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() (err error) {
	// Initialiser le cache des projets
	InitProjectCache()

	c := cron.New()

	// Add task for daily messages
	_, err = c.AddFunc(os.Getenv("TRACKER_SLACK_CRON_MESSAGE"), listEventToday)
	if err != nil {
		logger.Error("Error adding scheduled task", slog.Any("error", err))
		return err
	}

	// Add task for cache refresh (every hour)
	_, err = c.AddFunc("0 * * * *", func() {
		logger.Info("Refreshing project cache")
		if err := RefreshProjectCache(); err != nil {
			logger.Error("Error refreshing project cache", slog.Any("error", err))
		} else {
			logger.Info("Project cache refreshed successfully")
		}
	})
	if err != nil {
		logger.Warn("Could not schedule cache refresh task", slog.Any("error", err))
	}

	// Start Task
	c.Start()
	logger.Info("Task planner started")

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
		logger.Info("HTTP server starting", slog.String("address", "localhost:8080"))
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		logger.Error("HTTP server error", slog.Any("error", err))
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		logger.Info("Shutdown signal received")
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	logger.Info("Shutting down HTTP server")
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Register handlers.
	mux.HandleFunc("/slack/command", handleCommand)
	mux.HandleFunc("/slack/interactive_api_endpoint", handleInteractiveAPIEndpoint)
	mux.HandleFunc("/slack/option-load-endpoint", handleOptionLoadEndpoint)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/cache/status", handleCacheStatus)

	handler := http.Handler(mux)
	return handler

}

// handleHealth endpoint de santé
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}

// handleCacheStatus endpoint pour vérifier le statut du cache
func handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	stats := GetCacheStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response, err := json.Marshal(stats)
	if err != nil {
		log.Printf("Error marshaling cache stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(response); err != nil {
		log.Printf("Error writing cache status response: %v", err)
	}
}
