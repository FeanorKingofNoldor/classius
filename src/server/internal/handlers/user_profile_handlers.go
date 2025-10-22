package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/models"
	"github.com/classius/server/internal/utils"
)

// UserProfileResponse represents user profile information
type UserProfileResponse struct {
	ID               uuid.UUID              `json:"id"`
	Username         string                 `json:"username"`
	Email            string                 `json:"email"`
	FullName         string                 `json:"full_name"`
	AvatarURL        string                 `json:"avatar_url"`
	SubscriptionTier string                 `json:"subscription_tier"`
	LastActive       *time.Time             `json:"last_active"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Stats            UserProfileStats       `json:"stats"`
	Preferences      UserPreferences        `json:"preferences"`
	ReadingGoals     UserReadingGoals       `json:"reading_goals"`
}

// UserProfileStats represents user statistics in profile
type UserProfileStats struct {
	TotalBooks          int     `json:"total_books"`
	BooksRead           int     `json:"books_read"`
	BooksInProgress     int     `json:"books_in_progress"`
	TotalPagesRead      int     `json:"total_pages_read"`
	TotalReadingTime    int     `json:"total_reading_time_minutes"`
	CurrentStreak       int     `json:"current_streak_days"`
	LongestStreak       int     `json:"longest_streak_days"`
	TotalAnnotations    int     `json:"total_annotations"`
	FavoriteGenres      []GenrePreference `json:"favorite_genres"`
	ReadingSpeedPPH     float64 `json:"reading_speed_pages_per_hour"`
}

// GenrePreference represents reading preference by genre
type GenrePreference struct {
	Genre     string  `json:"genre"`
	BookCount int     `json:"book_count"`
	Percentage float64 `json:"percentage"`
}

// UserPreferences represents user reading preferences
type UserPreferences struct {
	Theme                string   `json:"theme"`                  // light, dark, auto
	Language             string   `json:"language"`               // en, es, fr, etc.
	TimeZone             string   `json:"timezone"`               // UTC offset
	NotificationSettings NotificationSettings `json:"notification_settings"`
	ReadingSettings      ReadingSettings      `json:"reading_settings"`
	PrivacySettings      PrivacySettings      `json:"privacy_settings"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailNotifications    bool `json:"email_notifications"`
	ReadingReminders     bool `json:"reading_reminders"`
	GoalReminders        bool `json:"goal_reminders"`
	NewBooksNotification bool `json:"new_books_notification"`
	WeeklyProgress       bool `json:"weekly_progress"`
}

// ReadingSettings represents reading behavior preferences
type ReadingSettings struct {
	DefaultFontSize      int    `json:"default_font_size"`       // 12-24
	DefaultLineHeight    float64 `json:"default_line_height"`    // 1.0-2.0
	DefaultMargin        int    `json:"default_margin"`         // pixels
	PreferredFormat      string `json:"preferred_format"`       // epub, pdf, etc
	AutoSaveProgress     bool   `json:"auto_save_progress"`
	AutoSyncAnnotations  bool   `json:"auto_sync_annotations"`
	DefaultAnnotationColor string `json:"default_annotation_color"`
}

// PrivacySettings represents privacy preferences
type PrivacySettings struct {
	ProfileVisibility    string `json:"profile_visibility"`     // public, friends, private
	AnnotationsPublic    bool   `json:"annotations_public"`
	ReadingStatsPublic   bool   `json:"reading_stats_public"`
	AllowDataExport      bool   `json:"allow_data_export"`
	ShareReadingActivity bool   `json:"share_reading_activity"`
}

// UserReadingGoals represents user reading goals
type UserReadingGoals struct {
	YearlyBooksGoal     int     `json:"yearly_books_goal"`
	MonthlyBooksGoal    int     `json:"monthly_books_goal"`
	DailyPagesGoal      int     `json:"daily_pages_goal"`
	DailyMinutesGoal    int     `json:"daily_minutes_goal"`
	CurrentYearProgress float64 `json:"current_year_progress"`
	CurrentMonthProgress float64 `json:"current_month_progress"`
	GoalsEnabled        bool    `json:"goals_enabled"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// UpdatePreferencesRequest represents preferences update request
type UpdatePreferencesRequest struct {
	Theme                string                `json:"theme,omitempty"`
	Language             string                `json:"language,omitempty"`
	TimeZone             string                `json:"timezone,omitempty"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
	ReadingSettings      *ReadingSettings      `json:"reading_settings,omitempty"`
	PrivacySettings      *PrivacySettings      `json:"privacy_settings,omitempty"`
}

