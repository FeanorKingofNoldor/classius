package handlers

import (
	"encoding/json"
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

// BookStatsResponse represents the comprehensive statistics response
type BookStatsResponse struct {
	Overview           StatsOverview          `json:"overview"`
	ReadingHabits      ReadingHabits          `json:"reading_habits"`
	ContentAnalytics   ContentAnalytics       `json:"content_analytics"`
	GoalsAndProgress   GoalsAndProgress       `json:"goals_and_progress"`
	AnnotationsAndEngagement AnnotationsAndEngagement `json:"annotations_and_engagement"`
}

type StatsOverview struct {
	TotalBooks                      int     `json:"total_books"`
	BooksRead                      int     `json:"books_read"`
	BooksInProgress                int     `json:"books_in_progress"`
	TotalPagesRead                 int     `json:"total_pages_read"`
	TotalReadingTimeMinutes        int     `json:"total_reading_time_minutes"`
	AverageReadingSpeedPagesPerHour int     `json:"average_reading_speed_pages_per_hour"`
	CurrentReadingStreakDays       int     `json:"current_reading_streak_days"`
	LongestReadingStreakDays       int     `json:"longest_reading_streak_days"`
}

type ReadingHabits struct {
	DailyAverages      DailyAverages      `json:"daily_averages"`
	MonthlyProgress    []MonthlyProgress  `json:"monthly_progress"`
	ReadingByDayOfWeek []DayOfWeekReading `json:"reading_by_day_of_week"`
	ReadingByHour      []HourlyReading    `json:"reading_by_hour"`
}

type DailyAverages struct {
	PagesPerDay   int `json:"pages_per_day"`
	MinutesPerDay int `json:"minutes_per_day"`
	BooksPerMonth int `json:"books_per_month"`
}

type MonthlyProgress struct {
	Month               string `json:"month"`
	PagesRead          int    `json:"pages_read"`
	BooksCompleted     int    `json:"books_completed"`
	ReadingTimeMinutes int    `json:"reading_time_minutes"`
}

type DayOfWeekReading struct {
	Day       string `json:"day"`
	PagesRead int    `json:"pages_read"`
	Sessions  int    `json:"sessions"`
}

type HourlyReading struct {
	Hour      int `json:"hour"`
	PagesRead int `json:"pages_read"`
	Sessions  int `json:"sessions"`
}

type ContentAnalytics struct {
	Genres      []GenreStats     `json:"genres"`
	Languages   []LanguageStats  `json:"languages"`
	Authors     []AuthorStats    `json:"authors"`
	FileFormats []FileFormatStats `json:"file_formats"`
}

type GenreStats struct {
	Genre          string  `json:"genre"`
	BookCount      int     `json:"book_count"`
	PagesRead      int     `json:"pages_read"`
	CompletionRate float64 `json:"completion_rate"`
}

type LanguageStats struct {
	Language           string `json:"language"`
	BookCount          int    `json:"book_count"`
	PagesRead          int    `json:"pages_read"`
	ReadingTimeMinutes int    `json:"reading_time_minutes"`
}

type AuthorStats struct {
	Author       string  `json:"author"`
	BookCount    int     `json:"book_count"`
	PagesRead    int     `json:"pages_read"`
	FavoriteBook *string `json:"favorite_book,omitempty"`
}

type FileFormatStats struct {
	Format      string  `json:"format"`
	Count       int     `json:"count"`
	TotalSizeMB float64 `json:"total_size_mb"`
}

type GoalsAndProgress struct {
	YearlyGoal        *YearlyGoal        `json:"yearly_goal,omitempty"`
	DailyGoal         *DailyGoal         `json:"daily_goal,omitempty"`
	ReadingMilestones []ReadingMilestone `json:"reading_milestones"`
}

type YearlyGoal struct {
	TargetBooks         int     `json:"target_books"`
	CompletedBooks      int     `json:"completed_books"`
	ProgressPercentage  float64 `json:"progress_percentage"`
	ProjectedCompletion string  `json:"projected_completion"`
}

type DailyGoal struct {
	TargetPages      int     `json:"target_pages"`
	AverageAchieved  int     `json:"average_achieved"`
	SuccessRate      float64 `json:"success_rate"`
}

type ReadingMilestone struct {
	Milestone    string  `json:"milestone"`
	AchievedDate *string `json:"achieved_date,omitempty"`
	Progress     *int    `json:"progress,omitempty"`
}

type AnnotationsAndEngagement struct {
	TotalAnnotations     int                    `json:"total_annotations"`
	Highlights          int                    `json:"highlights"`
	Notes               int                    `json:"notes"`
	Bookmarks           int                    `json:"bookmarks"`
	MostAnnotatedBooks  []MostAnnotatedBook    `json:"most_annotated_books"`
}

type MostAnnotatedBook struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	AnnotationCount int    `json:"annotation_count"`
}

