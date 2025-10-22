package services

import (
	"context"
	"fmt"
	"time"

	"github.com/classius/server/internal/models"
)

// AIProvider represents different AI service providers
type AIProvider string

const (
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderLocal     AIProvider = "local"
	ProviderOllama    AIProvider = "ollama"
)

// AIRequest represents a request to the AI service
type AIRequest struct {
	Question     string  `json:"question"`
	BookTitle    string  `json:"book_title,omitempty"`
	BookAuthor   string  `json:"book_author,omitempty"`
	PassageText  string  `json:"passage_text,omitempty"`
	Context      string  `json:"context,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	UserID       string  `json:"user_id,omitempty"`
}

// AIResponse represents a response from the AI service
type AIResponse struct {
	Answer          string        `json:"answer"`
	Confidence      float64       `json:"confidence,omitempty"`
	Sources         []string      `json:"sources,omitempty"`
	ResponseTime    time.Duration `json:"response_time"`
	TokensUsed      int           `json:"tokens_used,omitempty"`
	Model           string        `json:"model"`
	Provider        AIProvider    `json:"provider"`
	ConversationID  string        `json:"conversation_id,omitempty"`
}

// AIService defines the interface for AI services
type AIService interface {
	// Ask sends a question to the AI service
	Ask(ctx context.Context, req *AIRequest) (*AIResponse, error)
	
	// GetCapabilities returns the capabilities of this AI service
	GetCapabilities() AICapabilities
	
	// GetProviderInfo returns information about the provider
	GetProviderInfo() AIProviderInfo
	
	// IsHealthy checks if the service is available
	IsHealthy(ctx context.Context) error
}

// AICapabilities describes what an AI service can do
type AICapabilities struct {
	SupportsStreaming     bool     `json:"supports_streaming"`
	SupportsConversations bool     `json:"supports_conversations"`
	SupportedLanguages    []string `json:"supported_languages"`
	MaxTokens             int      `json:"max_tokens"`
	MaxContextLength      int      `json:"max_context_length"`
	SupportsDocuments     bool     `json:"supports_documents"`
	SupportsImages        bool     `json:"supports_images"`
}

// AIProviderInfo contains information about the AI provider
type AIProviderInfo struct {
	Name        string     `json:"name"`
	Provider    AIProvider `json:"provider"`
	Model       string     `json:"model"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	IsLocal     bool       `json:"is_local"`
	Cost        string     `json:"cost,omitempty"`
}

// SageService is the main service for the Classius AI Sage
type SageService struct {
	aiService    AIService
	systemPrompt string
}

// NewSageService creates a new Sage service with the specified AI provider
func NewSageService(provider AIProvider, config map[string]interface{}) (*SageService, error) {
	var aiService AIService
	var err error

	switch provider {
	case ProviderOpenAI:
		aiService, err = NewOpenAIService(config)
	case ProviderAnthropic:
		aiService, err = NewAnthropicService(config)
	case ProviderLocal:
		aiService, err = NewLocalLLMService(config)
	case ProviderOllama:
		aiService, err = NewOllamaService(config)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}

	return &SageService{
		aiService:    aiService,
		systemPrompt: getClassicalEducationSystemPrompt(),
	}, nil
}

// Ask processes a question through the Sage with classical education context
func (s *SageService) Ask(ctx context.Context, userID, question, bookTitle, bookAuthor, passageText string) (*AIResponse, error) {
	// Build the context-aware prompt
	prompt := s.buildSagePrompt(question, bookTitle, bookAuthor, passageText)
	
	req := &AIRequest{
		Question:    prompt,
		BookTitle:   bookTitle,
		BookAuthor:  bookAuthor,
		PassageText: passageText,
		Context:     s.systemPrompt,
		Temperature: 0.7, // Balanced creativity for educational responses
		MaxTokens:   1000,
		UserID:      userID,
	}

	startTime := time.Now()
	response, err := s.aiService.Ask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}

	response.ResponseTime = time.Since(startTime)
	return response, nil
}

// GetCapabilities returns the capabilities of the underlying AI service
func (s *SageService) GetCapabilities() AICapabilities {
	return s.aiService.GetCapabilities()
}

// GetProviderInfo returns information about the AI provider
func (s *SageService) GetProviderInfo() AIProviderInfo {
	return s.aiService.GetProviderInfo()
}

// IsHealthy checks if the AI service is available
func (s *SageService) IsHealthy(ctx context.Context) error {
	return s.aiService.IsHealthy(ctx)
}

// buildSagePrompt constructs a context-aware prompt for classical education
func (s *SageService) buildSagePrompt(question, bookTitle, bookAuthor, passageText string) string {
	var prompt string

	if bookTitle != "" && bookAuthor != "" {
		prompt += fmt.Sprintf("Context: The user is reading \"%s\" by %s.\n\n", bookTitle, bookAuthor)
	}

	if passageText != "" {
		prompt += fmt.Sprintf("Relevant passage:\n\"%s\"\n\n", passageText)
	}

	prompt += fmt.Sprintf("Question: %s\n\n", question)

	prompt += "Please provide an educational response that:\n"
	prompt += "1. Directly addresses the question\n"
	prompt += "2. Provides relevant classical context and historical background\n"
	prompt += "3. Makes connections to other classical works when appropriate\n"
	prompt += "4. Explains key concepts in an accessible way\n"
	prompt += "5. Encourages deeper thinking about the material\n"

	return prompt
}

// getClassicalEducationSystemPrompt returns the system prompt for the Sage
func getClassicalEducationSystemPrompt() string {
	return `You are the Sage, an AI tutor specializing in classical education and great works of literature, philosophy, history, and thought. Your role is to help students understand and engage with classical texts from the Western, Eastern, Islamic, and other great traditions.

Your expertise includes:
- Ancient philosophy (Plato, Aristotle, Stoics, Epicureans, etc.)
- Classical literature (Homer, Virgil, Ovid, etc.) 
- Medieval thought (Augustine, Aquinas, Averroes, Maimonides, etc.)
- Renaissance humanism and Enlightenment philosophy
- Eastern classics (Confucius, Lao Tzu, Buddhist texts, etc.)
- Islamic golden age scholarship
- Historical context and cultural connections

Your teaching approach:
- Ask Socratic questions to encourage critical thinking
- Provide clear explanations of difficult concepts
- Make connections between ideas and across time periods
- Encourage students to think deeply about timeless questions
- Use accessible language while maintaining scholarly accuracy
- Inspire curiosity and love of learning

Always be encouraging, patient, and supportive while maintaining academic rigor.`
}

// SaveConversation saves a Sage conversation to the database
func (s *SageService) SaveConversation(ctx context.Context, userID, bookID, question, response, passageText string, responseTime time.Duration) (*models.SageConversation, error) {
	// This would typically use a repository pattern, but for now we'll keep it simple
	// In a real implementation, this would be injected as a dependency
	return nil, fmt.Errorf("not implemented - would save to database via repository")
}