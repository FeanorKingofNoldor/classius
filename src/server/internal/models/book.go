package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Book represents a book in the user's library
type Book struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string         `json:"title" gorm:"not null"`
	Author      string         `json:"author" gorm:"not null"`
	Language    string         `json:"language" gorm:"default:'en'"`
	Genre       string         `json:"genre,omitempty"`
	Publisher   string         `json:"publisher,omitempty"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
	ISBN        string         `json:"isbn,omitempty" gorm:"index"`
	Description string         `json:"description,omitempty" gorm:"type:text"`
	CoverURL    string         `json:"cover_url,omitempty"`
	FileURL     string         `json:"file_url,omitempty"`
	FilePath    string         `json:"file_path,omitempty" gorm:"not null"`
	FileSize    int64          `json:"file_size" gorm:"not null"`
	FileType    string         `json:"file_type" gorm:"not null"` // epub, pdf, txt, etc.
	PageCount   int            `json:"page_count,omitempty"`
	WordCount   int            `json:"word_count,omitempty"`
	Status      BookStatus     `json:"status" gorm:"default:'active'"`
	IsPublic    bool           `json:"is_public" gorm:"default:false"`
	Tags        []Tag          `json:"tags,omitempty" gorm:"many2many:book_tags;"`
	Metadata    BookMetadata   `json:"metadata,omitempty" gorm:"embedded"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// Relationships
	User        User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Annotations []Annotation   `json:"annotations,omitempty" gorm:"foreignKey:BookID"`
	Progress    []UserProgress `json:"progress,omitempty" gorm:"foreignKey:BookID"`
}

// BookStatus represents the status of a book
type BookStatus string

const (
	BookStatusActive    BookStatus = "active"
	BookStatusArchived  BookStatus = "archived"
	BookStatusProcessing BookStatus = "processing"
	BookStatusError     BookStatus = "error"
)

// BookMetadata contains additional metadata about the book
type BookMetadata struct {
	OriginalFileName string            `json:"original_file_name,omitempty"`
	MimeType         string            `json:"mime_type,omitempty"`
	Encoding         string            `json:"encoding,omitempty"`
	HasImages        bool              `json:"has_images"`
	HasTOC           bool              `json:"has_toc"` // Table of Contents
	ChapterCount     int               `json:"chapter_count,omitempty"`
	Format           BookFormat        `json:"format,omitempty" gorm:"embedded"`
	ExtraData        map[string]string `json:"extra_data,omitempty" gorm:"type:jsonb"`
}

// BookFormat contains format-specific information
type BookFormat struct {
	Version     string `json:"version,omitempty"`     // e.g., EPUB 3.0, PDF 1.4
	DRM         bool   `json:"drm"`                   // Digital Rights Management
	Interactive bool   `json:"interactive"`           // Has interactive elements
	Rights      string `json:"rights,omitempty"`      // Copyright information
	Source      string `json:"source,omitempty"`      // Where the book was acquired
}

// Tag represents a tag that can be applied to books
type Tag struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Name      string         `json:"name" gorm:"not null"`
	Color     string         `json:"color,omitempty"` // Hex color code
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// Ensure unique tag names per user
	// UniqueIndex will be created in migration
}

// BookSearchResult represents search results for books
type BookSearchResult struct {
	Books      []Book `json:"books"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

// BookFilter represents filtering options for book queries
type BookFilter struct {
	UserID     uuid.UUID    `json:"user_id"`
	Query      string       `json:"query,omitempty"`      // Search in title, author, description
	Author     string       `json:"author,omitempty"`
	Genre      string       `json:"genre,omitempty"`
	Language   string       `json:"language,omitempty"`
	Status     BookStatus   `json:"status,omitempty"`
	Tags       []string     `json:"tags,omitempty"`
	FileType   string       `json:"file_type,omitempty"`
	IsPublic   *bool        `json:"is_public,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	SortBy     string       `json:"sort_by,omitempty"`    // title, author, created_at, updated_at
	SortOrder  string       `json:"sort_order,omitempty"` // asc, desc
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
}

// BookStats represents statistics about a user's book collection
type BookStats struct {
	TotalBooks      int            `json:"total_books"`
	TotalSize       int64          `json:"total_size_bytes"`
	BooksByGenre    map[string]int `json:"books_by_genre"`
	BooksByLanguage map[string]int `json:"books_by_language"`
	BooksByFileType map[string]int `json:"books_by_file_type"`
	BooksByStatus   map[string]int `json:"books_by_status"`
	RecentlyAdded   []Book         `json:"recently_added"`
	MostAnnotated   []Book         `json:"most_annotated"`
}

// BeforeCreate sets up the book before creation
func (b *Book) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Book model
func (Book) TableName() string {
	return "books"
}

// TableName returns the table name for the Tag model
func (Tag) TableName() string {
	return "tags"
}

// GetFileExtension returns the file extension based on file type
func (b *Book) GetFileExtension() string {
	switch b.FileType {
	case "epub":
		return ".epub"
	case "pdf":
		return ".pdf"
	case "txt":
		return ".txt"
	case "mobi":
		return ".mobi"
	case "azw", "azw3":
		return ".azw3"
	default:
		return ""
	}
}

// IsValidFileType checks if the file type is supported
func IsValidFileType(fileType string) bool {
	validTypes := map[string]bool{
		"epub": true,
		"pdf":  true,
		"txt":  true,
		"mobi": true,
		"azw":  true,
		"azw3": true,
	}
	return validTypes[fileType]
}

// GetMimeType returns the MIME type for the file type
func GetMimeType(fileType string) string {
	mimeTypes := map[string]string{
		"epub": "application/epub+zip",
		"pdf":  "application/pdf",
		"txt":  "text/plain",
		"mobi": "application/x-mobipocket-ebook",
		"azw":  "application/vnd.amazon.ebook",
		"azw3": "application/vnd.amazon.ebook",
	}
	if mime, exists := mimeTypes[fileType]; exists {
		return mime
	}
	return "application/octet-stream"
}