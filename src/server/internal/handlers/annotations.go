package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/middleware"
	"github.com/classius/server/internal/models"
)

// CreateAnnotationRequest represents the request to create an annotation
type CreateAnnotationRequest struct {
	BookID        uuid.UUID `json:"book_id" binding:"required"`
	Type          string    `json:"type" binding:"required,oneof=highlight note bookmark"`
	PageNumber    int       `json:"page_number"`
	StartPosition int       `json:"start_position"`
	EndPosition   int       `json:"end_position"`
	SelectedText  string    `json:"selected_text"`
	Content       string    `json:"content"`
	Color         string    `json:"color"`
	Tags          []string  `json:"tags"`
	IsPrivate     bool      `json:"is_private"`
}

// UpdateAnnotationRequest represents the request to update an annotation
type UpdateAnnotationRequest struct {
	Content   string   `json:"content,omitempty"`
	Color     string   `json:"color,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	IsPrivate *bool    `json:"is_private,omitempty"`
}

// AnnotationResponse represents an annotation in API responses
type AnnotationResponse struct {
	ID            uuid.UUID `json:"id"`
	BookID        uuid.UUID `json:"book_id"`
	BookTitle     string    `json:"book_title,omitempty"`
	Type          string    `json:"type"`
	PageNumber    int       `json:"page_number"`
	StartPosition int       `json:"start_position"`
	EndPosition   int       `json:"end_position"`
	SelectedText  string    `json:"selected_text"`
	Content       string    `json:"content"`
	Color         string    `json:"color"`
	Tags          []string  `json:"tags"`
	IsPrivate     bool      `json:"is_private"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncAnnotationsRequest represents bulk annotation sync request
type SyncAnnotationsRequest struct {
	LastSyncTime *time.Time                `json:"last_sync_time,omitempty"`
	Annotations  []CreateAnnotationRequest `json:"annotations"`
}

// SyncAnnotationsResponse represents bulk annotation sync response
type SyncAnnotationsResponse struct {
	ServerAnnotations []AnnotationResponse `json:"server_annotations"`
	ConflictCount     int                  `json:"conflict_count"`
	SyncTime          time.Time            `json:"sync_time"`
}

// ToAnnotationResponse converts an Annotation model to AnnotationResponse
func ToAnnotationResponse(annotation *models.Annotation) AnnotationResponse {
	response := AnnotationResponse{
		ID:            annotation.ID,
		BookID:        annotation.BookID,
		Type:          annotation.Type,
		PageNumber:    annotation.PageNumber,
		StartPosition: annotation.StartPosition,
		EndPosition:   annotation.EndPosition,
		SelectedText:  annotation.SelectedText,
		Content:       annotation.Content,
		Color:         annotation.Color,
		Tags:          annotation.Tags,
		IsPrivate:     annotation.IsPrivate,
		CreatedAt:     annotation.CreatedAt,
		UpdatedAt:     annotation.UpdatedAt,
	}

	// Add book title if book is loaded
	if annotation.Book.ID != uuid.Nil {
		response.BookTitle = annotation.Book.Title
	}

	return response
}

// GetAnnotations returns user's annotations with filtering and pagination
func GetAnnotations(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to access your annotations",
		})
		return
	}

	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	bookID := c.Query("book_id")
	annotationType := c.Query("type")
	searchQuery := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	query := db.DB.Preload("Book").Where("user_id = ?", user.ID)

	if bookID != "" {
		query = query.Where("book_id = ?", bookID)
	}
	if annotationType != "" {
		query = query.Where("type = ?", annotationType)
	}
	if searchQuery != "" {
		query = query.Where("selected_text ILIKE ? OR content ILIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	var annotations []models.Annotation
	var total int64

	// Count total records
	query.Model(&models.Annotation{}).Count(&total)

	// Get paginated records
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&annotations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load annotations",
			"message": "Unable to retrieve annotations",
		})
		return
	}

	// Convert to response format
	var annotationData []AnnotationResponse
	for _, annotation := range annotations {
		annotationData = append(annotationData, ToAnnotationResponse(&annotation))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Annotations retrieved successfully",
		"data": gin.H{
			"annotations": annotationData,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		},
	})
}

// CreateAnnotation creates a new annotation
func CreateAnnotation(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to create annotations",
		})
		return
	}

	var req CreateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Verify the book exists and user has access
	var userBook models.UserBook
	if err := db.DB.Where("user_id = ? AND book_id = ?", user.ID, req.BookID).First(&userBook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Book not found",
			"message": "This book is not in your library",
		})
		return
	}

	// Create annotation
	annotation := models.Annotation{
		UserID:        user.ID,
		BookID:        req.BookID,
		Type:          req.Type,
		PageNumber:    req.PageNumber,
		StartPosition: req.StartPosition,
		EndPosition:   req.EndPosition,
		SelectedText:  req.SelectedText,
		Content:       req.Content,
		Color:         req.Color,
		Tags:          req.Tags,
		IsPrivate:     req.IsPrivate,
	}

	if err := db.DB.Create(&annotation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create annotation",
			"message": "Unable to save annotation",
		})
		return
	}

	// Load annotation with book information
	if err := db.DB.Preload("Book").First(&annotation, annotation.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load annotation",
			"message": "Annotation created but unable to retrieve details",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Annotation created successfully",
		"data":    ToAnnotationResponse(&annotation),
	})
}