// UpdateGoalsRequest represents reading goals update request
type UpdateGoalsRequest struct {
	YearlyBooksGoal  int  `json:"yearly_books_goal,omitempty"`
	MonthlyBooksGoal int  `json:"monthly_books_goal,omitempty"`
	DailyPagesGoal   int  `json:"daily_pages_goal,omitempty"`
	DailyMinutesGoal int  `json:"daily_minutes_goal,omitempty"`
	GoalsEnabled     *bool `json:"goals_enabled,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// AccountStatsResponse represents comprehensive account statistics
type AccountStatsResponse struct {
	JoinDate          time.Time            `json:"join_date"`
	DaysActive        int                  `json:"days_active"`
	LibraryStats      LibraryStats         `json:"library_stats"`
	ReadingStats      DetailedReadingStats `json:"reading_stats"`
	AnnotationStats   DetailedAnnotationStats `json:"annotation_stats"`
	DeviceStats       []DeviceUsage        `json:"device_stats"`
	MonthlyActivity   []MonthlyActivity    `json:"monthly_activity"`
}

// LibraryStats represents library statistics
type LibraryStats struct {
	TotalBooks      int                `json:"total_books"`
	TotalSizeMB     float64            `json:"total_size_mb"`
	BooksByFormat   map[string]int     `json:"books_by_format"`
	BooksByLanguage map[string]int     `json:"books_by_language"`
	BooksByGenre    map[string]int     `json:"books_by_genre"`
	RecentlyAdded   []RecentBook       `json:"recently_added"`
}

// DetailedReadingStats represents detailed reading statistics
type DetailedReadingStats struct {
	TotalReadingTime    int     `json:"total_reading_time_minutes"`
	TotalPagesRead      int     `json:"total_pages_read"`
	TotalSessions       int     `json:"total_sessions"`
	AverageSessionTime  int     `json:"average_session_time_minutes"`
	BooksCompleted      int     `json:"books_completed"`
	CompletionRate      float64 `json:"completion_rate_percentage"`
	ReadingSpeedPPH     float64 `json:"reading_speed_pages_per_hour"`
	FavoriteReadingTime string  `json:"favorite_reading_time"`
	LongestSession      int     `json:"longest_session_minutes"`
}

// DetailedAnnotationStats represents annotation statistics
type DetailedAnnotationStats struct {
	TotalAnnotations    int                    `json:"total_annotations"`
	HighlightsCount     int                    `json:"highlights_count"`
	NotesCount          int                    `json:"notes_count"`
	BookmarksCount      int                    `json:"bookmarks_count"`
	AnnotationsPerBook  float64                `json:"annotations_per_book_average"`
	MostUsedColors      []ColorUsage           `json:"most_used_colors"`
	TagUsage            []TagUsageStats        `json:"tag_usage"`
	AnnotationActivity  []AnnotationActivity   `json:"annotation_activity"`
}

// RecentBook represents recently added book info
type RecentBook struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	AddedAt   time.Time `json:"added_at"`
}

// DeviceUsage represents device usage statistics
type DeviceUsage struct {
	DeviceType string `json:"device_type"`
	Sessions   int    `json:"sessions"`
	Minutes    int    `json:"minutes"`
	LastUsed   time.Time `json:"last_used"`
}

// MonthlyActivity represents monthly activity data
type MonthlyActivity struct {
	Month        string `json:"month"`
	BooksAdded   int    `json:"books_added"`
	PagesRead    int    `json:"pages_read"`
	ReadingTime  int    `json:"reading_time_minutes"`
	Annotations  int    `json:"annotations"`
}

// ColorUsage represents color usage in annotations
type ColorUsage struct {
	Color string `json:"color"`
	Count int    `json:"count"`
}

// TagUsageStats represents tag usage statistics
type TagUsageStats struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// AnnotationActivity represents annotation activity over time
type AnnotationActivity struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetUserProfileDetailed returns comprehensive user profile information (renamed to avoid conflict)
func GetUserProfileDetailed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	database := db.DB

	// Get user details
	var user models.User
	if err := database.Where("id = ?", userUUID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user profile", err)
		return
	}

	// Build comprehensive profile response
	profile, err := buildUserProfile(database, &user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to build user profile", err)
		return
	}

	utils.SuccessResponse(c, "User profile retrieved successfully", profile)
}

// UpdateUserProfileDetailed updates user profile information (renamed to avoid conflict)
func UpdateUserProfileDetailed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid profile data: "+err.Error(), err)
		return
	}

	database := db.DB

	// Find user
	var user models.User
	if err := database.Where("id = ?", userUUID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// Build update data
	updateData := make(map[string]interface{})
	if req.FullName != "" {
		updateData["full_name"] = req.FullName
	}
	if req.AvatarURL != "" {
		updateData["avatar_url"] = req.AvatarURL
	}
	updateData["last_active"] = time.Now()

	if len(updateData) > 1 { // More than just last_active
		if err := database.Model(&user).Updates(updateData).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile", err)
			return
		}
	}

	// Build and return updated profile
	profile, err := buildUserProfile(database, &user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Profile updated but failed to load details", err)
		return
	}

	utils.SuccessResponse(c, "Profile updated successfully", profile)
}

// GetUserPreferences returns user preferences
func GetUserPreferences(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// For now, return default preferences since we haven't implemented the preferences table
	// In a full implementation, this would load from a user_preferences table
	preferences := getDefaultPreferences()

	utils.SuccessResponse(c, "Preferences retrieved successfully", preferences)
}

// UpdateUserPreferences updates user preferences
func UpdateUserPreferences(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid preferences data", err)
		return
	}

	// For now, simulate successful update
	// In a full implementation, this would update the user_preferences table
	preferences := getDefaultPreferences()

	// Apply updates from request
	if req.Theme != "" {
		preferences.Theme = req.Theme
	}
	if req.Language != "" {
		preferences.Language = req.Language
	}
	if req.TimeZone != "" {
		preferences.TimeZone = req.TimeZone
	}
	if req.NotificationSettings != nil {
		preferences.NotificationSettings = *req.NotificationSettings
	}
	if req.ReadingSettings != nil {
		preferences.ReadingSettings = *req.ReadingSettings
	}
	if req.PrivacySettings != nil {
		preferences.PrivacySettings = *req.PrivacySettings
	}

	utils.SuccessResponse(c, "Preferences updated successfully", preferences)
}

// GetReadingGoals returns user reading goals
func GetReadingGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	database := db.DB

	// Calculate goal progress
	goals, err := calculateReadingGoals(database, userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate reading goals", err)
		return
	}

	utils.SuccessResponse(c, "Reading goals retrieved successfully", goals)
}

// UpdateReadingGoals updates user reading goals
func UpdateReadingGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req UpdateGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid goals data", err)
		return
	}

	database := db.DB

	// For now, simulate successful update and recalculate
	// In a full implementation, this would update user goals in database
	goals, err := calculateReadingGoals(database, userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update reading goals", err)
		return
	}

	// Apply updates from request
	if req.YearlyBooksGoal > 0 {
		goals.YearlyBooksGoal = req.YearlyBooksGoal
	}
	if req.MonthlyBooksGoal > 0 {
		goals.MonthlyBooksGoal = req.MonthlyBooksGoal
	}
	if req.DailyPagesGoal > 0 {
		goals.DailyPagesGoal = req.DailyPagesGoal
	}
	if req.DailyMinutesGoal > 0 {
		goals.DailyMinutesGoal = req.DailyMinutesGoal
	}
	if req.GoalsEnabled != nil {
		goals.GoalsEnabled = *req.GoalsEnabled
	}

	utils.SuccessResponse(c, "Reading goals updated successfully", goals)
}

// ChangePassword changes user password
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid password data", err)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		utils.ErrorResponse(c, http.StatusBadRequest, "New passwords do not match", nil)
		return
	}

	database := db.DB

	// Find user and verify current password
	var user models.User
	if err := database.Where("id = ? AND deleted_at IS NULL", userUUID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// Check current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Current password is incorrect", nil)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	// Update password
	if err := database.Model(&user).Updates(map[string]interface{}{
		"password_hash": string(hashedPassword),
		"last_active":   time.Now(),
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update password", err)
		return
	}

	utils.SuccessResponse(c, "Password changed successfully", nil)
}

// GetAccountStats returns comprehensive account statistics
func GetAccountStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	database := db.DB

	// Get user
	var user models.User
	if err := database.Where("id = ? AND deleted_at IS NULL", userUUID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// Build comprehensive stats
	stats, err := buildAccountStats(database, &user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate account statistics", err)
		return
	}

	utils.SuccessResponse(c, "Account statistics retrieved successfully", stats)
}

// DeleteAccount handles account deletion (soft delete)
func DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
		Confirm  bool   `json:"confirm" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid deletion data: ", err)
		return
	}

	if !req.Confirm {
		utils.ErrorResponse(c, http.StatusBadRequest, "Account deletion not confirmed", nil)
		return
	}

	database := db.DB

	// Find user and verify password
	var user models.User
	if err := database.Where("id = ? AND deleted_at IS NULL", userUUID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Password is incorrect", nil)
		return
	}

	// Soft delete user account
	if err := database.Delete(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account", err)
		return
	}

	utils.SuccessResponse(c, "Account deleted successfully", nil)
}

