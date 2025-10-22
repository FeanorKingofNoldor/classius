package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/models"
	"github.com/classius/server/internal/utils"
)

// ReadingProgressRequest represents a reading progress update request
type ReadingProgressRequest struct {
	BookID           uuid.UUID `json:"book_id" binding:"required"`
	CurrentPage      int       `json:"current_page"`
	TotalPages       int       `json:"total_pages"`
	CurrentPosition  int       `json:"current_position"`
	Percentage       float64   `json:"percentage"`
	TimeSpentMinutes int       `json:"time_spent_minutes"`
	ReadingStreak    int       `json:"reading_streak_days"`
	NotesCount       int       `json:"notes_count"`
	HighlightsCount  int       `json:"highlights_count"`
}

// ReadingSessionRequest represents a reading session start/update request
type ReadingSessionRequest struct {
	BookID       uuid.UUID `json:"book_id" binding:"required"`
	StartPage    int       `json:"start_page"`
	EndPage      int       `json:"end_page"`
	StartPos     int       `json:"start_position"`
	EndPos       int       `json:"end_position"`
	PagesRead    int       `json:"pages_read"`
	DeviceType   string    `json:"device_type"`
}

// BookmarkRequest represents a bookmark creation/update request
type BookmarkRequest struct {
	BookID     uuid.UUID `json:"book_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	PageNumber int       `json:"page_number" binding:"required"`
	Position   int       `json:"position"`
}

// ReadingProgressResponse represents reading progress information
type ReadingProgressResponse struct {
	ID               uuid.UUID `json:"id"`
	BookID           uuid.UUID `json:"book_id"`
	BookTitle        string    `json:"book_title"`
	BookAuthor       string    `json:"book_author"`
	BookCover        string    `json:"book_cover"`
	CurrentPage      int       `json:"current_page"`
	TotalPages       int       `json:"total_pages"`
	CurrentPosition  int       `json:"current_position"`
	Percentage       float64   `json:"percentage"`
	TimeSpentMinutes int       `json:"time_spent_minutes"`
	LastRead         time.Time `json:"last_read"`
	ReadingStreak    int       `json:"reading_streak_days"`
	NotesCount       int       `json:"notes_count"`
	HighlightsCount  int       `json:"highlights_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ReadingSessionResponse represents reading session information
type ReadingSessionResponse struct {
	ID              uuid.UUID `json:"id"`
	BookID          uuid.UUID `json:"book_id"`
	BookTitle       string    `json:"book_title"`
	BookAuthor      string    `json:"book_author"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationMinutes int       `json:"duration_minutes"`
	PagesRead       int       `json:"pages_read"`
	StartPage       int       `json:"start_page"`
	EndPage         int       `json:"end_page"`
	StartPos        int       `json:"start_position"`
	EndPos          int       `json:"end_position"`
	DeviceType      string    `json:"device_type"`
	CreatedAt       time.Time `json:"created_at"`
}

// BookmarkResponse represents bookmark information
type BookmarkResponse struct {
	ID         uuid.UUID `json:"id"`
	BookID     uuid.UUID `json:"book_id"`
	BookTitle  string    `json:"book_title"`
	BookAuthor string    `json:"book_author"`
	Name       string    `json:"name"`
	PageNumber int       `json:"page_number"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ReadingStatsResponse represents comprehensive reading statistics
type ReadingStatsResponse struct {
	TotalReadingTime    int                      `json:"total_reading_time_minutes"`
	TotalPagesRead      int                      `json:"total_pages_read"`
	BooksInProgress     int                      `json:"books_in_progress"`
	BooksCompleted      int                      `json:"books_completed"`
	CurrentStreak       int                      `json:"current_streak_days"`
	LongestStreak       int                      `json:"longest_streak_days"`
	AverageReadingSpeed float64                  `json:"average_reading_speed_pages_per_hour"`
	RecentSessions      []ReadingSessionResponse `json:"recent_sessions"`
	ReadingGoals        ReadingGoalsResponse     `json:"reading_goals"`
}

// ReadingGoalsResponse represents reading goals information
type ReadingGoalsResponse struct {
	DailyPagesGoal    int     `json:"daily_pages_goal"`
	WeeklyPagesGoal   int     `json:"weekly_pages_goal"`
	MonthlyBooksGoal  int     `json:"monthly_books_goal"`
	YearlyBooksGoal   int     `json:"yearly_books_goal"`
	DailyProgress     float64 `json:"daily_progress_percentage"`
	WeeklyProgress    float64 `json:"weekly_progress_percentage"`
	MonthlyProgress   float64 `json:"monthly_progress_percentage"`
	YearlyProgress    float64 `json:"yearly_progress_percentage"`
}

// GetReadingProgress gets reading progress for a specific book or all books
func GetReadingProgress(c *gin.Context) {
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

	bookID := c.Query("book_id")
	database := db.GetDB()

	if bookID != "" {
		// Get progress for specific book
		bookUUID, err := uuid.Parse(bookID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID")
			return
		}

		var progress models.ReadingProgress
		if err := database.Preload("Book").Where("user_id = ? AND book_id = ? AND deleted_at IS NULL", userUUID, bookUUID).
			First(&progress).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusNotFound, "Reading progress not found")
				return
			}
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reading progress")
			return
		}

		response := convertProgressToResponse(&progress)
		utils.SuccessResponse(c, http.StatusOK, "Reading progress retrieved successfully", response)
		return
	}

	// Get progress for all books
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	var progresses []models.ReadingProgress
	var total int64

	query := database.Preload("Book").Where("user_id = ? AND deleted_at IS NULL", userUUID)
	
	// Count total
	query.Model(&models.ReadingProgress{}).Count(&total)

	// Get paginated results
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("last_read DESC").Find(&progresses).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reading progress")
		return
	}

	responses := make([]ReadingProgressResponse, len(progresses))
	for i, progress := range progresses {
		responses[i] = convertProgressToResponse(&progress)
	}

	totalPages := (int(total) + perPage - 1) / perPage
	
	utils.SuccessResponse(c, http.StatusOK, "Reading progress retrieved successfully", map[string]interface{}{
		"progress": responses,
		"pagination": map[string]interface{}{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// UpdateReadingProgress updates reading progress for a book
func UpdateReadingProgress(c *gin.Context) {
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

	var req ReadingProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid progress data: "+err.Error())
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

	// Update or create reading progress
	var progress models.ReadingProgress
	err = database.Where("user_id = ? AND book_id = ?", userUUID, req.BookID).First(&progress).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new progress record
		progress = models.ReadingProgress{
			UserID:            userUUID,
			BookID:            req.BookID,
			CurrentPage:       req.CurrentPage,
			TotalPages:        req.TotalPages,
			CurrentPosition:   req.CurrentPosition,
			Percentage:        req.Percentage,
			TimeSpentMinutes:  req.TimeSpentMinutes,
			LastRead:          time.Now(),
			ReadingStreakDays: req.ReadingStreak,
			NotesCount:        req.NotesCount,
			HighlightsCount:   req.HighlightsCount,
		}

		if err := database.Create(&progress).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create reading progress")
			return
		}
	} else if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reading progress")
		return
	} else {
		// Update existing progress
		updateData := map[string]interface{}{
			"current_page":        req.CurrentPage,
			"total_pages":        req.TotalPages,
			"current_position":   req.CurrentPosition,
			"percentage":         req.Percentage,
			"time_spent_minutes": req.TimeSpentMinutes,
			"last_read":          time.Now(),
			"reading_streak_days": req.ReadingStreak,
			"notes_count":        req.NotesCount,
			"highlights_count":   req.HighlightsCount,
		}

		if err := database.Model(&progress).Updates(updateData).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update reading progress")
			return
		}
	}

	// Reload with book details
	if err := database.Preload("Book").First(&progress, progress.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Progress updated but failed to load details")
		return
	}

	response := convertProgressToResponse(&progress)
	utils.SuccessResponse(c, http.StatusOK, "Reading progress updated successfully", response)
}

