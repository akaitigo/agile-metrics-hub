package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Security + CORS middleware
	handler := securityMiddleware(mux, corsOrigin)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Graceful shutdown: SIGTERM/SIGINT を受けたら in-flight リクエストを完了してから停止
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Agile Metrics Hub API listening on :%s", port)
		errCh <- server.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-quit:
		log.Printf("received signal %v, shutting down gracefully...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("graceful shutdown failed: %v", err)
		}
		log.Println("server stopped")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}
}

func securityMiddleware(next http.Handler, allowOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セキュリティヘッダー
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-XSS-Protection", "0")

		// CORS: Origin検証してからヘッダーを設定
		origin := r.Header.Get("Origin")
		if origin == allowOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