// Helper functions

func buildUserProfile(database *gorm.DB, user *models.User) (*UserProfileResponse, error) {
	// Get user stats
	stats, err := buildUserStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	// Get preferences
	preferences := getDefaultPreferences()

	// Get reading goals
	goals, err := calculateReadingGoals(database, user.ID)
	if err != nil {
		return nil, err
	}

	profile := &UserProfileResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FullName:         user.FullName,
		AvatarURL:        user.AvatarURL,
		SubscriptionTier: user.SubscriptionTier,
		LastActive:       user.LastActive,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Stats:            *stats,
		Preferences:      preferences,
		ReadingGoals:     *goals,
	}

	return profile, nil
}

func buildUserStats(database *gorm.DB, userID uuid.UUID) (*UserProfileStats, error) {
	var totalBooks, booksRead, booksInProgress, totalPagesRead, totalReadingTime, currentStreak, longestStreak, totalAnnotations int64

	// Get basic counts
	database.Model(&models.UserBook{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&totalBooks)

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND percentage >= 100 AND deleted_at IS NULL", userID).Count(&booksRead)

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND percentage > 0 AND percentage < 100 AND deleted_at IS NULL", userID).Count(&booksInProgress)

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(current_page), 0)").Scan(&totalPagesRead)

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&totalReadingTime)

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("MAX(reading_streak_days)").Scan(&longestStreak)

	database.Model(&models.Annotation{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&totalAnnotations)

	// Calculate reading speed
	var readingSpeed float64
	if totalReadingTime > 0 {
		readingSpeed = float64(totalPagesRead) / (float64(totalReadingTime) / 60.0)
	}

	// Get favorite genres (top 3)
	type genreCount struct {
		Genre string `json:"genre"`
		Count int64  `json:"count"`
	}

	var genreCounts []genreCount
	database.Table("user_books ub").
		Select("b.genre, COUNT(*) as count").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL AND b.genre != ''", userID).
		Group("b.genre").
		Order("count DESC").
		Limit(3).
		Find(&genreCounts)

	favoriteGenres := make([]GenrePreference, len(genreCounts))
	for i, gc := range genreCounts {
		percentage := float64(gc.Count) / float64(totalBooks) * 100
		favoriteGenres[i] = GenrePreference{
			Genre:      gc.Genre,
			BookCount:  int(gc.Count),
			Percentage: percentage,
		}
	}

	// Simplified current streak calculation
	currentStreak = longestStreak

	return &UserProfileStats{
		TotalBooks:          int(totalBooks),
		BooksRead:           int(booksRead),
		BooksInProgress:     int(booksInProgress),
		TotalPagesRead:      int(totalPagesRead),
		TotalReadingTime:    int(totalReadingTime),
		CurrentStreak:       int(currentStreak),
		LongestStreak:       int(longestStreak),
		TotalAnnotations:    int(totalAnnotations),
		FavoriteGenres:      favoriteGenres,
		ReadingSpeedPPH:     readingSpeed,
	}, nil
}

