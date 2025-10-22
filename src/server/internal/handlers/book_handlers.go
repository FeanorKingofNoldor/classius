package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/classius/server/internal/models"
	"github.com/classius/server/internal/services"
	"github.com/classius/server/internal/utils"
)

// BookHandlers manages book-related endpoints
type BookHandlers struct {
	bookService *services.BookService
}

// NewBookHandlers creates new book handlers
func NewBookHandlers(bookService *services.BookService) *BookHandlers {
	return &BookHandlers{
		bookService: bookService,
	}
}

// UploadBook handles book file uploads
// POST /api/books/upload
func (h *BookHandlers) UploadBook(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100MB max
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to parse form", err)
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided", err)
		return
	}

	// Parse metadata from form
	req := &services.BookUploadRequest{
		Title:       c.PostForm("title"),
		Author:      c.PostForm("author"),
		Language:    c.PostForm("language"),
		Genre:       c.PostForm("genre"),
		Publisher:   c.PostForm("publisher"),
		ISBN:        c.PostForm("isbn"),
		Description: c.PostForm("description"),
		IsPublic:    c.PostForm("is_public") == "true",
	}

	// Validate required fields
	if req.Title == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Title is required", nil)
		return
	}
	if req.Author == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Author is required", nil)
		return
	}

	// Parse published date if provided
	if publishedAtStr := c.PostForm("published_at"); publishedAtStr != "" {
		if publishedAt, err := time.Parse("2006-01-02", publishedAtStr); err == nil {
			req.PublishedAt = &publishedAt
		}
	}

	// Parse tags if provided
	if tagsStr := c.PostForm("tags"); tagsStr != "" {
		req.Tags = strings.Split(tagsStr, ",")
		// Trim whitespace from tags
		for i, tag := range req.Tags {
			req.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Upload book
	book, err := h.bookService.UploadBook(c.Request.Context(), userUUID, file, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload book", err)
		return
	}

	utils.SuccessResponse(c, "Book uploaded successfully", book)
}

// GetBooks retrieves books with filtering and pagination
// GET /api/books
func (h *BookHandlers) GetBooks(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Build filter from query parameters
	filter := &models.BookFilter{
		UserID:   userUUID,
		Query:    c.Query("q"),
		Author:   c.Query("author"),
		Genre:    c.Query("genre"),
		Language: c.Query("language"),
		FileType: c.Query("file_type"),
		SortBy:   c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
		Page:     utils.GetIntQuery(c, "page", 1, 1, 1000),
		PerPage:  utils.GetIntQuery(c, "per_page", 20, 1, 100),
	}

	// Parse status if provided
	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = models.BookStatus(statusStr)
	}

	// Parse is_public if provided
	if isPublicStr := c.Query("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			filter.IsPublic = &isPublic
		}
	}

	// Parse tags if provided
	if tagsStr := c.Query("tags"); tagsStr != "" {
		filter.Tags = strings.Split(tagsStr, ",")
		for i, tag := range filter.Tags {
			filter.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Parse date filters if provided
	if createdAfterStr := c.Query("created_after"); createdAfterStr != "" {
		if createdAfter, err := time.Parse("2006-01-02", createdAfterStr); err == nil {
			filter.CreatedAfter = &createdAfter
		}
	}
	if createdBeforeStr := c.Query("created_before"); createdBeforeStr != "" {
		if createdBefore, err := time.Parse("2006-01-02", createdBeforeStr); err == nil {
			filter.CreatedBefore = &createdBefore
		}
	}

	// Get books
	result, err := h.bookService.GetBooks(c.Request.Context(), filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve books", err)
		return
	}

	utils.SuccessResponse(c, "Books retrieved successfully", result)
}

// GetBook retrieves a specific book
// GET /api/books/:id
func (h *BookHandlers) GetBook(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID", err)
		return
	}

	// Get book
	book, err := h.bookService.GetBook(c.Request.Context(), userUUID, bookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Book not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve book", err)
		}
		return
	}

	utils.SuccessResponse(c, "Book retrieved successfully", book)
}