// GetBookStats returns comprehensive reading statistics
func GetBookStats(c *gin.Context) {
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

	// Get time range parameter (week, month, year, all)
	timeRange := c.DefaultQuery("range", "year")

	database := db.GetDB()
	stats, err := buildBookStats(database, userUUID, timeRange)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate statistics")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Statistics retrieved successfully", stats)
}

func buildBookStats(database *gorm.DB, userID uuid.UUID, timeRange string) (*BookStatsResponse, error) {
	// Calculate date filter
	var dateFilter *time.Time
	now := time.Now()

	switch timeRange {
	case "week":
		date := now.AddDate(0, 0, -7)
		dateFilter = &date
	case "month":
		date := now.AddDate(0, -1, 0)
		dateFilter = &date
	case "year":
		date := now.AddDate(-1, 0, 0)
		dateFilter = &date
	default: // "all"
		dateFilter = nil
	}

	// Build overview stats
	overview, err := buildOverviewStats(database, userID, dateFilter)
	if err != nil {
		return nil, err
	}

	// Build reading habits
	readingHabits, err := buildReadingHabits(database, userID, dateFilter)
	if err != nil {
		return nil, err
	}

	// Build content analytics
	contentAnalytics, err := buildContentAnalytics(database, userID, dateFilter)
	if err != nil {
		return nil, err
	}

	// Build goals and progress
	goalsAndProgress, err := buildGoalsAndProgress(database, userID, dateFilter)
	if err != nil {
		return nil, err
	}

	// Build annotations and engagement
	annotationsAndEngagement, err := buildAnnotationsAndEngagement(database, userID, dateFilter)
	if err != nil {
		return nil, err
	}

	return &BookStatsResponse{
		Overview:                 *overview,
		ReadingHabits:           *readingHabits,
		ContentAnalytics:        *contentAnalytics,
		GoalsAndProgress:        *goalsAndProgress,
		AnnotationsAndEngagement: *annotationsAndEngagement,
	}, nil
}