func getDefaultPreferences() UserPreferences {
	return UserPreferences{
		Theme:    "light",
		Language: "en",
		TimeZone: "UTC",
		NotificationSettings: NotificationSettings{
			EmailNotifications:    true,
			ReadingReminders:     true,
			GoalReminders:        true,
			NewBooksNotification: false,
			WeeklyProgress:       true,
		},
		ReadingSettings: ReadingSettings{
			DefaultFontSize:          16,
			DefaultLineHeight:        1.5,
			DefaultMargin:           20,
			PreferredFormat:         "epub",
			AutoSaveProgress:        true,
			AutoSyncAnnotations:     true,
			DefaultAnnotationColor:  "#ffff00",
		},
		PrivacySettings: PrivacySettings{
			ProfileVisibility:     "public",
			AnnotationsPublic:     false,
			ReadingStatsPublic:    true,
			AllowDataExport:       true,
			ShareReadingActivity:  false,
		},
	}
}

func calculateReadingGoals(database *gorm.DB, userID uuid.UUID) (*UserReadingGoals, error) {
	// Get current year progress
	yearStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	var yearlyBooksRead int64
	var monthlyBooksRead int64

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND percentage >= 100 AND updated_at >= ? AND deleted_at IS NULL", userID, yearStart).
		Count(&yearlyBooksRead)

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND percentage >= 100 AND updated_at >= ? AND deleted_at IS NULL", userID, monthStart).
		Count(&monthlyBooksRead)

	// Default goals
	yearlyGoal := 24
	monthlyGoal := 2

	// Calculate progress
	yearlyProgress := float64(yearlyBooksRead) / float64(yearlyGoal) * 100
	monthlyProgress := float64(monthlyBooksRead) / float64(monthlyGoal) * 100

	if yearlyProgress > 100 {
		yearlyProgress = 100
	}
	if monthlyProgress > 100 {
		monthlyProgress = 100
	}

	return &UserReadingGoals{
		YearlyBooksGoal:      yearlyGoal,
		MonthlyBooksGoal:     monthlyGoal,
		DailyPagesGoal:       25,
		DailyMinutesGoal:     30,
		CurrentYearProgress:  yearlyProgress,
		CurrentMonthProgress: monthlyProgress,
		GoalsEnabled:         true,
	}, nil
}

