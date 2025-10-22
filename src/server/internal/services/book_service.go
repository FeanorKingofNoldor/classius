package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/classius/server/internal/models"
)

// BookService handles book-related operations
type BookService struct {
	db          *gorm.DB
	uploadPath  string
	maxFileSize int64 // Maximum file size in bytes
}

// NewBookService creates a new book service
func NewBookService(db *gorm.DB, uploadPath string, maxFileSize int64) *BookService {
	return &BookService{
		db:          db,
		uploadPath:  uploadPath,
		maxFileSize: maxFileSize,
	}
}

// BookUploadRequest represents a book upload request
type BookUploadRequest struct {
	Title       string    `json:"title" binding:"required"`
	Author      string    `json:"author" binding:"required"`
	Language    string    `json:"language,omitempty"`
	Genre       string    `json:"genre,omitempty"`
	Publisher   string    `json:"publisher,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ISBN        string    `json:"isbn,omitempty"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	IsPublic    bool      `json:"is_public,omitempty"`
}

// BookUpdateRequest represents a book update request
type BookUpdateRequest struct {
	Title       *string    `json:"title,omitempty"`
	Author      *string    `json:"author,omitempty"`
	Language    *string    `json:"language,omitempty"`
	Genre       *string    `json:"genre,omitempty"`
	Publisher   *string    `json:"publisher,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ISBN        *string    `json:"isbn,omitempty"`
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	IsPublic    *bool      `json:"is_public,omitempty"`
	Status      *models.BookStatus `json:"status,omitempty"`
}

// UploadBook uploads a book file and creates a book record
func (s *BookService) UploadBook(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, req *BookUploadRequest) (*models.Book, error) {
	// Validate file size
	if file.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", file.Size, s.maxFileSize)
	}

	// Detect file type
	fileType, err := s.detectFileType(file.Filename)
	if err != nil {
		return nil, fmt.Errorf("unsupported file type: %w", err)
	}

	// Generate unique file path
	bookID := uuid.New()
	// Create a temporary book instance to call the method
	tempBook := &models.Book{FileType: fileType}
	fileName := fmt.Sprintf("%s_%s%s", bookID.String(), sanitizeFilename(file.Filename), tempBook.GetFileExtension())
	filePath := filepath.Join(s.uploadPath, userID.String(), fileName)

	// Ensure upload directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save file to disk
	if err := s.saveUploadedFile(file, filePath); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	// Extract additional metadata from file if possible
	metadata, err := s.extractMetadata(filePath, fileType)
	if err != nil {
		// Log the error but don't fail the upload
		fmt.Printf("Warning: failed to extract metadata from %s: %v\n", filePath, err)
		metadata = &models.BookMetadata{}
	}

	// Set original filename in metadata
	metadata.OriginalFileName = file.Filename
	metadata.MimeType = models.GetMimeType(fileType)

	// Create book record
	book := &models.Book{
		ID:       bookID,
		UserID:   userID,
		Title:    req.Title,
		Author:   req.Author,
		Language: req.Language,
		Genre:    req.Genre,
		Publisher: req.Publisher,
		PublishedAt: req.PublishedAt,
		ISBN:     req.ISBN,
		Description: req.Description,
		FilePath: filePath,
		FileSize: file.Size,
		FileType: fileType,
		Status:   models.BookStatusProcessing,
		IsPublic: req.IsPublic,
		Metadata: *metadata,
	}

	// Set default language if not provided
	if book.Language == "" {
		book.Language = "en"
	}

	// Start database transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save book to database
	if err := tx.Create(book).Error; err != nil {
		tx.Rollback()
		os.Remove(filePath) // Clean up uploaded file
		return nil, fmt.Errorf("failed to save book to database: %w", err)
	}

	// Handle tags if provided
	if len(req.Tags) > 0 {
		if err := s.handleBookTags(tx, userID, book.ID, req.Tags); err != nil {
			tx.Rollback()
			os.Remove(filePath) // Clean up uploaded file
			return nil, fmt.Errorf("failed to handle tags: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		os.Remove(filePath) // Clean up uploaded file
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Process book in background (extract word count, page count, etc.)
	go s.processBookInBackground(book.ID)

	return book, nil
}

// GetBook retrieves a book by ID for a specific user
func (s *BookService) GetBook(ctx context.Context, userID, bookID uuid.UUID) (*models.Book, error) {
	var book models.Book
	if err := s.db.WithContext(ctx).
		Preload("Tags").
		Where("id = ? AND user_id = ?", bookID, userID).
		First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("book not found")
		}
		return nil, fmt.Errorf("failed to retrieve book: %w", err)
	}
	return &book, nil
}

// GetBooks retrieves books for a user with filtering and pagination
func (s *BookService) GetBooks(ctx context.Context, filter *models.BookFilter) (*models.BookSearchResult, error) {
	query := s.db.WithContext(ctx).Model(&models.Book{}).
		Preload("Tags").
		Where("user_id = ?", filter.UserID)

	// Apply filters
	if filter.Query != "" {
		searchTerm := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(author) LIKE ? OR LOWER(description) LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	if filter.Author != "" {
		query = query.Where("LOWER(author) LIKE ?", "%"+strings.ToLower(filter.Author)+"%")
	}

	if filter.Genre != "" {
		query = query.Where("genre = ?", filter.Genre)
	}

	if filter.Language != "" {
		query = query.Where("language = ?", filter.Language)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.FileType != "" {
		query = query.Where("file_type = ?", filter.FileType)
	}

	if filter.IsPublic != nil {
		query = query.Where("is_public = ?", *filter.IsPublic)
	}

	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filter.CreatedAfter)
	}

	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filter.CreatedBefore)
	}

	// Handle tag filtering
	if len(filter.Tags) > 0 {
		query = query.Joins("JOIN book_tags ON books.id = book_tags.book_id").
			Joins("JOIN tags ON book_tags.tag_id = tags.id").
			Where("tags.name IN ?", filter.Tags).
			Group("books.id").
			Having("COUNT(DISTINCT tags.id) = ?", len(filter.Tags))
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count books: %w", err)
	}

	// Apply sorting
	orderBy := "created_at DESC" // default
	if filter.SortBy != "" {
		validSorts := map[string]bool{
			"title":      true,
			"author":     true,
			"created_at": true,
			"updated_at": true,
			"file_size":  true,
		}
		if validSorts[filter.SortBy] {
			order := "ASC"
			if filter.SortOrder == "desc" {
				order = "DESC"
			}
			orderBy = filter.SortBy + " " + order
		}
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	offset := (filter.Page - 1) * filter.PerPage
	query = query.Offset(offset).Limit(filter.PerPage)

	// Execute query
	var books []models.Book
	if err := query.Find(&books).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve books: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &models.BookSearchResult{
		Books:      books,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateBook updates a book's metadata
func (s *BookService) UpdateBook(ctx context.Context, userID, bookID uuid.UUID, req *BookUpdateRequest) (*models.Book, error) {
	// Get existing book
	book, err := s.GetBook(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Author != nil {
		updates["author"] = *req.Author
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.Genre != nil {
		updates["genre"] = *req.Genre
	}
	if req.Publisher != nil {
		updates["publisher"] = *req.Publisher
	}
	if req.PublishedAt != nil {
		updates["published_at"] = *req.PublishedAt
	}
	if req.ISBN != nil {
		updates["isbn"] = *req.ISBN
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	// Apply updates
	if len(updates) > 0 {
		if err := tx.Model(book).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update book: %w", err)
		}
	}

	// Handle tags update if provided
	if req.Tags != nil {
		// Clear existing tags
		if err := tx.Exec("DELETE FROM book_tags WHERE book_id = ?", bookID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to clear existing tags: %w", err)
		}

		// Add new tags
		if len(req.Tags) > 0 {
			if err := s.handleBookTags(tx, userID, bookID, req.Tags); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to handle tags: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated book
	return s.GetBook(ctx, userID, bookID)
}

// DeleteBook deletes a book and its associated file
func (s *BookService) DeleteBook(ctx context.Context, userID, bookID uuid.UUID) error {
	book, err := s.GetBook(ctx, userID, bookID)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete book record (this will cascade to book_tags due to foreign key)
	if err := tx.Delete(&models.Book{}, "id = ? AND user_id = ?", bookID, userID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete book from database: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Delete file from disk
	if book.FilePath != "" {
		if err := os.Remove(book.FilePath); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to delete file %s: %v\n", book.FilePath, err)
		}
	}

	return nil
}

// GetBookStats returns statistics about a user's book collection
func (s *BookService) GetBookStats(ctx context.Context, userID uuid.UUID) (*models.BookStats, error) {
	var stats models.BookStats

	// Count total books
	var totalBooks int64
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Where("user_id = ?", userID).
		Count(&totalBooks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total books: %w", err)
	}
	stats.TotalBooks = int(totalBooks)

	// Calculate total size
	var totalSize int64
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total size: %w", err)
	}
	stats.TotalSize = totalSize

	// Books by genre
	stats.BooksByGenre = make(map[string]int)
	var genreResults []struct {
		Genre string
		Count int
	}
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Select("genre, COUNT(*) as count").
		Where("user_id = ? AND genre IS NOT NULL AND genre != ''", userID).
		Group("genre").
		Scan(&genreResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get books by genre: %w", err)
	}
	for _, result := range genreResults {
		stats.BooksByGenre[result.Genre] = result.Count
	}

	// Books by language
	stats.BooksByLanguage = make(map[string]int)
	var languageResults []struct {
		Language string
		Count    int
	}
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Select("language, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("language").
		Scan(&languageResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get books by language: %w", err)
	}
	for _, result := range languageResults {
		stats.BooksByLanguage[result.Language] = result.Count
	}

	// Books by file type
	stats.BooksByFileType = make(map[string]int)
	var fileTypeResults []struct {
		FileType string
		Count    int
	}
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Select("file_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("file_type").
		Scan(&fileTypeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get books by file type: %w", err)
	}
	for _, result := range fileTypeResults {
		stats.BooksByFileType[result.FileType] = result.Count
	}

	// Books by status
	stats.BooksByStatus = make(map[string]int)
	var statusResults []struct {
		Status string
		Count  int
	}
	if err := s.db.WithContext(ctx).Model(&models.Book{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&statusResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get books by status: %w", err)
	}
	for _, result := range statusResults {
		stats.BooksByStatus[result.Status] = result.Count
	}

	// Recently added books (last 5)
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&stats.RecentlyAdded).Error; err != nil {
		return nil, fmt.Errorf("failed to get recently added books: %w", err)
	}

	// Most annotated books (top 5)
	if err := s.db.WithContext(ctx).
		Preload("Annotations").
		Where("user_id = ?", userID).
		Joins("LEFT JOIN annotations ON books.id = annotations.book_id").
		Group("books.id").
		Order("COUNT(annotations.id) DESC").
		Limit(5).
		Find(&stats.MostAnnotated).Error; err != nil {
		return nil, fmt.Errorf("failed to get most annotated books: %w", err)
	}

	return &stats, nil
}

// Helper methods

// detectFileType determines the file type based on the filename extension
func (s *BookService) detectFileType(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".epub":
		return "epub", nil
	case ".pdf":
		return "pdf", nil
	case ".txt":
		return "txt", nil
	case ".mobi":
		return "mobi", nil
	case ".azw", ".azw3":
		return "azw3", nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// saveUploadedFile saves a multipart file to disk
func (s *BookService) saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// extractMetadata extracts metadata from uploaded files
func (s *BookService) extractMetadata(filePath, fileType string) (*models.BookMetadata, error) {
	metadata := &models.BookMetadata{}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return metadata, err
	}

	// Basic metadata available for all file types
	metadata.MimeType = models.GetMimeType(fileType)

	// TODO: Implement format-specific metadata extraction
	// For now, return basic metadata
	switch fileType {
	case "epub":
		// Could use a library like github.com/bmaupin/go-epub to extract EPUB metadata
		metadata.Format.Version = "EPUB"
	case "pdf":
		// Could use a library like github.com/ledongthuc/pdf to extract PDF metadata
		metadata.Format.Version = "PDF"
	case "txt":
		metadata.Format.Version = "Plain Text"
		metadata.Encoding = "UTF-8"
	}

	_ = fileInfo // Use fileInfo for additional metadata if needed

	return metadata, nil
}

// handleBookTags creates or associates tags with a book
func (s *BookService) handleBookTags(tx *gorm.DB, userID, bookID uuid.UUID, tagNames []string) error {
	for _, tagName := range tagNames {
		if strings.TrimSpace(tagName) == "" {
			continue
		}

		var tag models.Tag
		// Find or create tag
		if err := tx.Where("user_id = ? AND name = ?", userID, tagName).
			First(&tag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new tag
				tag = models.Tag{
					ID:     uuid.New(),
					UserID: userID,
					Name:   tagName,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return fmt.Errorf("failed to create tag %s: %w", tagName, err)
				}
			} else {
				return fmt.Errorf("failed to find tag %s: %w", tagName, err)
			}
		}

		// Associate tag with book
		if err := tx.Exec("INSERT INTO book_tags (book_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			bookID, tag.ID).Error; err != nil {
			return fmt.Errorf("failed to associate tag %s with book: %w", tagName, err)
		}
	}
	return nil
}

// processBookInBackground processes a book to extract additional metadata
func (s *BookService) processBookInBackground(bookID uuid.UUID) {
	// This would run text analysis, word counting, etc.
	// For now, just update the status
	time.Sleep(2 * time.Second) // Simulate processing

	if err := s.db.Model(&models.Book{}).
		Where("id = ?", bookID).
		Update("status", models.BookStatusActive).Error; err != nil {
		fmt.Printf("Failed to update book %s status: %v\n", bookID, err)
	}
}

// GetDB returns the database instance (for handlers that need direct access)
func (s *BookService) GetDB() *gorm.DB {
	return s.db
}

// sanitizeFilename removes or replaces invalid filename characters
func sanitizeFilename(filename string) string {
	// Remove file extension for sanitization
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Replace problematic characters
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	
	return replacer.Replace(name)
}