func buildOverviewStats(database *gorm.DB, userID uuid.UUID, dateFilter *time.Time) (*StatsOverview, error) {
	var totalBooks int64
	var booksRead int64
	var booksInProgress int64
	var totalPagesRead int64
	var totalReadingTime int64
	var longestStreak int64

	// Query with date filter if provided
	booksQuery := database.Model(&models.UserBook{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	progressQuery := database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID)

	if dateFilter != nil {
		booksQuery = booksQuery.Where("created_at >= ?", dateFilter)
		progressQuery = progressQuery.Where("created_at >= ?", dateFilter)
	}

	// Get total books
	if err := booksQuery.Count(&totalBooks).Error; err != nil {
		return nil, err
	}

	// Get reading progress stats
	if err := progressQuery.Where("percentage >= 100").Count(&booksRead).Error; err != nil {
		return nil, err
	}

	if err := progressQuery.Where("percentage > 0 AND percentage < 100").Count(&booksInProgress).Error; err != nil {
		return nil, err
	}

	if err := progressQuery.Select("COALESCE(SUM(current_page), 0)").Scan(&totalPagesRead).Error; err != nil {
		return nil, err
	}

	if err := progressQuery.Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&totalReadingTime).Error; err != nil {
		return nil, err
	}

	if err := progressQuery.Select("MAX(reading_streak_days)").Scan(&longestStreak).Error; err != nil {
		return nil, err
	}

	// Calculate current reading streak
	var currentStreak int64
	var lastRead time.Time
	if err := progressQuery.Select("MAX(last_read)").Scan(&lastRead).Error; err == nil {
		daysSinceLastRead := int(time.Since(lastRead).Hours() / 24)
		if daysSinceLastRead <= 1 {
			// Still in streak - get the actual current streak
			if err := progressQuery.Where("last_read >= ?", time.Now().AddDate(0, 0, -1)).
				Select("MAX(reading_streak_days)").Scan(&currentStreak).Error; err != nil {
				currentStreak = 0
			}
		}
	}

	// Calculate average reading speed (pages per hour)
	averageSpeed := 0
	if totalReadingTime > 0 {
		averageSpeed = int(float64(totalPagesRead) / (float64(totalReadingTime) / 60.0))
	}

	return &StatsOverview{
		TotalBooks:                      int(totalBooks),
		BooksRead:                      int(booksRead),
		BooksInProgress:                int(booksInProgress),
		TotalPagesRead:                 int(totalPagesRead),
		TotalReadingTimeMinutes:        int(totalReadingTime),
		AverageReadingSpeedPagesPerHour: averageSpeed,
		CurrentReadingStreakDays:       int(currentStreak),
		LongestReadingStreakDays:       int(longestStreak),
	}, nil
}

