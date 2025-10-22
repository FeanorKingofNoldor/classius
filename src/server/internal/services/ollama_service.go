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

// OllamaService implements the AIService interface for Ollama local models
// Ollama provides a simple API for running LLMs locally
type OllamaService struct {
	baseURL     string
	model       string
	client      *http.Client
	maxTokens   int
	temperature float64
}

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Options OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents options for Ollama requests
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"` // Max tokens to generate
	TopK        int     `json:"top_k,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

// NewOllamaService creates a new Ollama service
func NewOllamaService(config map[string]interface{}) (AIService, error) {
	baseURL, ok := config["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama port
	}

	model, ok := config["model"].(string)
	if !ok || model == "" {
		model = "llama3:8b" // Default Ollama model
	}

	maxTokens, ok := config["max_tokens"].(int)
	if !ok {
		maxTokens = 2048
	}

	temperature, ok := config["temperature"].(float64)
	if !ok {
		temperature = 0.7
	}

	return &OllamaService{
		baseURL:     baseURL,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		client: &http.Client{
			Timeout: 120 * time.Second, // Ollama can be slow on first load
		},
	}, nil
}

// Ask sends a request to Ollama and returns the response
func (o *OllamaService) Ask(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	// Build the full prompt with context
	fullPrompt := ""
	if req.Context != "" {
		fullPrompt += req.Context + "\n\n"
	}
	if req.PassageText != "" {
		fullPrompt += fmt.Sprintf("Passage: %s\n\n", req.PassageText)
	}
	fullPrompt += req.Question

	temperature := req.Temperature
	if temperature == 0 {
		temperature = o.temperature
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = o.maxTokens
	}

	ollamaReq := OllamaRequest{
		Model:  o.model,
		Prompt: fullPrompt,
		Stream: false,
		Options: OllamaOptions{
			Temperature: temperature,
			NumPredict:  maxTokens,
			TopK:        40,
			TopP:        0.9,
		},
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	startTime := time.Now()
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	responseTime := time.Since(startTime)

	return &AIResponse{
		Answer:       ollamaResp.Response,
		ResponseTime: responseTime,
		Model:        ollamaResp.Model,
		Provider:     ProviderOllama,
		TokensUsed:   ollamaResp.EvalCount + ollamaResp.PromptEvalCount,
	}, nil
}

// GetCapabilities returns the capabilities of the Ollama service
func (o *OllamaService) GetCapabilities() AICapabilities {
	return AICapabilities{
		SupportsStreaming:     true,
		SupportsConversations: true,
		SupportedLanguages:    []string{"en"}, // Depends on model
		MaxTokens:             o.maxTokens,
		MaxContextLength:      getOllamaModelContextLength(o.model),
		SupportsDocuments:     true,
		SupportsImages:        false, // Most text models don't support images
	}
}

// GetProviderInfo returns information about the Ollama provider
func (o *OllamaService) GetProviderInfo() AIProviderInfo {
	return AIProviderInfo{
		Name:        "Ollama",
		Provider:    ProviderOllama,
		Model:       o.model,
		Version:     "1.0",
		Description: "Local LLM runner with easy model management",
		IsLocal:     true,
		Cost:        "Free (local compute only)",
	}
}

// IsHealthy checks if the Ollama service is available
func (o *OllamaService) IsHealthy(ctx context.Context) error {
	// Check if Ollama is running
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama is not running or not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// getOllamaModelContextLength returns estimated context length for Ollama models
func getOllamaModelContextLength(model string) int {
	// Map common Ollama model names to context lengths
	modelContextMap := map[string]int{
		"llama3:8b":          8192,
		"llama3:70b":         8192,
		"llama3.1:8b":        131072, // Llama 3.1 has much larger context
		"llama3.1:70b":       131072,
		"llama3.2:3b":        131072,
		"mistral:7b":         8192,
		"mixtral:8x7b":       32768,
		"codellama:13b":      16384,
		"phi3:3.8b":          4096,
		"qwen2:7b":           32768,
		"gemma2:9b":          8192,
		"neural-chat:7b":     4096,
	}

	if contextLength, exists := modelContextMap[model]; exists {
		return contextLength
	}

	// Default conservative estimate for unknown models
	return 4096
}