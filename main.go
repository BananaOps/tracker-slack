package main

import (
	"context"
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

	c := cron.New()

	// Ajouter une tâche
	_, err = c.AddFunc(os.Getenv("TRACKER_SLACK_CRON_MESSAGE"), listEventToday)
	if err != nil {
		log.Fatalf("Error adding scheduled task : %v", err)
	}

	// Démarrer le planificateur
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

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	//handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	//	// Configure the "http.route" for the HTTP instrumentation.
	//	//handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
	//	mux.Handle(pattern, http.HandlerFunc(handlerFunc))
	//}

	// Register handlers.
	mux.HandleFunc("/slack/command", handleCommand)
	mux.HandleFunc("/slack/interactive_api_endpoint", handleInteractiveAPIEndpoint)

	handler := http.Handler(mux)
	return handler

}
