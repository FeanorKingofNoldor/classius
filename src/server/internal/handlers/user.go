package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/middleware"
	"github.com/classius/server/internal/models"
)

// UpdateUserProfileRequest represents the request to update user profile
type UpdateUserProfileRequest struct {
	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// UserProgressResponse represents reading progress data
type UserProgressResponse struct {
	BookID            uuid.UUID `json:"book_id"`
	BookTitle         string    `json:"book_title"`
	BookAuthor        string    `json:"book_author"`
	CurrentPage       int       `json:"current_page"`
	TotalPages        int       `json:"total_pages"`
	Percentage        float64   `json:"percentage"`
	TimeSpentMinutes  int       `json:"time_spent_minutes"`
	LastRead          string    `json:"last_read"`
	ReadingStreakDays int       `json:"reading_streak_days"`
	NotesCount        int       `json:"notes_count"`
	HighlightsCount   int       `json:"highlights_count"`
}

// SaveProgressRequest represents the request to save reading progress
type SaveProgressRequest struct {
	BookID           uuid.UUID `json:"book_id" binding:"required"`
	CurrentPage      int       `json:"current_page"`
	TotalPages       int       `json:"total_pages"`
	CurrentPosition  int       `json:"current_position"`
	Percentage       float64   `json:"percentage"`
	TimeSpentMinutes int       `json:"time_spent_minutes"`
}

// GetUserProfile returns the current user's profile information
func GetUserProfile(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to access your profile",
		})
		return
	}

	// Load additional user data if needed
	if err := db.DB.Preload("UserBooks.Book").Preload("ReadingProgress").First(user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load user profile",
			"message": "Unable to retrieve user information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile retrieved successfully",
		"data":    ToUserResponse(user),
	})
}

// UpdateUserProfile updates the current user's profile information
func UpdateUserProfile(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to update your profile",
		})
		return
	}

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Update user fields
	updateData := make(map[string]interface{})
	if req.FullName != "" {
		updateData["full_name"] = req.FullName
	}
	if req.AvatarURL != "" {
		updateData["avatar_url"] = req.AvatarURL
	}

	if len(updateData) > 0 {
		if err := db.DB.Model(user).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update profile",
				"message": "Unable to save profile changes",
			})
			return
		}
	}

	// Fetch updated user data
	if err := db.DB.First(user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load updated profile",
			"message": "Profile updated but unable to retrieve updated information",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    ToUserResponse(user),
	})
}

// GetUserProgress returns the user's reading progress for all books
func GetUserProgress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to access your reading progress",
		})
		return
	}

	// Get query parameters for pagination and filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	bookID := c.Query("book_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := db.DB.Preload("Book").Where("user_id = ?", user.ID)
	
	if bookID != "" {
		query = query.Where("book_id = ?", bookID)
	}

	var progressRecords []models.ReadingProgress
	var total int64

	// Count total records
	query.Model(&models.ReadingProgress{}).Count(&total)

	// Get paginated records
	if err := query.Offset(offset).Limit(limit).Order("last_read DESC").Find(&progressRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load reading progress",
			"message": "Unable to retrieve reading progress data",
		})
		return
	}

	// Convert to response format
	var progressData []UserProgressResponse
	for _, progress := range progressRecords {
		responseItem := UserProgressResponse{
			BookID:            progress.BookID,
			CurrentPage:       progress.CurrentPage,
			TotalPages:        progress.TotalPages,
			Percentage:        progress.Percentage,
			TimeSpentMinutes:  progress.TimeSpentMinutes,
			LastRead:          progress.LastRead.Format("2006-01-02T15:04:05Z07:00"),
			ReadingStreakDays: progress.ReadingStreakDays,
			NotesCount:        progress.NotesCount,
			HighlightsCount:   progress.HighlightsCount,
		}

		// Add book information if available
		if progress.Book.ID != uuid.Nil {
			responseItem.BookTitle = progress.Book.Title
			responseItem.BookAuthor = progress.Book.Author
		}

		progressData = append(progressData, responseItem)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reading progress retrieved successfully",
		"data": gin.H{
			"progress": progressData,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		},
	})
}

// SaveUserProgress saves or updates reading progress for a book
func SaveUserProgress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "Please log in to save reading progress",
		})
		return
	}

	var req SaveProgressRequest
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

	// Find existing progress or create new
	var progress models.ReadingProgress
	result := db.DB.Where("user_id = ? AND book_id = ?", user.ID, req.BookID).First(&progress)

	if result.Error != nil {
		// Create new progress record
		progress = models.ReadingProgress{
			UserID:           user.ID,
			BookID:           req.BookID,
			CurrentPage:      req.CurrentPage,
			TotalPages:       req.TotalPages,
			CurrentPosition:  req.CurrentPosition,
			Percentage:       req.Percentage,
			TimeSpentMinutes: req.TimeSpentMinutes,
		}

		if err := db.DB.Create(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save progress",
				"message": "Unable to create progress record",
			})
			return
		}
	} else {
		// Update existing progress
		updateData := map[string]interface{}{
			"current_page":       req.CurrentPage,
			"current_position":   req.CurrentPosition,
			"percentage":         req.Percentage,
			"last_read":         db.DB.NowFunc(),
		}

		if req.TotalPages > 0 {
			updateData["total_pages"] = req.TotalPages
		}
		if req.TimeSpentMinutes > progress.TimeSpentMinutes {
			updateData["time_spent_minutes"] = req.TimeSpentMinutes
		}

		if err := db.DB.Model(&progress).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update progress",
				"message": "Unable to save progress changes",
			})
			return
		}
	}

	// Load the updated progress with book information
	if err := db.DB.Preload("Book").First(&progress, progress.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load updated progress",
			"message": "Progress saved but unable to retrieve updated information",
		})
		return
	}

	responseData := UserProgressResponse{
		BookID:            progress.BookID,
		BookTitle:         progress.Book.Title,
		BookAuthor:        progress.Book.Author,
		CurrentPage:       progress.CurrentPage,
		TotalPages:        progress.TotalPages,
		Percentage:        progress.Percentage,
		TimeSpentMinutes:  progress.TimeSpentMinutes,
		LastRead:          progress.LastRead.Format("2006-01-02T15:04:05Z07:00"),
		ReadingStreakDays: progress.ReadingStreakDays,
		NotesCount:        progress.NotesCount,
		HighlightsCount:   progress.HighlightsCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reading progress saved successfully",
		"data":    responseData,
	})
}