func buildReadingHabits(database *gorm.DB, userID uuid.UUID, dateFilter *time.Time) (*ReadingHabits, error) {
	// Calculate daily averages
	var totalDays int64 = 365 // Default to year
	if dateFilter != nil {
		totalDays = int64(time.Since(*dateFilter).Hours() / 24)
		if totalDays == 0 {
			totalDays = 1
		}
	}

	progressQuery := database.Model(&models.ReadingProgress{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if dateFilter != nil {
		progressQuery = progressQuery.Where("updated_at >= ?", dateFilter)
	}

	var totalPagesRead int64
	var totalMinutesRead int64
	var booksCompleted int64

	progressQuery.Select("COALESCE(SUM(current_page), 0)").Scan(&totalPagesRead)
	progressQuery.Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&totalMinutesRead)
	progressQuery.Where("percentage >= 100").Count(&booksCompleted)

	dailyAverages := DailyAverages{
		PagesPerDay:   int(totalPagesRead / totalDays),
		MinutesPerDay: int(totalMinutesRead / totalDays),
		BooksPerMonth: int(booksCompleted * 30 / totalDays),
	}

	// Get monthly progress (last 12 months)
	monthlyProgress := []MonthlyProgress{}
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		var pagesRead int64
		var booksCompleted int64
		var readingTime int64

		database.Model(&models.ReadingProgress{}).
			Where("user_id = ? AND updated_at >= ? AND updated_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Select("COALESCE(SUM(current_page), 0)").Scan(&pagesRead)

		database.Model(&models.ReadingProgress{}).
			Where("user_id = ? AND updated_at >= ? AND updated_at <= ? AND percentage >= 100 AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Count(&booksCompleted)

		database.Model(&models.ReadingProgress{}).
			Where("user_id = ? AND updated_at >= ? AND updated_at <= ? AND deleted_at IS NULL", userID, monthStart, monthEnd).
			Select("COALESCE(SUM(time_spent_minutes), 0)").Scan(&readingTime)

		monthlyProgress = append(monthlyProgress, MonthlyProgress{
			Month:              monthStart.Format("January 2006"),
			PagesRead:          int(pagesRead),
			BooksCompleted:     int(booksCompleted),
			ReadingTimeMinutes: int(readingTime),
		})
	}

	// Reading by day of week (dummy data for now)
	dayOfWeekReading := []DayOfWeekReading{
		{Day: "Monday", PagesRead: 45, Sessions: 3},
		{Day: "Tuesday", PagesRead: 52, Sessions: 4},
		{Day: "Wednesday", PagesRead: 38, Sessions: 2},
		{Day: "Thursday", PagesRead: 61, Sessions: 5},
		{Day: "Friday", PagesRead: 34, Sessions: 2},
		{Day: "Saturday", PagesRead: 78, Sessions: 6},
		{Day: "Sunday", PagesRead: 69, Sessions: 5},
	}

	// Reading by hour (top 8 most active hours)
	hourlyReading := []HourlyReading{
		{Hour: 20, PagesRead: 125, Sessions: 15},
		{Hour: 21, PagesRead: 98, Sessions: 12},
		{Hour: 19, PagesRead: 87, Sessions: 11},
		{Hour: 9, PagesRead: 76, Sessions: 9},
		{Hour: 22, PagesRead: 65, Sessions: 8},
		{Hour: 10, PagesRead: 54, Sessions: 7},
		{Hour: 18, PagesRead: 43, Sessions: 6},
		{Hour: 8, PagesRead: 32, Sessions: 4},
	}

	return &ReadingHabits{
		DailyAverages:      dailyAverages,
		MonthlyProgress:    monthlyProgress,
		ReadingByDayOfWeek: dayOfWeekReading,
		ReadingByHour:      hourlyReading,
	}, nil
}

func buildContentAnalytics(database *gorm.DB, userID uuid.UUID, dateFilter *time.Time) (*ContentAnalytics, error) {
	// Get genre statistics
	type GenreResult struct {
		Genre      string  `json:"genre"`
		BookCount  int64   `json:"book_count"`
		PagesRead  int64   `json:"pages_read"`
		Completed  int64   `json:"completed"`
	}

	var genreResults []GenreResult
	genreQuery := `
		SELECT 
			COALESCE(b.genre, 'Unknown') as genre,
			COUNT(DISTINCT ub.book_id) as book_count,
			COALESCE(SUM(rp.current_page), 0) as pages_read,
			COUNT(DISTINCT CASE WHEN rp.percentage >= 100 THEN ub.book_id END) as completed
		FROM user_books ub
		JOIN books b ON ub.book_id = b.id
		LEFT JOIN reading_progress rp ON ub.user_id = rp.user_id AND ub.book_id = rp.book_id AND rp.deleted_at IS NULL
		WHERE ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL
	`

	args := []interface{}{userID}
	if dateFilter != nil {
		genreQuery += " AND ub.created_at >= ?"
		args = append(args, dateFilter)
	}

	genreQuery += " GROUP BY b.genre ORDER BY book_count DESC LIMIT 10"

	if err := database.Raw(genreQuery, args...).Scan(&genreResults).Error; err != nil {
		return nil, err
	}

	genres := make([]GenreStats, len(genreResults))
	for i, result := range genreResults {
		completionRate := float64(0)
		if result.BookCount > 0 {
			completionRate = float64(result.Completed) / float64(result.BookCount) * 100
		}

		genres[i] = GenreStats{
			Genre:          result.Genre,
			BookCount:      int(result.BookCount),
			PagesRead:      int(result.PagesRead),
			CompletionRate: completionRate,
		}
	}

	// Get language statistics
	type LanguageResult struct {
		Language           string `json:"language"`
		BookCount          int64  `json:"book_count"`
		PagesRead          int64  `json:"pages_read"`
		ReadingTimeMinutes int64  `json:"reading_time_minutes"`
	}

	var languageResults []LanguageResult
	languageQuery := `
		SELECT 
			UPPER(COALESCE(b.language, 'en')) as language,
			COUNT(DISTINCT ub.book_id) as book_count,
			COALESCE(SUM(rp.current_page), 0) as pages_read,
			COALESCE(SUM(rp.time_spent_minutes), 0) as reading_time_minutes
		FROM user_books ub
		JOIN books b ON ub.book_id = b.id
		LEFT JOIN reading_progress rp ON ub.user_id = rp.user_id AND ub.book_id = rp.book_id AND rp.deleted_at IS NULL
		WHERE ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL
	`

	args = []interface{}{userID}
	if dateFilter != nil {
		languageQuery += " AND ub.created_at >= ?"
		args = append(args, dateFilter)
	}

	languageQuery += " GROUP BY b.language ORDER BY book_count DESC"

	if err := database.Raw(languageQuery, args...).Scan(&languageResults).Error; err != nil {
		return nil, err
	}

	languages := make([]LanguageStats, len(languageResults))
	for i, result := range languageResults {
		languages[i] = LanguageStats{
			Language:           result.Language,
			BookCount:          int(result.BookCount),
			PagesRead:          int(result.PagesRead),
			ReadingTimeMinutes: int(result.ReadingTimeMinutes),
		}
	}

	// Get author statistics
	type AuthorResult struct {
		Author    string `json:"author"`
		BookCount int64  `json:"book_count"`
		PagesRead int64  `json:"pages_read"`
	}

	var authorResults []AuthorResult
	authorQuery := `
		SELECT 
			b.author,
			COUNT(DISTINCT ub.book_id) as book_count,
			COALESCE(SUM(rp.current_page), 0) as pages_read
		FROM user_books ub
		JOIN books b ON ub.book_id = b.id
		LEFT JOIN reading_progress rp ON ub.user_id = rp.user_id AND ub.book_id = rp.book_id AND rp.deleted_at IS NULL
		WHERE ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL
	`

	args = []interface{}{userID}
	if dateFilter != nil {
		authorQuery += " AND ub.created_at >= ?"
		args = append(args, dateFilter)
	}

	authorQuery += " GROUP BY b.author ORDER BY book_count DESC LIMIT 10"

	if err := database.Raw(authorQuery, args...).Scan(&authorResults).Error; err != nil {
		return nil, err
	}

	authors := make([]AuthorStats, len(authorResults))
	for i, result := range authorResults {
		authors[i] = AuthorStats{
			Author:    result.Author,
			BookCount: int(result.BookCount),
			PagesRead: int(result.PagesRead),
		}
	}

	// Get file format statistics
	type FormatResult struct {
		Format    string  `json:"format"`
		Count     int64   `json:"count"`
		TotalSize int64   `json:"total_size"`
	}

	var formatResults []FormatResult
	formatQuery := `
		SELECT 
			UPPER(b.file_type) as format,
			COUNT(*) as count,
			COALESCE(SUM(b.file_size), 0) as total_size
		FROM user_books ub
		JOIN books b ON ub.book_id = b.id
		WHERE ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL
	`

	args = []interface{}{userID}
	if dateFilter != nil {
		formatQuery += " AND ub.created_at >= ?"
		args = append(args, dateFilter)
	}

	formatQuery += " GROUP BY b.file_type ORDER BY count DESC"

	if err := database.Raw(formatQuery, args...).Scan(&formatResults).Error; err != nil {
		return nil, err
	}

	fileFormats := make([]FileFormatStats, len(formatResults))
	for i, result := range formatResults {
		fileFormats[i] = FileFormatStats{
			Format:      result.Format,
			Count:       int(result.Count),
			TotalSizeMB: float64(result.TotalSize) / (1024 * 1024),
		}
	}

	return &ContentAnalytics{
		Genres:      genres,
		Languages:   languages,
		Authors:     authors,
		FileFormats: fileFormats,
	}, nil
}

func buildGoalsAndProgress(database *gorm.DB, userID uuid.UUID, dateFilter *time.Time) (*GoalsAndProgress, error) {
	// Get current year's reading progress
	yearStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	var booksReadThisYear int64

	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND percentage >= 100 AND updated_at >= ? AND deleted_at IS NULL", userID, yearStart).
		Count(&booksReadThisYear)

	// Create sample yearly goal
	targetBooks := 24 // Sample target
	progressPercentage := float64(booksReadThisYear) / float64(targetBooks) * 100
	if progressPercentage > 100 {
		progressPercentage = 100
	}

	// Calculate projected completion
	daysPassed := time.Since(yearStart).Hours() / 24
	dailyRate := float64(booksReadThisYear) / daysPassed
	daysRemaining := 365 - daysPassed
	projectedTotal := booksReadThisYear + int64(dailyRate*daysRemaining)

	var projectedCompletion string
	if projectedTotal >= int64(targetBooks) {
		projectedCompletion = "On track"
	} else {
		projectedCompletion = "Behind schedule"
	}

	yearlyGoal := &YearlyGoal{
		TargetBooks:         targetBooks,
		CompletedBooks:      int(booksReadThisYear),
		ProgressPercentage:  progressPercentage,
		ProjectedCompletion: projectedCompletion,
	}

	// Create sample daily goal
	var avgPagesPerDay int64
	database.Model(&models.ReadingProgress{}).
		Where("user_id = ? AND updated_at >= ? AND deleted_at IS NULL", userID, yearStart).
		Select("AVG(current_page)").Scan(&avgPagesPerDay)

	dailyGoal := &DailyGoal{
		TargetPages:     25, // Sample target
		AverageAchieved: int(avgPagesPerDay),
		SuccessRate:     75.5, // Sample success rate
	}

	// Create sample milestones
	milestones := []ReadingMilestone{
		{Milestone: "Read 10 books", AchievedDate: stringPtr("2024-03-15")},
		{Milestone: "Read 25 books", Progress: intPtr(int(booksReadThisYear * 100 / 25))},
		{Milestone: "Read 50 books", Progress: intPtr(int(booksReadThisYear * 100 / 50))},
		{Milestone: "Read 1000 pages", AchievedDate: stringPtr("2024-02-28")},
	}

	return &GoalsAndProgress{
		YearlyGoal:        yearlyGoal,
		DailyGoal:         dailyGoal,
		ReadingMilestones: milestones,
	}, nil
}

func buildAnnotationsAndEngagement(database *gorm.DB, userID uuid.UUID, dateFilter *time.Time) (*AnnotationsAndEngagement, error) {
	annotationQuery := database.Model(&models.Annotation{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if dateFilter != nil {
		annotationQuery = annotationQuery.Where("created_at >= ?", dateFilter)
	}

	var totalAnnotations int64
	var highlights int64
	var notes int64
	var bookmarks int64

	annotationQuery.Count(&totalAnnotations)
	annotationQuery.Where("type = 'highlight'").Count(&highlights)
	annotationQuery.Where("type = 'note'").Count(&notes)
	annotationQuery.Where("type = 'bookmark'").Count(&bookmarks)

	// Get most annotated books
	type AnnotatedBookResult struct {
		Title           string `json:"title"`
		Author          string `json:"author"`
		AnnotationCount int64  `json:"annotation_count"`
	}

	var annotatedBookResults []AnnotatedBookResult
	bookQuery := `
		SELECT 
			b.title,
			b.author,
			COUNT(a.id) as annotation_count
		FROM annotations a
		JOIN books b ON a.book_id = b.id
		WHERE a.user_id = ? AND a.deleted_at IS NULL AND b.deleted_at IS NULL
	`

	args := []interface{}{userID}
	if dateFilter != nil {
		bookQuery += " AND a.created_at >= ?"
		args = append(args, dateFilter)
	}

	bookQuery += " GROUP BY b.id, b.title, b.author ORDER BY annotation_count DESC LIMIT 5"

	if err := database.Raw(bookQuery, args...).Scan(&annotatedBookResults).Error; err != nil {
		return nil, err
	}

	mostAnnotatedBooks := make([]MostAnnotatedBook, len(annotatedBookResults))
	for i, result := range annotatedBookResults {
		mostAnnotatedBooks[i] = MostAnnotatedBook{
			Title:           result.Title,
			Author:          result.Author,
			AnnotationCount: int(result.AnnotationCount),
		}
	}

	return &AnnotationsAndEngagement{
		TotalAnnotations:   int(totalAnnotations),
		Highlights:         int(highlights),
		Notes:              int(notes),
		Bookmarks:          int(bookmarks),
		MostAnnotatedBooks: mostAnnotatedBooks,
	}, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}