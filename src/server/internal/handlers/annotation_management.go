package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/models"
	"github.com/classius/server/internal/utils"
)

// AnnotationFilterRequest represents advanced filtering options for annotations
type AnnotationFilterRequest struct {
	Query        string    `json:"query" form:"query"`
	BookIDs      []string  `json:"book_ids" form:"book_ids"`
	Types        []string  `json:"types" form:"types"`         // highlight, note, bookmark
	Authors      []string  `json:"authors" form:"authors"`
	Tags         []string  `json:"tags" form:"tags"`
	Colors       []string  `json:"colors" form:"colors"`
	DateFrom     string    `json:"date_from" form:"date_from"`
	DateTo       string    `json:"date_to" form:"date_to"`
	IsPrivate    *bool     `json:"is_private" form:"is_private"`
	MinPage      *int      `json:"min_page" form:"min_page"`
	MaxPage      *int      `json:"max_page" form:"max_page"`
	SortBy       string    `json:"sort_by" form:"sort_by"`     // date, book, page, type
	SortOrder    string    `json:"sort_order" form:"sort_order"` // asc, desc
	Page         int       `json:"page" form:"page"`
	PerPage      int       `json:"per_page" form:"per_page"`
	IncludeStats bool      `json:"include_stats" form:"include_stats"`
}

// AnnotationListResponse represents the response for annotation listing
type AnnotationListResponse struct {
	Annotations []EnhancedAnnotationResponse `json:"annotations"`
	Pagination  PaginationResponse           `json:"pagination"`
	Filters     AnnotationFiltersResponse    `json:"filters,omitempty"`
	Stats       *AnnotationStatsResponse     `json:"stats,omitempty"`
}

// EnhancedAnnotationResponse represents an enhanced annotation response with book details
type EnhancedAnnotationResponse struct {
	ID           uuid.UUID `json:"id"`
	BookID       uuid.UUID `json:"book_id"`
	BookTitle    string    `json:"book_title"`
	BookAuthor   string    `json:"book_author"`
	BookCover    string    `json:"book_cover"`
	Type         string    `json:"type"`
	PageNumber   int       `json:"page_number"`
	StartPos     int       `json:"start_position"`
	EndPos       int       `json:"end_position"`
	SelectedText string    `json:"selected_text"`
	Content      string    `json:"content"`
	Color        string    `json:"color"`
	Tags         []string  `json:"tags"`
	IsPrivate    bool      `json:"is_private"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// AnnotationFiltersResponse represents available filter options
type AnnotationFiltersResponse struct {
	AvailableTypes    []string `json:"available_types"`
	AvailableBooks    []BookFilterOption `json:"available_books"`
	AvailableAuthors  []string `json:"available_authors"`
	AvailableTags     []string `json:"available_tags"`
	AvailableColors   []string `json:"available_colors"`
}

// BookFilterOption represents a book option for filtering
type BookFilterOption struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// AnnotationStatsResponse represents annotation statistics
type AnnotationStatsResponse struct {
	TotalAnnotations int `json:"total_annotations"`
	TotalHighlights  int `json:"total_highlights"`
	TotalNotes       int `json:"total_notes"`
	TotalBookmarks   int `json:"total_bookmarks"`
	BooksWithAnnotations int `json:"books_with_annotations"`
	MostUsedTags     []TagUsage `json:"most_used_tags"`
	AnnotationsByMonth []MonthlyAnnotations `json:"annotations_by_month"`
}

// TagUsage represents tag usage statistics
type TagUsage struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// MonthlyAnnotations represents annotations grouped by month
type MonthlyAnnotations struct {
	Month       string `json:"month"`
	Count       int    `json:"count"`
	Highlights  int    `json:"highlights"`
	Notes       int    `json:"notes"`
	Bookmarks   int    `json:"bookmarks"`
}

// BulkActionRequest represents bulk operations on annotations
type BulkActionRequest struct {
	AnnotationIDs []string               `json:"annotation_ids" binding:"required"`
	Action        string                 `json:"action" binding:"required"` // delete, update_tags, update_color, toggle_private
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// BulkActionResponse represents the result of bulk operations
type BulkActionResponse struct {
	Success   int      `json:"success"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAnnotationsAdvanced returns annotations with advanced filtering and pagination
func GetAnnotationsAdvanced(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req AnnotationFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid filter parameters")
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}
	if req.SortBy == "" {
		req.SortBy = "date"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	database := db.GetDB()
	annotations, pagination, filters, stats, err := getAnnotationsWithFilters(database, userUUID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve annotations")
		return
	}

	response := AnnotationListResponse{
		Annotations: annotations,
		Pagination:  *pagination,
	}

	if req.Query == "" && req.Page == 1 {
		response.Filters = *filters
	}

	if req.IncludeStats {
		response.Stats = stats
	}

	utils.SuccessResponse(c, http.StatusOK, "Annotations retrieved successfully", response)
}

// CreateAnnotationEnhanced creates a new annotation with enhanced validation
func CreateAnnotationEnhanced(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req CreateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid annotation data: "+err.Error())
		return
	}

	// Validate annotation type
	validTypes := map[string]bool{"highlight": true, "note": true, "bookmark": true}
	if !validTypes[req.Type] {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid annotation type")
		return
	}

	database := db.GetDB()

	// Verify user has access to the book
	var userBook models.UserBook
	if err := database.Where("user_id = ? AND book_id = ? AND deleted_at IS NULL", userUUID, req.BookID).
		First(&userBook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusForbidden, "Book not found in your library")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify book access")
		return
	}

	// Create annotation
	annotation := models.Annotation{
		UserID:        userUUID,
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

	if err := database.Create(&annotation).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create annotation")
		return
	}

	// Load annotation with book details
	if err := database.Preload("Book").First(&annotation, annotation.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Annotation created but failed to load details")
		return
	}

	response := convertToEnhancedResponse(&annotation)
	utils.SuccessResponse(c, http.StatusCreated, "Annotation created successfully", response)
}

