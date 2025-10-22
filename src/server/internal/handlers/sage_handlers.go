package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/classius/server/internal/services"
	"github.com/classius/server/internal/utils"
)

// SageHandlers manages AI Sage-related endpoints
type SageHandlers struct {
	sageService *services.SageService
}

// NewSageHandlers creates new Sage handlers
func NewSageHandlers(sageService *services.SageService) *SageHandlers {
	return &SageHandlers{
		sageService: sageService,
	}
}

// SageQuestionRequest represents a question to the Sage
type SageQuestionRequest struct {
	Question      string `json:"question" binding:"required"`
	BookTitle     string `json:"book_title,omitempty"`
	BookAuthor    string `json:"book_author,omitempty"`
	BookID        string `json:"book_id,omitempty"`
	PassageText   string `json:"passage_text,omitempty"`
	AnnotationID  string `json:"annotation_id,omitempty"`
	Context       string `json:"context,omitempty"`
}

// SageQuestionResponse represents the Sage's response
type SageQuestionResponse struct {
	Answer          string        `json:"answer"`
	ResponseTime    string        `json:"response_time"`
	Model           string        `json:"model"`
	Provider        string        `json:"provider"`
	TokensUsed      int           `json:"tokens_used,omitempty"`
	ConversationID  string        `json:"conversation_id,omitempty"`
	Sources         []string      `json:"sources,omitempty"`
	Confidence      float64       `json:"confidence,omitempty"`
}

// SageCapabilitiesResponse represents the Sage's capabilities
type SageCapabilitiesResponse struct {
	SupportsStreaming     bool     `json:"supports_streaming"`
	SupportsConversations bool     `json:"supports_conversations"`
	SupportedLanguages    []string `json:"supported_languages"`
	MaxTokens             int      `json:"max_tokens"`
	MaxContextLength      int      `json:"max_context_length"`
	SupportsDocuments     bool     `json:"supports_documents"`
	SupportsImages        bool     `json:"supports_images"`
	ProviderInfo          services.AIProviderInfo `json:"provider_info"`
}

// AskSage handles questions submitted to the AI Sage
// POST /api/sage/ask
func (h *SageHandlers) AskSage(c *gin.Context) {
	var req SageQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format", nil)
		return
	}

	// Validate question length
	if len(req.Question) > 5000 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Question too long (max 5000 characters)", nil)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Ask the Sage
	startTime := time.Now()
	response, err := h.sageService.Ask(ctx, userIDStr, req.Question, req.BookTitle, req.BookAuthor, req.PassageText)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get response from Sage", err)
		return
	}

	// Format response
	sageResponse := SageQuestionResponse{
		Answer:       response.Answer,
		ResponseTime: time.Since(startTime).String(),
		Model:        response.Model,
		Provider:     string(response.Provider),
		TokensUsed:   response.TokensUsed,
		Sources:      response.Sources,
		Confidence:   response.Confidence,
	}

	// TODO: Save conversation to database
	// This would typically involve injecting a repository and saving the conversation

	utils.SuccessResponse(c, "Sage response generated", sageResponse)
}

// GetSageCapabilities returns the Sage's current capabilities
// GET /api/sage/capabilities
func (h *SageHandlers) GetSageCapabilities(c *gin.Context) {
	capabilities := h.sageService.GetCapabilities()
	providerInfo := h.sageService.GetProviderInfo()

	response := SageCapabilitiesResponse{
		SupportsStreaming:     capabilities.SupportsStreaming,
		SupportsConversations: capabilities.SupportsConversations,
		SupportedLanguages:    capabilities.SupportedLanguages,
		MaxTokens:             capabilities.MaxTokens,
		MaxContextLength:      capabilities.MaxContextLength,
		SupportsDocuments:     capabilities.SupportsDocuments,
		SupportsImages:        capabilities.SupportsImages,
		ProviderInfo:          providerInfo,
	}

	utils.SuccessResponse(c, "Sage capabilities retrieved", response)
}

// CheckSageHealth checks if the Sage service is healthy
// GET /api/sage/health
func (h *SageHandlers) CheckSageHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.sageService.IsHealthy(ctx)
	if err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Sage service is unhealthy", err)
		return
	}

	providerInfo := h.sageService.GetProviderInfo()
	healthResponse := gin.H{
		"status": "healthy",
		"provider": providerInfo.Name,
		"model": providerInfo.Model,
		"is_local": providerInfo.IsLocal,
		"timestamp": time.Now().UTC(),
	}

	utils.SuccessResponse(c, "Sage service is healthy", healthResponse)
}