// StartReadingSession starts a new reading session
func StartReadingSession(c *gin.Context) {
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

	bookIDStr := c.Param("book_id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID")
		return
	}

	var req struct {
		StartPage    int    `json:"start_page"`
		StartPos     int    `json:"start_position"`
		DeviceType   string `json:"device_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid session data: "+err.Error())
		return
	}

	database := db.GetDB()

	// Verify user has access to the book
	var userBook models.UserBook
	if err := database.Where("user_id = ? AND book_id = ? AND deleted_at IS NULL", userUUID, bookID).
		First(&userBook).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusForbidden, "Book not found in your library")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify book access")
		return
	}

	// Create reading session
	session := models.ReadingSession{
		UserID:       userUUID,
		BookID:       bookID,
		StartedAt:    time.Now(),
		StartPage:    &req.StartPage,
		StartPosition: req.StartPos,
		DeviceType:   &req.DeviceType,
	}

	if err := database.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start reading session")
		return
	}

	// Load session with book details
	if err := database.Preload("Book").First(&session, session.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Session created but failed to load details")
		return
	}

	response := convertSessionToResponse(&session)
	utils.SuccessResponse(c, http.StatusCreated, "Reading session started successfully", response)
}

// EndReadingSession ends a reading session
func EndReadingSession(c *gin.Context) {
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

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req struct {
		EndPage    int `json:"end_page"`
		EndPos     int `json:"end_position"`
		PagesRead  int `json:"pages_read"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid session end data: "+err.Error())
		return
	}

	database := db.GetDB()

	// Find and verify ownership of session
	var session models.ReadingSession
	if err := database.Where("id = ? AND user_id = ?", sessionID, userUUID).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Reading session not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reading session")
		return
	}

	// Update session with end data
	now := time.Now()
	duration := int(now.Sub(session.StartedAt).Minutes())

	updateData := map[string]interface{}{
		"ended_at":        now,
		"duration_minutes": duration,
		"end_page":        req.EndPage,
		"end_position":    req.EndPos,
		"pages_read":      req.PagesRead,
	}

	if err := database.Model(&session).Updates(updateData).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to end reading session")
		return
	}

	// Reload with book details
	if err := database.Preload("Book").First(&session, session.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Session updated but failed to load details")
		return
	}

	response := convertSessionToResponse(&session)
	utils.SuccessResponse(c, http.StatusOK, "Reading session ended successfully", response)
}