// UpdateAnnotationEnhanced updates an annotation with enhanced validation
func UpdateAnnotationEnhanced(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	annotationID := c.Param("id")
	if annotationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Annotation ID is required")
		return
	}

	var req UpdateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid update data: "+err.Error())
		return
	}

	database := db.GetDB()

	// Find and verify ownership
	var annotation models.Annotation
	if err := database.Where("id = ? AND user_id = ? AND deleted_at IS NULL", annotationID, userUUID).
		First(&annotation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Annotation not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve annotation")
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
		utils.ErrorResponse(c, http.StatusBadRequest, "No valid updates provided")
		return
	}

	// Update annotation
	if err := database.Model(&annotation).Updates(updateData).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update annotation")
		return
	}

	// Reload with book details
	if err := database.Preload("Book").First(&annotation, annotation.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Update completed but failed to load details")
		return
	}

	response := convertToEnhancedResponse(&annotation)
	utils.SuccessResponse(c, http.StatusOK, "Annotation updated successfully", response)
}

// DeleteAnnotationEnhanced deletes an annotation with enhanced validation
func DeleteAnnotationEnhanced(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	annotationID := c.Param("id")
	if annotationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Annotation ID is required")
		return
	}

	database := db.GetDB()

	// Find and verify ownership
	var annotation models.Annotation
	if err := database.Where("id = ? AND user_id = ? AND deleted_at IS NULL", annotationID, userUUID).
		First(&annotation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Annotation not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve annotation")
		return
	}

	// Soft delete
	if err := database.Delete(&annotation).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete annotation")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Annotation deleted successfully", nil)
}

// BulkAnnotationActions performs bulk operations on annotations
func BulkAnnotationActions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bulk action data: "+err.Error())
		return
	}

	if len(req.AnnotationIDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No annotations specified")
		return
	}

	if len(req.AnnotationIDs) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Too many annotations (max 100)")
		return
	}

	database := db.GetDB()
	response, err := performBulkAction(database, userUUID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Bulk action failed: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Bulk action completed", response)
}

