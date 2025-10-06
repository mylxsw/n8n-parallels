package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/mylxsw/n8n-parallels/internal/models"
)

// WebhookService handles parallel webhook execution
type WebhookService struct {
	client *http.Client
	logger *slog.Logger
}

// NewWebhookService creates a new webhook service instance
func NewWebhookService(logger *slog.Logger) *WebhookService {
	return &WebhookService{
		client: &http.Client{
			Timeout: 0, // We'll handle timeout per request
		},
		logger: logger,
	}
}

// ExecuteParallel executes webhook requests in parallel and returns results in order
func (ws *WebhookService) ExecuteParallel(ctx context.Context, request *models.ParallelExecuteRequest) *models.ParallelExecuteResponse {
	startTime := time.Now()
	totalRequests := len(request.Payloads)
	
	ws.logger.Info("Starting parallel webhook execution",
		"webhook_url", request.WebhookURL,
		"total_requests", totalRequests,
		"timeout_seconds", request.Timeout)

	// Create tasks
	tasks := make([]models.WebhookExecutionTask, totalRequests)
	for i, payload := range request.Payloads {
		tasks[i] = models.WebhookExecutionTask{
			Index:      i,
			WebhookURL: request.WebhookURL,
			AuthHeader: request.AuthHeader,
			Payload:    payload,
			TimeoutSec: request.Timeout,
		}
	}

	// Execute tasks in parallel
	results := ws.executeTasksParallel(ctx, tasks)

	// Sort results by index to maintain order
	sortedResults := make([]models.WebhookExecutionResult, totalRequests)
	for _, result := range results {
		sortedResults[result.Index] = result
	}

	// Convert to response format and calculate summary
	webhookResults := make([]models.WebhookResult, totalRequests)
	summary := models.ExecutionSummary{
		TotalRequests: totalRequests,
		TotalDuration: time.Since(startTime).Milliseconds(),
	}

	for i, result := range sortedResults {
		webhookResult := models.WebhookResult{
			Index:    i,
			Success:  result.Success,
			Duration: result.Duration,
		}

		if result.Success {
			webhookResult.Response = result.Response
			summary.SuccessfulRequests++
		} else {
			if result.IsTimeout {
				webhookResult.Error = "timeout"
				summary.TimeoutRequests++
			} else if result.Error != nil {
				webhookResult.Error = result.Error.Error()
			} else {
				webhookResult.Error = "unknown error"
			}
			summary.FailedRequests++
		}

		webhookResults[i] = webhookResult
	}

	ws.logger.Info("Completed parallel webhook execution",
		"total_requests", summary.TotalRequests,
		"successful", summary.SuccessfulRequests,
		"failed", summary.FailedRequests,
		"timeout", summary.TimeoutRequests,
		"duration_ms", summary.TotalDuration)

	return &models.ParallelExecuteResponse{
		Results: webhookResults,
		Summary: summary,
	}
}

// executeTasksParallel executes webhook tasks in parallel using goroutines
func (ws *WebhookService) executeTasksParallel(ctx context.Context, tasks []models.WebhookExecutionTask) []models.WebhookExecutionResult {
	var wg sync.WaitGroup
	results := make([]models.WebhookExecutionResult, len(tasks))
	
	// Use buffered channel to prevent goroutine leaks
	resultChan := make(chan models.WebhookExecutionResult, len(tasks))

	// Start goroutines for each task
	for i, task := range tasks {
		wg.Add(1)
		go func(taskIndex int, t models.WebhookExecutionTask) {
			defer wg.Done()
			result := ws.executeTask(ctx, t)
			resultChan <- result
		}(i, task)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	i := 0
	for result := range resultChan {
		results[i] = result
		i++
	}

	return results
}

// executeTask executes a single webhook task
func (ws *WebhookService) executeTask(ctx context.Context, task models.WebhookExecutionTask) models.WebhookExecutionResult {
	startTime := time.Now()
	
	result := models.WebhookExecutionResult{
		Index: task.Index,
	}

	// Create request context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(task.TimeoutSec)*time.Second)
	defer cancel()

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(task.Payload)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal payload: %w", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(taskCtx, "POST", task.WebhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if task.AuthHeader != "" {
		req.Header.Set("Authorization", task.AuthHeader)
	}

	ws.logger.Debug("Executing webhook request",
		"index", task.Index,
		"url", task.WebhookURL,
		"payload_size", len(payloadBytes))

	// Execute request
	resp, err := ws.client.Do(req)
	if err != nil {
		result.Duration = time.Since(startTime).Milliseconds()
		if taskCtx.Err() == context.DeadlineExceeded {
			result.IsTimeout = true
			result.Error = fmt.Errorf("request timeout after %d seconds", task.TimeoutSec)
		} else {
			result.Error = fmt.Errorf("request failed: %w", err)
		}
		return result
	}
	defer resp.Body.Close()

	result.Duration = time.Since(startTime).Milliseconds()

	// Read response body
	var responseBytes bytes.Buffer
	_, err = responseBytes.ReadFrom(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		return result
	}

	// Check if response is successful (2xx status codes)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Response = json.RawMessage(responseBytes.Bytes())
		ws.logger.Debug("Webhook request successful",
			"index", task.Index,
			"status_code", resp.StatusCode,
			"duration_ms", result.Duration)
	} else {
		result.Error = fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, responseBytes.String())
		ws.logger.Debug("Webhook request failed",
			"index", task.Index,
			"status_code", resp.StatusCode,
			"duration_ms", result.Duration)
	}

	return result
}