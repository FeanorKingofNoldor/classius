package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LocalLLMService implements the AIService interface for local LLM models
// This can work with OpenAI-compatible APIs (like vLLM, FastChat, etc.)
type LocalLLMService struct {
	baseURL     string
	model       string
	client      *http.Client
	maxTokens   int
	temperature float64
	apiKey      string // Optional for some local setups
}

// LocalLLMRequest represents the request structure for local LLM API
// Uses OpenAI-compatible format for maximum compatibility
type LocalLLMRequest struct {
	Model       string                 `json:"model"`
	Messages    []LocalLLMMessage      `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream"`
	Stop        []string               `json:"stop,omitempty"`
}

// LocalLLMMessage represents a message in the local LLM chat format
type LocalLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LocalLLMResponse represents the response from local LLM API
type LocalLLMResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []LocalLLMChoice         `json:"choices"`
	Usage   *LocalLLMUsage           `json:"usage,omitempty"`
	Error   *LocalLLMError           `json:"error,omitempty"`
}

// LocalLLMChoice represents a choice in the local LLM response
type LocalLLMChoice struct {
	Index        int                      `json:"index"`
	Message      LocalLLMMessage          `json:"message"`
	FinishReason string                   `json:"finish_reason"`
}

// LocalLLMUsage represents token usage information
type LocalLLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LocalLLMError represents an error from local LLM API
type LocalLLMError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewLocalLLMService creates a new local LLM service
func NewLocalLLMService(config map[string]interface{}) (AIService, error) {
	baseURL, ok := config["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = "http://localhost:8000" // Default local LLM server
	}

	model, ok := config["model"].(string)
	if !ok || model == "" {
		model = "classius-sage-7b" // Default local model name
	}

	maxTokens, ok := config["max_tokens"].(int)
	if !ok {
		maxTokens = 2048 // Conservative default for local models
	}

	temperature, ok := config["temperature"].(float64)
	if !ok {
		temperature = 0.7
	}

	// API key is optional for local models
	apiKey, _ := config["api_key"].(string)

	return &LocalLLMService{
		baseURL:     baseURL,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		apiKey:      apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for local models
		},
	}, nil
}

// Ask sends a request to the local LLM and returns the response
func (l *LocalLLMService) Ask(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	messages := []LocalLLMMessage{
		{
			Role:    "system",
			Content: req.Context,
		},
		{
			Role:    "user",
			Content: req.Question,
		},
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = l.temperature
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = l.maxTokens
	}

	localReq := LocalLLMRequest{
		Model:       l.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
		Stop:        []string{"<|endoftext|>", "<|end|>", "</s>"}, // Common stop tokens
	}

	jsonData, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Try different endpoint formats for compatibility
	endpoints := []string{
		l.baseURL + "/v1/chat/completions", // OpenAI-compatible
		l.baseURL + "/chat/completions",    // Alternative
		l.baseURL + "/generate",            // Direct generation endpoint
	}

	var lastError error
	for _, endpoint := range endpoints {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			lastError = fmt.Errorf("failed to create request for %s: %w", endpoint, err)
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if l.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+l.apiKey)
		}

		startTime := time.Now()
		resp, err := l.client.Do(httpReq)
		if err != nil {
			lastError = fmt.Errorf("failed to send request to %s: %w", endpoint, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastError = fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, endpoint, string(body))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastError = fmt.Errorf("failed to read response from %s: %w", endpoint, err)
			continue
		}

		var localResp LocalLLMResponse
		if err := json.Unmarshal(body, &localResp); err != nil {
			lastError = fmt.Errorf("failed to unmarshal response from %s: %w", endpoint, err)
			continue
		}

		if localResp.Error != nil {
			lastError = fmt.Errorf("local llm api error from %s: %s", endpoint, localResp.Error.Message)
			continue
		}

		if len(localResp.Choices) == 0 {
			lastError = fmt.Errorf("no choices in response from %s", endpoint)
			continue
		}

		responseTime := time.Since(startTime)

		// Build response
		aiResp := &AIResponse{
			Answer:       localResp.Choices[0].Message.Content,
			ResponseTime: responseTime,
			Model:        localResp.Model,
			Provider:     ProviderLocal,
		}

		// Add token usage if available
		if localResp.Usage != nil {
			aiResp.TokensUsed = localResp.Usage.TotalTokens
		}

		return aiResp, nil
	}

	return nil, fmt.Errorf("all endpoints failed, last error: %w", lastError)
}

// GetCapabilities returns the capabilities of the local LLM
func (l *LocalLLMService) GetCapabilities() AICapabilities {
	return AICapabilities{
		SupportsStreaming:     false, // Could be enabled if the local model supports it
		SupportsConversations: true,
		SupportedLanguages:    []string{"en"}, // Depends on the local model training
		MaxTokens:             l.maxTokens,
		MaxContextLength:      getLocalModelContextLength(l.model),
		SupportsDocuments:     true,
		SupportsImages:        false, // Most local text models don't support images yet
	}
}

// GetProviderInfo returns information about the local LLM
func (l *LocalLLMService) GetProviderInfo() AIProviderInfo {
	return AIProviderInfo{
		Name:        "Local LLM",
		Provider:    ProviderLocal,
		Model:       l.model,
		Version:     "1.0",
		Description: "Self-hosted LLM optimized for classical education",
		IsLocal:     true,
		Cost:        "Free (local compute only)",
	}
}

// IsHealthy checks if the local LLM service is available
func (l *LocalLLMService) IsHealthy(ctx context.Context) error {
	// Try to ping the health endpoint first
	healthEndpoints := []string{
		l.baseURL + "/health",
		l.baseURL + "/v1/health",
		l.baseURL + "/ping",
	}

	for _, endpoint := range healthEndpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := l.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil // Health endpoint is available
		}
	}

	// If no health endpoint, try a simple generation request
	healthReq := &AIRequest{
		Question:  "Hello",
		Context:   "You are a helpful assistant. Respond with just 'Hi'.",
		MaxTokens: 5,
	}

	_, err := l.Ask(ctx, healthReq)
	return err
}

// getLocalModelContextLength returns estimated context length for local models
func getLocalModelContextLength(model string) int {
	// These are estimates - you'll need to adjust based on your actual model
	if contains(model, "7b") || contains(model, "7B") {
		return 4096
	}
	if contains(model, "13b") || contains(model, "13B") {
		return 4096
	}
	if contains(model, "30b") || contains(model, "30B") {
		return 2048
	}
	if contains(model, "70b") || contains(model, "70B") {
		return 4096
	}

	// Default conservative estimate
	return 2048
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			(len(s) > len(substr) && 
			 (s[:len(substr)] == substr || 
			  s[len(s)-len(substr):] == substr ||
			  containsSubstring(s, substr))))
}

// containsSubstring checks if string contains substring anywhere
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}