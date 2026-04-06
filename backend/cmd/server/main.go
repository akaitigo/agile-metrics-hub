package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akaitigo/agile-metrics-hub/internal/adapter"
	"github.com/akaitigo/agile-metrics-hub/internal/adapter/clickup"
	"github.com/akaitigo/agile-metrics-hub/internal/adapter/jira"
	"github.com/akaitigo/agile-metrics-hub/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	corsOrigin := os.Getenv("CORS_ORIGINS")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	// Adapter Registry
	registry := adapter.NewRegistry()
	registry.Register("clickup", clickup.NewAdapter)
	registry.Register("jira", jira.NewAdapter)

	// Handlers
	connHandler := &handler.ConnectionsHandler{Registry: registry}
	metricsHandler := &handler.MetricsHandler{}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, `{"status":"ok"}`); err != nil {
			log.Printf("health check write failed: %v", err)
		}
	})

	// Connections
	mux.HandleFunc("POST /api/connections/test", connHandler.TestConnection)

	// Metrics
	mux.HandleFunc("GET /api/metrics/burndown", metricsHandler.Burndown)
	mux.HandleFunc("GET /api/metrics/velocity", metricsHandler.Velocity)
	mux.HandleFunc("GET /api/metrics/cumulative-flow", metricsHandler.CumulativeFlow)
	mux.HandleFunc("GET /api/metrics/lead-time", metricsHandler.LeadTime)

	// CORS middleware
	corsHandler := corsMiddleware(mux, corsOrigin)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           corsHandler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	log.Printf("Agile Metrics Hub API listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler, allowOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