// UpdateAnnotation updates an existing annotation
func UpdateAnnotation(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to update annotations",
		})
		return
	}

	annotationID := c.Param("id")
	if annotationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing annotation ID",
			"message": "Annotation ID is required",
		})
		return
	}

	var req UpdateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Find annotation and verify ownership
	var annotation models.Annotation
	if err := db.DB.Where("id = ? AND user_id = ?", annotationID, user.ID).First(&annotation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Annotation not found",
				"message": "The requested annotation does not exist or you don't have permission to access it",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"message": "Unable to retrieve annotation",
		})
		return
	}

	// Build update data
	updateData := make(map[string]interface{})
	if req.Content != "" {
		updateData["content"] = req.Content
	}
	if req.Color != "" {
		updateData["color"] = req.Color
	}
	if req.Tags != nil {
		updateData["tags"] = req.Tags
	}
	if req.IsPrivate != nil {
		updateData["is_private"] = *req.IsPrivate
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No updates provided",
			"message": "At least one field must be updated",
		})
		return
	}

	// Update annotation
	if err := db.DB.Model(&annotation).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update annotation",
			"message": "Unable to save annotation changes",
		})
		return
	}

	// Load updated annotation with book information
	if err := db.DB.Preload("Book").First(&annotation, annotation.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load updated annotation",
			"message": "Annotation updated but unable to retrieve updated details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Annotation updated successfully",
		"data":    ToAnnotationResponse(&annotation),
	})
}

// DeleteAnnotation deletes an annotation
func DeleteAnnotation(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to delete annotations",
		})
		return
	}

	annotationID := c.Param("id")
	if annotationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing annotation ID",
			"message": "Annotation ID is required",
		})
		return
	}

	// Find annotation and verify ownership
	var annotation models.Annotation
	if err := db.DB.Where("id = ? AND user_id = ?", annotationID, user.ID).First(&annotation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Annotation not found",
				"message": "The requested annotation does not exist or you don't have permission to access it",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"message": "Unable to retrieve annotation",
		})
		return
	}

	// Delete annotation
	if err := db.DB.Delete(&annotation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete annotation",
			"message": "Unable to delete annotation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Annotation deleted successfully",
	})
}

// SyncAnnotations handles bidirectional annotation synchronization
func SyncAnnotations(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to sync annotations",
		})
		return
	}

	var req SyncAnnotationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	now := time.Now()
	var conflicts int

	// Process incoming annotations from device
	for _, annotationReq := range req.Annotations {
		// Verify user has access to the book
		var userBook models.UserBook
		if err := db.DB.Where("user_id = ? AND book_id = ?", user.ID, annotationReq.BookID).First(&userBook).Error; err != nil {
			conflicts++
			continue // Skip this annotation if user doesn't have access to book
		}

		// Create annotation
		annotation := models.Annotation{
			UserID:        user.ID,
			BookID:        annotationReq.BookID,
			Type:          annotationReq.Type,
			PageNumber:    annotationReq.PageNumber,
			StartPosition: annotationReq.StartPosition,
			EndPosition:   annotationReq.EndPosition,
			SelectedText:  annotationReq.SelectedText,
			Content:       annotationReq.Content,
			Color:         annotationReq.Color,
			Tags:          annotationReq.Tags,
			IsPrivate:     annotationReq.IsPrivate,
		}

		if err := db.DB.Create(&annotation).Error; err != nil {
			conflicts++
			// Log error but continue with other annotations
		}
	}

	// Get server annotations modified since last sync
	query := db.DB.Preload("Book").Where("user_id = ?", user.ID)
	if req.LastSyncTime != nil {
		query = query.Where("updated_at > ?", *req.LastSyncTime)
	}

	var serverAnnotations []models.Annotation
	if err := query.Order("updated_at ASC").Find(&serverAnnotations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load server annotations",
			"message": "Unable to retrieve server annotations for sync",
		})
		return
	}

	// Convert to response format
	var annotationData []AnnotationResponse
	for _, annotation := range serverAnnotations {
		annotationData = append(annotationData, ToAnnotationResponse(&annotation))
	}

	response := SyncAnnotationsResponse{
		ServerAnnotations: annotationData,
		ConflictCount:     conflicts,
		SyncTime:          now,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Annotations synchronized successfully",
		"data":    response,
	})
}