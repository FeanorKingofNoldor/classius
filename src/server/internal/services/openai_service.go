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

// OpenAIService implements the AIService interface for OpenAI
type OpenAIService struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	maxTokens int
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIMessage        `json:"messages"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream"`
}

// OpenAIMessage represents a message in the OpenAI chat format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIChoice       `json:"choices"`
	Usage   OpenAIUsage          `json:"usage"`
	Error   *OpenAIError         `json:"error,omitempty"`
}

// OpenAIChoice represents a choice in the OpenAI response
type OpenAIChoice struct {
	Index        int                    `json:"index"`
	Message      OpenAIMessage          `json:"message"`
	FinishReason string                 `json:"finish_reason"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error from OpenAI API
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService(config map[string]interface{}) (AIService, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("openai api_key is required")
	}

	model, ok := config["model"].(string)
	if !ok || model == "" {
		model = "gpt-4" // Default model
	}

	baseURL, ok := config["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	maxTokens, ok := config["max_tokens"].(int)
	if !ok {
		maxTokens = 4096
	}

	return &OpenAIService{
		apiKey:    apiKey,
		model:     model,
		baseURL:   baseURL,
		maxTokens: maxTokens,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Ask sends a request to OpenAI and returns the response
func (o *OpenAIService) Ask(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	messages := []OpenAIMessage{
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
		temperature = 0.7
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = o.maxTokens
	}

	openAIReq := OpenAIRequest{
		Model:       o.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	startTime := time.Now()
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("openai api error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	responseTime := time.Since(startTime)

	return &AIResponse{
		Answer:       openAIResp.Choices[0].Message.Content,
		ResponseTime: responseTime,
		TokensUsed:   openAIResp.Usage.TotalTokens,
		Model:        openAIResp.Model,
		Provider:     ProviderOpenAI,
	}, nil
}

// GetCapabilities returns the capabilities of OpenAI
func (o *OpenAIService) GetCapabilities() AICapabilities {
	return AICapabilities{
		SupportsStreaming:     true,
		SupportsConversations: true,
		SupportedLanguages:    []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
		MaxTokens:             o.maxTokens,
		MaxContextLength:      getModelContextLength(o.model),
		SupportsDocuments:     true,
		SupportsImages:        isVisionModel(o.model),
	}
}

// GetProviderInfo returns information about OpenAI
func (o *OpenAIService) GetProviderInfo() AIProviderInfo {
	return AIProviderInfo{
		Name:        "OpenAI",
		Provider:    ProviderOpenAI,
		Model:       o.model,
		Version:     "1.0",
		Description: "OpenAI GPT models for classical education",
		IsLocal:     false,
		Cost:        getCostInfo(o.model),
	}
}

// IsHealthy checks if OpenAI service is available
func (o *OpenAIService) IsHealthy(ctx context.Context) error {
	// Simple health check - try to make a minimal request
	healthReq := &AIRequest{
		Question:  "Test",
		Context:   "You are a helpful assistant. Respond with just 'OK'.",
		MaxTokens: 5,
	}

	_, err := o.Ask(ctx, healthReq)
	return err
}

// getModelContextLength returns the context length for different OpenAI models
func getModelContextLength(model string) int {
	switch model {
	case "gpt-4":
		return 8192
	case "gpt-4-32k":
		return 32768
	case "gpt-4-turbo-preview", "gpt-4-0125-preview":
		return 128000
	case "gpt-3.5-turbo":
		return 4096
	case "gpt-3.5-turbo-16k":
		return 16384
	default:
		return 4096
	}
}

// isVisionModel checks if the model supports vision
func isVisionModel(model string) bool {
	visionModels := []string{"gpt-4-vision-preview", "gpt-4-turbo", "gpt-4o"}
	for _, vm := range visionModels {
		if model == vm {
			return true
		}
	}
	return false
}

// getCostInfo returns cost information for different models
func getCostInfo(model string) string {
	switch model {
	case "gpt-4":
		return "~$0.03/1K tokens (input), ~$0.06/1K tokens (output)"
	case "gpt-4-turbo-preview":
		return "~$0.01/1K tokens (input), ~$0.03/1K tokens (output)"
	case "gpt-3.5-turbo":
		return "~$0.001/1K tokens (input), ~$0.002/1K tokens (output)"
	default:
		return "Variable pricing - check OpenAI pricing page"
	}
}