// GetReadingSessions gets reading sessions for a user
func GetReadingSessions(c *gin.Context) {
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

	bookID := c.Query("book_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	database := db.GetDB()
	query := database.Preload("Book").Where("user_id = ?", userUUID)

	if bookID != "" {
		bookUUID, err := uuid.Parse(bookID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID")
			return
		}
		query = query.Where("book_id = ?", bookUUID)
	}

	var sessions []models.ReadingSession
	var total int64

	// Count total
	query.Model(&models.ReadingSession{}).Count(&total)

	// Get paginated results
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("started_at DESC").Find(&sessions).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve reading sessions")
		return
	}

	responses := make([]ReadingSessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = convertSessionToResponse(&session)
	}

	totalPages := (int(total) + perPage - 1) / perPage
	
	utils.SuccessResponse(c, http.StatusOK, "Reading sessions retrieved successfully", map[string]interface{}{
		"sessions": responses,
		"pagination": map[string]interface{}{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// CreateBookmark creates a new bookmark
func CreateBookmark(c *gin.Context) {
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

	var req BookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bookmark data: "+err.Error())
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

	// Create bookmark
	bookmark := models.Bookmark{
		UserID:     userUUID,
		BookID:     req.BookID,
		Name:       req.Name,
		PageNumber: req.PageNumber,
		Position:   req.Position,
	}

	if err := database.Create(&bookmark).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create bookmark")
		return
	}

	// Load bookmark with book details
	if err := database.Preload("Book").First(&bookmark, bookmark.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Bookmark created but failed to load details")
		return
	}

	response := convertBookmarkToResponse(&bookmark)
	utils.SuccessResponse(c, http.StatusCreated, "Bookmark created successfully", response)
}

// GetBookmarks gets bookmarks for a user or book
func GetBookmarks(c *gin.Context) {
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

	bookID := c.Query("book_id")
	database := db.GetDB()

	query := database.Preload("Book").Where("user_id = ? AND deleted_at IS NULL", userUUID)

	if bookID != "" {
		bookUUID, err := uuid.Parse(bookID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid book ID")
			return
		}
		query = query.Where("book_id = ?", bookUUID)
	}

	var bookmarks []models.Bookmark
	if err := query.Order("page_number ASC").Find(&bookmarks).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve bookmarks")
		return
	}

	responses := make([]BookmarkResponse, len(bookmarks))
	for i, bookmark := range bookmarks {
		responses[i] = convertBookmarkToResponse(&bookmark)
	}

	utils.SuccessResponse(c, http.StatusOK, "Bookmarks retrieved successfully", responses)
}

// UpdateBookmark updates a bookmark
func UpdateBookmark(c *gin.Context) {
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

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	var req struct {
		Name       string `json:"name"`
		PageNumber int    `json:"page_number"`
		Position   int    `json:"position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bookmark update data: "+err.Error())
		return
	}

	database := db.GetDB()

	// Find and verify ownership
	var bookmark models.Bookmark
	if err := database.Where("id = ? AND user_id = ? AND deleted_at IS NULL", bookmarkID, userUUID).
		First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve bookmark")
		return
	}

	// Update bookmark
	updateData := map[string]interface{}{}
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.PageNumber > 0 {
		updateData["page_number"] = req.PageNumber
	}
	if req.Position >= 0 {
		updateData["position"] = req.Position
	}

	if len(updateData) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No valid updates provided")
		return
	}

	if err := database.Model(&bookmark).Updates(updateData).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update bookmark")
		return
	}

	// Reload with book details
	if err := database.Preload("Book").First(&bookmark, bookmark.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Bookmark updated but failed to load details")
		return
	}

	response := convertBookmarkToResponse(&bookmark)
	utils.SuccessResponse(c, http.StatusOK, "Bookmark updated successfully", response)
}

// DeleteBookmark deletes a bookmark
func DeleteBookmark(c *gin.Context) {
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

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bookmark ID")
		return
	}

	database := db.GetDB()

	// Find and verify ownership
	var bookmark models.Bookmark
	if err := database.Where("id = ? AND user_id = ? AND deleted_at IS NULL", bookmarkID, userUUID).
		First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Bookmark not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve bookmark")
		return
	}

	// Soft delete
	if err := database.Delete(&bookmark).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete bookmark")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Bookmark deleted successfully", nil)
}

// GetReadingStats gets comprehensive reading statistics
func GetReadingStats(c *gin.Context) {
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

	database := db.GetDB()
	
	// Get basic reading stats
	var totalReadingTime int64
	var totalPagesRead int64
	var booksInProgress int64
	var booksCompleted int64
	var currentStreak int64
	var longestStreak int64

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND deleted_at IS NULL", userUUID).
		Select("COALESCE(SUM(time_spent_minutes), 0)").
		Scan(&totalReadingTime)

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND deleted_at IS NULL", userUUID).
		Select("COALESCE(SUM(current_page), 0)").
		Scan(&totalPagesRead)

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND percentage > 0 AND percentage < 100 AND deleted_at IS NULL", userUUID).
		Count(&booksInProgress)

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND percentage >= 100 AND deleted_at IS NULL", userUUID).
		Count(&booksCompleted)

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND deleted_at IS NULL", userUUID).
		Select("MAX(reading_streak_days)").
		Scan(&longestStreak)

	// Calculate current streak (simplified)
	currentStreak = longestStreak

	// Calculate average reading speed
	var averageSpeed float64
	if totalReadingTime > 0 {
		averageSpeed = float64(totalPagesRead) / (float64(totalReadingTime) / 60.0)
	}

	// Get recent sessions
	var sessions []models.ReadingSession
	database.Preload("Book").Where("user_id = ?", userUUID).
		Order("started_at DESC").
		Limit(10).
		Find(&sessions)

	recentSessions := make([]ReadingSessionResponse, len(sessions))
	for i, session := range sessions {
		recentSessions[i] = convertSessionToResponse(&session)
	}

	// Create reading goals (placeholder values)
	goals := ReadingGoalsResponse{
		DailyPagesGoal:   25,
		WeeklyPagesGoal:  175,
		MonthlyBooksGoal: 2,
		YearlyBooksGoal:  24,
		DailyProgress:    75.0,
		WeeklyProgress:   82.5,
		MonthlyProgress:  50.0,
		YearlyProgress:   45.8,
	}

	response := ReadingStatsResponse{
		TotalReadingTime:    int(totalReadingTime),
		TotalPagesRead:      int(totalPagesRead),
		BooksInProgress:     int(booksInProgress),
		BooksCompleted:      int(booksCompleted),
		CurrentStreak:       int(currentStreak),
		LongestStreak:       int(longestStreak),
		AverageReadingSpeed: averageSpeed,
		RecentSessions:      recentSessions,
		ReadingGoals:        goals,
	}

	utils.SuccessResponse(c, http.StatusOK, "Reading statistics retrieved successfully", response)
}

// Helper conversion functions

func convertProgressToResponse(progress *models.ReadingProgress) ReadingProgressResponse {
	return ReadingProgressResponse{
		ID:               progress.ID,
		BookID:           progress.BookID,
		BookTitle:        progress.Book.Title,
		BookAuthor:       progress.Book.Author,
		BookCover:        progress.Book.CoverURL,
		CurrentPage:      progress.CurrentPage,
		TotalPages:       progress.TotalPages,
		CurrentPosition:  progress.CurrentPosition,
		Percentage:       progress.Percentage,
		TimeSpentMinutes: progress.TimeSpentMinutes,
		LastRead:         progress.LastRead,
		ReadingStreak:    progress.ReadingStreakDays,
		NotesCount:       progress.NotesCount,
		HighlightsCount:  progress.HighlightsCount,
		CreatedAt:        progress.CreatedAt,
		UpdatedAt:        progress.UpdatedAt,
	}
}

func convertSessionToResponse(session *models.ReadingSession) ReadingSessionResponse {
	var bookTitle, bookAuthor string
	if session.Book.ID != uuid.Nil {
		bookTitle = session.Book.Title
		bookAuthor = session.Book.Author
	}

	response := ReadingSessionResponse{
		ID:              session.ID,
		BookID:          session.BookID,
		BookTitle:       bookTitle,
		BookAuthor:      bookAuthor,
		StartedAt:       session.StartedAt,
		EndedAt:         session.EndedAt,
		DurationMinutes: session.DurationMinutes,
		PagesRead:       session.PagesRead,
		StartPos:        session.StartPosition,
		EndPos:          session.EndPosition,
		CreatedAt:       session.CreatedAt,
	}

	if session.StartPage != nil {
		response.StartPage = *session.StartPage
	}
	if session.EndPage != nil {
		response.EndPage = *session.EndPage
	}
	if session.DeviceType != nil {
		response.DeviceType = *session.DeviceType
	}

	return response
}

func convertBookmarkToResponse(bookmark *models.Bookmark) BookmarkResponse {
	return BookmarkResponse{
		ID:         bookmark.ID,
		BookID:     bookmark.BookID,
		BookTitle:  bookmark.Book.Title,
		BookAuthor: bookmark.Book.Author,
		Name:       bookmark.Name,
		PageNumber: bookmark.PageNumber,
		Position:   bookmark.Position,
		CreatedAt:  bookmark.CreatedAt,
		UpdatedAt:  bookmark.UpdatedAt,
	}
}