func buildAccountStats(database *gorm.DB, user *models.User) (*AccountStatsResponse, error) {
	// Calculate days since joining
	daysActive := int(time.Since(user.CreatedAt).Hours() / 24)

	// Build library stats
	libraryStats, err := buildLibraryStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	// Build reading stats
	readingStats, err := buildDetailedReadingStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	// Build annotation stats
	annotationStats, err := buildDetailedAnnotationStats(database, user.ID)
	if err != nil {
		return nil, err
	}

	// Get monthly activity (last 12 months)
	monthlyActivity := []MonthlyActivity{}
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		var booksAdded, pagesRead, readingTime, annotations int64

		database.Model(&models.UserBook{}).Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", user.ID, monthStart, monthEnd).Count(&booksAdded)

		database.Table("reading_progress").Where("user_id = ? AND updated_at >= ? AND updated_at <= ? AND deleted_at IS NULL", user.ID, monthStart, monthEnd).Select("COALESCE(SUM(current_page), 0)").Scan(&pagesRead)

		database.Table("reading_progress").Where("user_id = ? AND updated_at >= ? AND updated_at <= ? AND deleted_at IS NULL", user.ID, monthStart, monthEnd).Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&readingTime)

		database.Model(&models.Annotation{}).Where("user_id = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL", user.ID, monthStart, monthEnd).Count(&annotations)

		monthlyActivity = append(monthlyActivity, MonthlyActivity{
			Month:       monthStart.Format("January 2006"),
			BooksAdded:  int(booksAdded),
			PagesRead:   int(pagesRead),
			ReadingTime: int(readingTime),
			Annotations: int(annotations),
		})
	}

	return &AccountStatsResponse{
		JoinDate:        user.CreatedAt,
		DaysActive:      daysActive,
		LibraryStats:    *libraryStats,
		ReadingStats:    *readingStats,
		AnnotationStats: *annotationStats,
		DeviceStats:     []DeviceUsage{}, // Would be populated from reading_sessions
		MonthlyActivity: monthlyActivity,
	}, nil
}