// ExportAnnotations exports annotations in various formats
func ExportAnnotations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	format := c.DefaultQuery("format", "csv") // csv, json
	bookID := c.Query("book_id")
	annotationType := c.Query("type")

	database := db.GetDB()

	// Build query
	query := database.Table("annotations a").
		Select(`a.id, a.type, a.page_number, a.selected_text, a.content, a.color, a.tags, 
		        a.created_at, a.updated_at, b.title as book_title, b.author as book_author`).
		Joins("JOIN books b ON a.book_id = b.id").
		Where("a.user_id = ? AND a.deleted_at IS NULL AND b.deleted_at IS NULL", userUUID)

	if bookID != "" {
		query = query.Where("a.book_id = ?", bookID)
	}
	if annotationType != "" {
		query = query.Where("a.type = ?", annotationType)
	}

	type ExportAnnotation struct {
		ID           string    `json:"id"`
		Type         string    `json:"type"`
		PageNumber   int       `json:"page_number"`
		SelectedText string    `json:"selected_text"`
		Content      string    `json:"content"`
		Color        string    `json:"color"`
		Tags         []string  `json:"tags"`
		BookTitle    string    `json:"book_title"`
		BookAuthor   string    `json:"book_author"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	var annotations []ExportAnnotation
	if err := query.Order("b.title, a.page_number").Find(&annotations).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to export annotations")
		return
	}

	switch format {
	case "csv":
		exportAnnotationsCSV(c, annotations)
	case "json":
		exportAnnotationsJSON(c, annotations)
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Unsupported export format")
	}
}

// Helper functions

func getAnnotationsWithFilters(database *gorm.DB, userID uuid.UUID, req AnnotationFilterRequest) ([]EnhancedAnnotationResponse, *PaginationResponse, *AnnotationFiltersResponse, *AnnotationStatsResponse, error) {
	// Build base query
	query := database.Table("annotations a").
		Select(`a.id, a.book_id, a.type, a.page_number, a.start_position, a.end_position, 
		        a.selected_text, a.content, a.color, a.tags, a.is_private, a.created_at, a.updated_at,
		        b.title as book_title, b.author as book_author, b.cover_url as book_cover`).
		Joins("JOIN books b ON a.book_id = b.id").
		Joins("JOIN user_books ub ON a.book_id = ub.book_id AND ub.user_id = a.user_id").
		Where("a.user_id = ? AND a.deleted_at IS NULL AND b.deleted_at IS NULL AND ub.deleted_at IS NULL", userID)

	// Apply filters
	if req.Query != "" {
		searchTerms := strings.Fields(strings.ToLower(req.Query))
		for _, term := range searchTerms {
			query = query.Where("(LOWER(a.selected_text) LIKE ? OR LOWER(a.content) LIKE ? OR LOWER(b.title) LIKE ? OR LOWER(b.author) LIKE ?)",
				"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%")
		}
	}

	if len(req.BookIDs) > 0 {
		query = query.Where("a.book_id IN ?", req.BookIDs)
	}
	if len(req.Types) > 0 {
		query = query.Where("a.type IN ?", req.Types)
	}
	if len(req.Authors) > 0 {
		query = query.Where("b.author IN ?", req.Authors)
	}
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("? = ANY(a.tags)", tag)
		}
	}
	if len(req.Colors) > 0 {
		query = query.Where("a.color IN ?", req.Colors)
	}
	if req.IsPrivate != nil {
		query = query.Where("a.is_private = ?", *req.IsPrivate)
	}
	if req.MinPage != nil {
		query = query.Where("a.page_number >= ?", *req.MinPage)
	}
	if req.MaxPage != nil {
		query = query.Where("a.page_number <= ?", *req.MaxPage)
	}
	if req.DateFrom != "" {
		query = query.Where("a.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("a.created_at <= ?", req.DateTo)
	}

	// Get total count
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// Apply sorting
	switch req.SortBy {
	case "book":
		if req.SortOrder == "asc" {
			query = query.Order("b.title ASC, a.page_number ASC")
		} else {
			query = query.Order("b.title DESC, a.page_number DESC")
		}
	case "page":
		if req.SortOrder == "asc" {
			query = query.Order("a.page_number ASC")
		} else {
			query = query.Order("a.page_number DESC")
		}
	case "type":
		if req.SortOrder == "asc" {
			query = query.Order("a.type ASC, a.created_at DESC")
		} else {
			query = query.Order("a.type DESC, a.created_at DESC")
		}
	default: // date
		if req.SortOrder == "asc" {
			query = query.Order("a.created_at ASC")
		} else {
			query = query.Order("a.created_at DESC")
		}
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	query = query.Offset(offset).Limit(req.PerPage)

	type AnnotationQueryResult struct {
		ID           string    `json:"id"`
		BookID       string    `json:"book_id"`
		Type         string    `json:"type"`
		PageNumber   int       `json:"page_number"`
		StartPos     int       `json:"start_position"`
		EndPos       int       `json:"end_position"`
		SelectedText string    `json:"selected_text"`
		Content      string    `json:"content"`
		Color        string    `json:"color"`
		Tags         []string  `json:"tags"`
		IsPrivate    bool      `json:"is_private"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		BookTitle    string    `json:"book_title"`
		BookAuthor   string    `json:"book_author"`
		BookCover    string    `json:"book_cover"`
	}

	var queryResults []AnnotationQueryResult
	if err := query.Find(&queryResults).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// Convert to response format
	annotations := make([]EnhancedAnnotationResponse, len(queryResults))
	for i, result := range queryResults {
		bookID, _ := uuid.Parse(result.BookID)
		id, _ := uuid.Parse(result.ID)

		annotations[i] = EnhancedAnnotationResponse{
			ID:           id,
			BookID:       bookID,
			BookTitle:    result.BookTitle,
			BookAuthor:   result.BookAuthor,
			BookCover:    result.BookCover,
			Type:         result.Type,
			PageNumber:   result.PageNumber,
			StartPos:     result.StartPos,
			EndPos:       result.EndPos,
			SelectedText: result.SelectedText,
			Content:      result.Content,
			Color:        result.Color,
			Tags:         result.Tags,
			IsPrivate:    result.IsPrivate,
			CreatedAt:    result.CreatedAt,
			UpdatedAt:    result.UpdatedAt,
		}
	}

	// Build pagination response
	totalPages := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))
	pagination := &PaginationResponse{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	// Get filters
	filters, err := getAnnotationFilters(database, userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get stats if requested
	var stats *AnnotationStatsResponse
	if req.IncludeStats {
		stats, err = getAnnotationStats(database, userID)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	return annotations, pagination, filters, stats, nil
}

func getAnnotationFilters(database *gorm.DB, userID uuid.UUID) (*AnnotationFiltersResponse, error) {
	filters := &AnnotationFiltersResponse{
		AvailableTypes: []string{"highlight", "note", "bookmark"},
	}

	// Get available books
	var books []BookFilterOption
	database.Table("user_books ub").
		Select("b.id, b.title, b.author").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Order("b.title").
		Find(&books)
	filters.AvailableBooks = books

	// Get available authors
	var authors []string
	database.Table("user_books ub").
		Select("DISTINCT b.author").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Order("b.author").
		Pluck("author", &authors)
	filters.AvailableAuthors = authors

	// Get available tags
	var tags []string
	database.Raw(`
		SELECT DISTINCT unnest(tags) as tag 
		FROM annotations 
		WHERE user_id = ? AND deleted_at IS NULL AND array_length(tags, 1) > 0
		ORDER BY tag
	`, userID).Pluck("tag", &tags)
	filters.AvailableTags = tags

	// Get available colors
	var colors []string
	database.Table("annotations").
		Select("DISTINCT color").
		Where("user_id = ? AND deleted_at IS NULL AND color != ''", userID).
		Order("color").
		Pluck("color", &colors)
	filters.AvailableColors = colors

	return filters, nil
}

func getAnnotationStats(database *gorm.DB, userID uuid.UUID) (*AnnotationStatsResponse, error) {
	stats := &AnnotationStatsResponse{}

	// Get basic counts
	database.Model(&models.Annotation{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&stats.TotalAnnotations)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'highlight' AND deleted_at IS NULL", userID).Count(&stats.TotalHighlights)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'note' AND deleted_at IS NULL", userID).Count(&stats.TotalNotes)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'bookmark' AND deleted_at IS NULL", userID).Count(&stats.TotalBookmarks)

	// Books with annotations
	database.Table("annotations").
		Select("COUNT(DISTINCT book_id)").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Scan(&stats.BooksWithAnnotations)

	// Most used tags
	type TagCount struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}

	var tagCounts []TagCount
	database.Raw(`
		SELECT unnest(tags) as tag, COUNT(*) as count
		FROM annotations 
		WHERE user_id = ? AND deleted_at IS NULL AND array_length(tags, 1) > 0
		GROUP BY unnest(tags)
		ORDER BY count DESC
		LIMIT 10
	`, userID).Find(&tagCounts)

	for _, tc := range tagCounts {
		stats.MostUsedTags = append(stats.MostUsedTags, TagUsage{Tag: tc.Tag, Count: tc.Count})
	}

	// Annotations by month (last 12 months)
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		var total, highlights, notes, bookmarks int64

		database.Model(&models.Annotation{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Count(&total)

		database.Model(&models.Annotation{}).
			Where("user_id = ? AND type = 'highlight' AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Count(&highlights)

		database.Model(&models.Annotation{}).
			Where("user_id = ? AND type = 'note' AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Count(&notes)

		database.Model(&models.Annotation{}).
			Where("user_id = ? AND type = 'bookmark' AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Count(&bookmarks)

		stats.AnnotationsByMonth = append(stats.AnnotationsByMonth, MonthlyAnnotations{
			Month:      monthStart.Format("January 2006"),
			Count:      int(total),
			Highlights: int(highlights),
			Notes:      int(notes),
			Bookmarks:  int(bookmarks),
		})
	}

	return stats, nil
}

func performBulkAction(database *gorm.DB, userID uuid.UUID, req BulkActionRequest) (*BulkActionResponse, error) {
	response := &BulkActionResponse{
		UpdatedAt: time.Now(),
	}

	// Verify all annotations belong to the user
	var count int64
	database.Model(&models.Annotation{}).
		Where("id IN ? AND user_id = ? AND deleted_at IS NULL", req.AnnotationIDs, userID).
		Count(&count)

	if int(count) != len(req.AnnotationIDs) {
		return nil, fmt.Errorf("some annotations not found or not accessible")
	}

	switch req.Action {
	case "delete":
		result := database.Where("id IN ? AND user_id = ?", req.AnnotationIDs, userID).Delete(&models.Annotation{})
		if result.Error != nil {
			return nil, result.Error
		}
		response.Success = int(result.RowsAffected)

	case "update_tags":
		tags, ok := req.Parameters["tags"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tags parameter")
		}
		
		var tagStrings []string
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				tagStrings = append(tagStrings, tagStr)
			}
		}

		result := database.Model(&models.Annotation{}).
			Where("id IN ? AND user_id = ?", req.AnnotationIDs, userID).
			Update("tags", tagStrings)
		if result.Error != nil {
			return nil, result.Error
		}
		response.Success = int(result.RowsAffected)

	case "update_color":
		color, ok := req.Parameters["color"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid color parameter")
		}

		result := database.Model(&models.Annotation{}).
			Where("id IN ? AND user_id = ?", req.AnnotationIDs, userID).
			Update("color", color)
		if result.Error != nil {
			return nil, result.Error
		}
		response.Success = int(result.RowsAffected)

	case "toggle_private":
		// This requires individual updates since we need to toggle each one
		for _, annotationID := range req.AnnotationIDs {
			var annotation models.Annotation
			if err := database.Where("id = ? AND user_id = ?", annotationID, userID).First(&annotation).Error; err != nil {
				response.Failed++
				response.Errors = append(response.Errors, fmt.Sprintf("Failed to find annotation %s", annotationID))
				continue
			}

			if err := database.Model(&annotation).Update("is_private", !annotation.IsPrivate).Error; err != nil {
				response.Failed++
				response.Errors = append(response.Errors, fmt.Sprintf("Failed to update annotation %s", annotationID))
				continue
			}

			response.Success++
		}

	default:
		return nil, fmt.Errorf("unsupported bulk action: %s", req.Action)
	}

	return response, nil
}

func convertToEnhancedResponse(annotation *models.Annotation) EnhancedAnnotationResponse {
	return EnhancedAnnotationResponse{
		ID:           annotation.ID,
		BookID:       annotation.BookID,
		BookTitle:    annotation.Book.Title,
		BookAuthor:   annotation.Book.Author,
		BookCover:    annotation.Book.CoverURL,
		Type:         annotation.Type,
		PageNumber:   annotation.PageNumber,
		StartPos:     annotation.StartPosition,
		EndPos:       annotation.EndPosition,
		SelectedText: annotation.SelectedText,
		Content:      annotation.Content,
		Color:        annotation.Color,
		Tags:         annotation.Tags,
		IsPrivate:    annotation.IsPrivate,
		CreatedAt:    annotation.CreatedAt,
		UpdatedAt:    annotation.UpdatedAt,
	}
}

func exportAnnotationsCSV(c *gin.Context, annotations []ExportAnnotation) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="annotations.csv"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Book Title", "Author", "Type", "Page", "Selected Text", "Content", "Color", "Tags", "Created At"})

	// Write data
	for _, annotation := range annotations {
		tags := strings.Join(annotation.Tags, ", ")
		writer.Write([]string{
			annotation.BookTitle,
			annotation.BookAuthor,
			annotation.Type,
			strconv.Itoa(annotation.PageNumber),
			annotation.SelectedText,
			annotation.Content,
			annotation.Color,
			tags,
			annotation.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func exportAnnotationsJSON(c *gin.Context, annotations []ExportAnnotation) {
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", `attachment; filename="annotations.json"`)

	utils.SuccessResponse(c, http.StatusOK, "Annotations exported successfully", map[string]interface{}{
		"annotations": annotations,
		"exported_at": time.Now(),
		"count":       len(annotations),
	})
}