// GetSageConversations retrieves user's conversation history with the Sage
// GET /api/sage/conversations
func (h *SageHandlers) GetSageConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get query parameters
	limit := utils.GetIntQuery(c, "limit", 50, 1, 100)
	offset := utils.GetIntQuery(c, "offset", 0, 0, 10000)
	bookID := c.Query("book_id")

	// TODO: Implement conversation retrieval from database
	// This would typically involve injecting a repository service
	
	// For now, return a placeholder response
	conversations := []gin.H{
		{
			"id":           "placeholder-conversation-1",
			"user_id":      userID,
			"book_id":      bookID,
			"question":     "What is the significance of this passage?",
			"response":     "This passage demonstrates...",
			"created_at":   time.Now().Add(-24 * time.Hour).UTC(),
			"response_time": "2.3s",
			"model":        "gpt-4",
			"provider":     "openai",
		},
	}

	response := gin.H{
		"conversations": conversations,
		"total":         len(conversations),
		"limit":         limit,
		"offset":        offset,
	}

	utils.SuccessResponse(c, "Conversations retrieved", response)
}

// GetSageConversation retrieves a specific conversation
// GET /api/sage/conversations/:id
func (h *SageHandlers) GetSageConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// TODO: Implement single conversation retrieval from database
	// This would typically involve injecting a repository service
	
	// For now, return a placeholder response
	conversation := gin.H{
		"id":           conversationID,
		"user_id":      userID,
		"book_id":      "sample-book-id",
		"book_title":   "The Republic",
		"book_author":  "Plato",
		"question":     "What is Plato's theory of Forms?",
		"response":     "Plato's theory of Forms suggests that there exists a realm of perfect, eternal, and unchangeable Forms or Ideas...",
		"passage_text": "Then if there are any such things as absolute essences...",
		"created_at":   time.Now().Add(-2 * time.Hour).UTC(),
		"response_time": "1.8s",
		"model":        "gpt-4",
		"provider":     "openai",
		"tokens_used":  450,
	}

	utils.SuccessResponse(c, "Conversation retrieved", conversation)
}

// DeleteSageConversation deletes a conversation
// DELETE /api/sage/conversations/:id
func (h *SageHandlers) DeleteSageConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	// Get user ID from context
	_, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// TODO: Implement conversation deletion from database
	// This would verify the conversation belongs to the user before deletion

	utils.SuccessResponse(c, "Conversation deleted", gin.H{
		"conversation_id": conversationID,
		"deleted_at":      time.Now().UTC(),
	})
}

// GetSageStats returns usage statistics for the Sage
// GET /api/sage/stats
func (h *SageHandlers) GetSageStats(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get time range from query parameters
	timeRange := c.DefaultQuery("range", "30d") // 7d, 30d, 90d, all

	// TODO: Implement stats retrieval from database
	// This would calculate actual usage statistics

	// For now, return placeholder stats
	stats := gin.H{
		"user_id":              userID,
		"time_range":           timeRange,
		"total_questions":      42,
		"total_conversations":  15,
		"avg_response_time":    "2.1s",
		"most_used_provider":   "openai",
		"total_tokens_used":    12450,
		"favorite_subjects":    []string{"Philosophy", "Literature", "History"},
		"questions_by_day":     []gin.H{
			{"date": "2024-01-15", "count": 3},
			{"date": "2024-01-14", "count": 5},
			{"date": "2024-01-13", "count": 2},
		},
		"provider_usage": gin.H{
			"openai": 35,
			"local":  7,
		},
	}

	utils.SuccessResponse(c, "Sage statistics retrieved", stats)
}

// ExportSageData exports user's Sage conversation data
// GET /api/sage/export
func (h *SageHandlers) ExportSageData(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	format := c.DefaultQuery("format", "json") // json, csv, txt
	
	if format != "json" && format != "csv" && format != "txt" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Unsupported export format", nil)
		return
	}

	// TODO: Implement data export from database
	// This would gather all user's Sage conversations and format them

	// For now, return a placeholder export
	exportData := gin.H{
		"user_id":     userID,
		"exported_at": time.Now().UTC(),
		"format":      format,
		"conversations": []gin.H{
			{
				"id":         "conv-1",
				"question":   "What is the meaning of virtue in Aristotle's ethics?",
				"response":   "According to Aristotle...",
				"book":       "Nicomachean Ethics",
				"created_at": "2024-01-15T10:00:00Z",
			},
		},
	}

	// Set appropriate content type based on format
	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sage-conversations-%s.json", time.Now().Format("2006-01-02")))
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sage-conversations-%s.csv", time.Now().Format("2006-01-02")))
	case "txt":
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=sage-conversations-%s.txt", time.Now().Format("2006-01-02")))
	}

	c.JSON(http.StatusOK, exportData)
}