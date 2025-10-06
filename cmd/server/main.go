package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/mylxsw/n8n-parallels/internal/config"
	"github.com/mylxsw/n8n-parallels/internal/handler"
	"github.com/mylxsw/n8n-parallels/internal/logger"
	"github.com/mylxsw/n8n-parallels/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logger)
	
	log.Info("Starting N8n Parallels Server",
		"version", "1.0.0",
		"port", cfg.Server.Port,
		"host", cfg.Server.Host,
		"log_level", cfg.Logger.Level,
		"log_format", cfg.Logger.Format)

	// Initialize services
	webhookService := service.NewWebhookService(log)
	
	// Initialize handlers
	parallelHandler := handler.NewParallelHandler(webhookService, log)

	// Setup routes
	router := mux.NewRouter()
	
	// API routes
	apiRouter := router.PathPrefix("/v1").Subrouter()
	apiRouter.HandleFunc("/parallels/execute", parallelHandler.Execute).Methods("POST")
	
	// Health check endpoint
	router.HandleFunc("/health", parallelHandler.Health).Methods("GET")
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/health", http.StatusFound)
	}).Methods("GET")

	// Add logging middleware
	router.Use(parallelHandler.LoggingMiddleware)

	// Add CORS middleware for cross-origin requests
	router.Use(corsMiddleware)

	// Setup HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("HTTP server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("N8n Parallels Server started successfully", "addr", server.Addr)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}