package models

import "encoding/json"

// ParallelExecuteRequest represents the request payload for parallel webhook execution
type ParallelExecuteRequest struct {
	WebhookURL string                   `json:"webhook_url" validate:"required,url"`
	AuthHeader string                   `json:"auth_header"`
	Payloads   []map[string]interface{} `json:"payloads" validate:"required,min=1"`
	Timeout    int                      `json:"timeout" validate:"min=1,max=3600"` // 1 second to 1 hour
}

// ParallelExecuteResponse represents the response for parallel webhook execution
type ParallelExecuteResponse struct {
	Results []WebhookResult `json:"results"`
	Summary ExecutionSummary `json:"summary"`
}

// WebhookResult represents the result of a single webhook call
type WebhookResult struct {
	Index    int             `json:"index"`
	Success  bool            `json:"success"`
	Response json.RawMessage `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
	Duration int64           `json:"duration_ms"` // Duration in milliseconds
}

// ExecutionSummary provides summary statistics of the parallel execution
type ExecutionSummary struct {
	TotalRequests     int   `json:"total_requests"`
	SuccessfulRequests int   `json:"successful_requests"`
	FailedRequests    int   `json:"failed_requests"`
	TimeoutRequests   int   `json:"timeout_requests"`
	TotalDuration     int64 `json:"total_duration_ms"` // Total execution time in milliseconds
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WebhookExecutionTask represents a single webhook execution task
type WebhookExecutionTask struct {
	Index       int
	WebhookURL  string
	AuthHeader  string
	Payload     map[string]interface{}
	TimeoutSec  int
}

// WebhookExecutionResult represents the result of a webhook execution task
type WebhookExecutionResult struct {
	Index     int
	Success   bool
	Response  json.RawMessage
	Error     error
	Duration  int64 // Duration in milliseconds
	IsTimeout bool
}