// UpdateBook updates book metadata
// PUT /api/books/:id
func (h *BookHandlers) UpdateBook(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID", err)
		return
	}

	// Parse request body
	var req services.BookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Update book
	book, err := h.bookService.UpdateBook(c.Request.Context(), userUUID, bookID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Book not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update book", err)
		}
		return
	}

	utils.SuccessResponse(c, "Book updated successfully", book)
}

// DeleteBook deletes a book
// DELETE /api/books/:id
func (h *BookHandlers) DeleteBook(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID", err)
		return
	}

	// Delete book
	if err := h.bookService.DeleteBook(c.Request.Context(), userUUID, bookID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Book not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete book", err)
		}
		return
	}

	utils.SuccessResponse(c, "Book deleted successfully", gin.H{
		"book_id":    bookID,
		"deleted_at": time.Now().UTC(),
	})
}

// DownloadBook serves book files for download
// GET /api/books/:id/download
func (h *BookHandlers) DownloadBook(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID", err)
		return
	}

	// Get book
	book, err := h.bookService.GetBook(c.Request.Context(), userUUID, bookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Book not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve book", err)
		}
		return
	}

	// Check if file exists
	if _, err := os.Stat(book.FilePath); os.IsNotExist(err) {
		utils.ErrorResponse(c, http.StatusNotFound, "Book file not found", err)
		return
	}

	// Set appropriate headers
	filename := fmt.Sprintf("%s - %s%s", book.Title, book.Author, book.GetFileExtension())
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", book.Metadata.MimeType)

	// Serve file
	c.File(book.FilePath)
}

// GetBookContent streams book content for reading
// GET /api/books/:id/content
func (h *BookHandlers) GetBookContent(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID", err)
		return
	}

	// Get book
	book, err := h.bookService.GetBook(c.Request.Context(), userUUID, bookID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.ErrorResponse(c, http.StatusNotFound, "Book not found", err)
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve book", err)
		}
		return
	}

	// Check if file exists
	if _, err := os.Stat(book.FilePath); os.IsNotExist(err) {
		utils.ErrorResponse(c, http.StatusNotFound, "Book file not found", err)
		return
	}

	// Set appropriate headers for inline viewing
	c.Header("Content-Type", book.Metadata.MimeType)
	c.Header("Content-Disposition", "inline")

	// Serve file for inline viewing
	c.File(book.FilePath)
}

// GetBookStats returns statistics about user's book collection
// GET /api/books/stats
func (h *BookHandlers) GetBookStats(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Get stats
	stats, err := h.bookService.GetBookStats(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve stats", err)
		return
	}

	utils.SuccessResponse(c, "Book statistics retrieved successfully", stats)
}

// GetTags retrieves all user tags
// GET /api/books/tags
func (h *BookHandlers) GetTags(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Get tags for user
	var tags []models.Tag
	// Note: This would typically be in the BookService, but adding here for simplicity
	if err := h.bookService.GetDB().
		Where("user_id = ?", userUUID).
		Order("name ASC").
		Find(&tags).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tags", err)
		return
	}

	utils.SuccessResponse(c, "Tags retrieved successfully", tags)
}

// CreateTag creates a new tag
// POST /api/books/tags
func (h *BookHandlers) CreateTag(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse request
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Create tag
	tag := models.Tag{
		ID:     uuid.New(),
		UserID: userUUID,
		Name:   req.Name,
		Color:  req.Color,
	}

	// Note: This would typically be in the BookService
	if err := h.bookService.GetDB().Create(&tag).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create tag", err)
		return
	}

	utils.SuccessResponse(c, "Tag created successfully", tag)
}

// DeleteTag deletes a tag
// DELETE /api/books/tags/:id
func (h *BookHandlers) DeleteTag(c *gin.Context) {
	// Get user ID from context
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

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	// Parse tag ID from URL
	tagIDStr := c.Param("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid tag ID", err)
		return
	}

	// Delete tag (this will also remove book-tag associations)
	result := h.bookService.GetDB().
		Where("id = ? AND user_id = ?", tagID, userUUID).
		Delete(&models.Tag{})

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete tag", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Tag not found", nil)
		return
	}

	utils.SuccessResponse(c, "Tag deleted successfully", gin.H{
		"tag_id":     tagID,
		"deleted_at": time.Now().UTC(),
	})
}