func buildLibraryStats(database *gorm.DB, userID uuid.UUID) (*LibraryStats, error) {
	var totalBooks int64
	var totalSize int64

	database.Model(&models.UserBook{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&totalBooks)

	database.Table("user_books ub").
		Select("COALESCE(SUM(b.file_size), 0)").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Scan(&totalSize)

	// Get format distribution
	formatCounts := make(map[string]int)
	database.Table("user_books ub").
		Select("b.file_type, COUNT(*)").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Group("b.file_type").
		Pluck("file_type", &formatCounts)

	// Get recently added books
	var recentBooks []RecentBook
	database.Table("user_books ub").
		Select("b.id, b.title, b.author, ub.created_at as added_at").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Order("ub.created_at DESC").
		Limit(5).
		Find(&recentBooks)

	return &LibraryStats{
		TotalBooks:      int(totalBooks),
		TotalSizeMB:     float64(totalSize) / (1024 * 1024),
		BooksByFormat:   formatCounts,
		BooksByLanguage: make(map[string]int),
		BooksByGenre:    make(map[string]int),
		RecentlyAdded:   recentBooks,
	}, nil
}

func buildDetailedReadingStats(database *gorm.DB, userID uuid.UUID) (*DetailedReadingStats, error) {
	var totalTime, totalPages, totalSessions, booksCompleted int64

	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&totalTime)
	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID).Select("COALESCE(SUM(current_page), 0)").Scan(&totalPages)
	database.Model(&models.ReadingProgress{}).Where("user_id = ? AND percentage >= 100 AND deleted_at IS NULL", userID).Count(&booksCompleted)
	database.Model(&models.ReadingSession{}).Where("user_id = ?", userID).Count(&totalSessions)

	var averageSession int64
	if totalSessions > 0 {
		database.Model(&models.ReadingSession{}).Where("user_id = ?", userID).Select("AVG(duration_minutes)").Scan(&averageSession)
	}

	var readingSpeed float64
	if totalTime > 0 {
		readingSpeed = float64(totalPages) / (float64(totalTime) / 60.0)
	}

	var totalBooksInLibrary int64
	database.Model(&models.UserBook{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&totalBooksInLibrary)

	var completionRate float64
	if totalBooksInLibrary > 0 {
		completionRate = float64(booksCompleted) / float64(totalBooksInLibrary) * 100
	}

	return &DetailedReadingStats{
		TotalReadingTime:    int(totalTime),
		TotalPagesRead:      int(totalPages),
		TotalSessions:       int(totalSessions),
		AverageSessionTime:  int(averageSession),
		BooksCompleted:      int(booksCompleted),
		CompletionRate:      completionRate,
		ReadingSpeedPPH:     readingSpeed,
		FavoriteReadingTime: "Evening", // Placeholder
		LongestSession:      0, // Would be calculated from reading_sessions
	}, nil
}

func buildDetailedAnnotationStats(database *gorm.DB, userID uuid.UUID) (*DetailedAnnotationStats, error) {
	var total, highlights, notes, bookmarks int64

	database.Model(&models.Annotation{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'highlight' AND deleted_at IS NULL", userID).Count(&highlights)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'note' AND deleted_at IS NULL", userID).Count(&notes)
	database.Model(&models.Annotation{}).Where("user_id = ? AND type = 'bookmark' AND deleted_at IS NULL", userID).Count(&bookmarks)

	var totalBooks int64
	database.Model(&models.UserBook{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&totalBooks)

	var avgPerBook float64
	if totalBooks > 0 {
		avgPerBook = float64(total) / float64(totalBooks)
	}

	return &DetailedAnnotationStats{
		TotalAnnotations:    int(total),
		HighlightsCount:     int(highlights),
		NotesCount:          int(notes),
		BookmarksCount:      int(bookmarks),
		AnnotationsPerBook:  avgPerBook,
		MostUsedColors:      []ColorUsage{},
		TagUsage:            []TagUsageStats{},
		AnnotationActivity:  []AnnotationActivity{},
	}, nil
}