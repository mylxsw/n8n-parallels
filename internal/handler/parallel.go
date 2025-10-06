package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/mylxsw/n8n-parallels/internal/models"
	"github.com/mylxsw/n8n-parallels/internal/service"
)

// ParallelHandler handles parallel execution requests
type ParallelHandler struct {
	webhookService *service.WebhookService
	validator      *validator.Validate
	logger         *slog.Logger
}

// NewParallelHandler creates a new parallel handler instance
func NewParallelHandler(webhookService *service.WebhookService, logger *slog.Logger) *ParallelHandler {
	return &ParallelHandler{
		webhookService: webhookService,
		validator:      validator.New(),
		logger:         logger,
	}
}

// Execute handles the /v1/parallels/execute endpoint
func (ph *ParallelHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Validate HTTP method
	if r.Method != http.MethodPost {
		ph.sendErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", "only POST method is supported")
		return
	}

	// Parse request body
	var request models.ParallelExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ph.logger.Error("Failed to decode request body", "error", err)
		ph.sendErrorResponse(w, http.StatusBadRequest, "invalid request body", "failed to parse JSON payload")
		return
	}

	// Set default timeout if not provided
	if request.Timeout == 0 {
		request.Timeout = 60 // Default 60 seconds
	}

	// Validate request
	if err := ph.validator.Struct(&request); err != nil {
		ph.logger.Error("Request validation failed", "error", err)
		ph.sendErrorResponse(w, http.StatusBadRequest, "validation failed", err.Error())
		return
	}

	// Additional validation for payloads
	if len(request.Payloads) == 0 {
		ph.sendErrorResponse(w, http.StatusBadRequest, "validation failed", "payloads array cannot be empty")
		return
	}

	// Log the incoming request
	ph.logger.Info("Received parallel execution request",
		"webhook_url", request.WebhookURL,
		"payloads_count", len(request.Payloads),
		"timeout", request.Timeout,
		"has_auth", request.AuthHeader != "",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"))

	// Create context for the request with a slightly longer timeout to allow cleanup
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(request.Timeout+5)*time.Second)
	defer cancel()

	// Execute parallel webhooks
	response := ph.webhookService.ExecuteParallel(ctx, &request)

	// Set appropriate status code based on results
	statusCode := http.StatusOK
	if response.Summary.SuccessfulRequests == 0 {
		statusCode = http.StatusMultiStatus // 207 when all requests failed but the operation itself succeeded
	}

	// Send response
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ph.logger.Error("Failed to encode response", "error", err)
		// At this point, headers are already sent, so we can't change the status code
		return
	}

	ph.logger.Info("Completed parallel execution request",
		"total_requests", response.Summary.TotalRequests,
		"successful_requests", response.Summary.SuccessfulRequests,
		"failed_requests", response.Summary.FailedRequests,
		"timeout_requests", response.Summary.TimeoutRequests,
		"duration_ms", response.Summary.TotalDuration,
		"status_code", statusCode)
}

// Health handles the health check endpoint
func (ph *ParallelHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		ph.sendErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", "only GET method is supported")
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "n8n-parallels",
		"version":   "1.0.0",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse sends a JSON error response
func (ph *ParallelHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, error string, message string) {
	w.WriteHeader(statusCode)
	
	errorResponse := models.ErrorResponse{
		Error:   error,
		Message: message,
	}
	
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		ph.logger.Error("Failed to encode error response", "error", err)
	}
}

// Middleware for logging requests
func (ph *ParallelHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a wrapped response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		
		ph.logger.Info("HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.Header.Get("User-Agent"))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}