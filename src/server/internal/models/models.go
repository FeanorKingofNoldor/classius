package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

// User represents a user in the system
type User struct {
	BaseModel
	Username         string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email            string    `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash     string    `json:"-" gorm:"not null;size:255"`
	FullName         string    `json:"full_name" gorm:"size:255"`
	AvatarURL        string    `json:"avatar_url" gorm:"type:text"`
	SubscriptionTier string    `json:"subscription_tier" gorm:"default:'free';size:20"`
	LastActive       *time.Time `json:"last_active"`
	
	// Relationships
	UserBooks        []UserBook        `json:"user_books,omitempty"`
	ReadingProgress  []ReadingProgress `json:"reading_progress,omitempty"`
	Annotations      []Annotation      `json:"annotations,omitempty"`
	Bookmarks        []Bookmark        `json:"bookmarks,omitempty"`
	SageConversations []SageConversation `json:"sage_conversations,omitempty"`
	UserSessions     []UserSession     `json:"user_sessions,omitempty"`
}

// UserProgress represents user reading progress (alias for ReadingProgress for backward compatibility)
type UserProgress = ReadingProgress

// UserBook represents a book in a user's library
type UserBook struct {
	BaseModel
	UserID   uuid.UUID `json:"user_id" gorm:"not null;index"`
	BookID   uuid.UUID `json:"book_id" gorm:"not null;index"`
	Source   string    `json:"source" gorm:"default:'manual';size:50"` // manual, purchased, shared
	AddedAt  time.Time `json:"added_at" gorm:"default:CURRENT_TIMESTAMP"`
	
	// Relationships
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}

// ReadingProgress tracks user's reading progress
type ReadingProgress struct {
	BaseModel
	UserID            uuid.UUID `json:"user_id" gorm:"not null;index"`
	BookID            uuid.UUID `json:"book_id" gorm:"not null;index"`
	CurrentPage       int       `json:"current_page" gorm:"default:0"`
	TotalPages        int       `json:"total_pages" gorm:"default:0"`
	CurrentPosition   int       `json:"current_position" gorm:"default:0"`
	Percentage        float64   `json:"percentage" gorm:"default:0.0"`
	TimeSpentMinutes  int       `json:"time_spent_minutes" gorm:"default:0"`
	LastRead          time.Time `json:"last_read" gorm:"default:CURRENT_TIMESTAMP"`
	ReadingStreakDays int       `json:"reading_streak_days" gorm:"default:0"`
	NotesCount        int       `json:"notes_count" gorm:"default:0"`
	HighlightsCount   int       `json:"highlights_count" gorm:"default:0"`
	
	// Relationships
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}

// TableName specifies the table name for ReadingProgress
func (ReadingProgress) TableName() string {
	return "reading_progress"
}

// Annotation represents highlights, notes, and other annotations
type Annotation struct {
	BaseModel
	UserID        uuid.UUID `json:"user_id" gorm:"not null;index"`
	BookID        uuid.UUID `json:"book_id" gorm:"not null;index"`
	Type          string    `json:"type" gorm:"not null;size:20;index"` // highlight, note, bookmark
	PageNumber    int       `json:"page_number"`
	StartPosition int       `json:"start_position"`
	EndPosition   int       `json:"end_position"`
	SelectedText  string    `json:"selected_text" gorm:"type:text"`
	Content       string    `json:"content" gorm:"type:text"` // Note content
	Color         string    `json:"color" gorm:"size:20"`     // For highlights
	Tags          []string  `json:"tags" gorm:"type:text[]"`
	IsPrivate     bool      `json:"is_private" gorm:"default:true"`
	
	// Relationships
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}

// Bookmark represents page bookmarks
type Bookmark struct {
	BaseModel
	UserID     uuid.UUID `json:"user_id" gorm:"not null;index"`
	BookID     uuid.UUID `json:"book_id" gorm:"not null;index"`
	Name       string    `json:"name" gorm:"size:255"`
	PageNumber int       `json:"page_number" gorm:"not null"`
	Position   int       `json:"position" gorm:"default:0"`
	
	// Relationships
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}

// SageConversation tracks AI conversations
type SageConversation struct {
	BaseModel
	UserID         uuid.UUID `json:"user_id" gorm:"not null;index"`
	BookID         *uuid.UUID `json:"book_id" gorm:"index"`
	PassageText    string    `json:"passage_text" gorm:"type:text"`
	Question       string    `json:"question" gorm:"not null;type:text"`
	Response       string    `json:"response" gorm:"not null;type:text"`
	ContextData    string    `json:"context_data" gorm:"type:jsonb"` // Additional context as JSON
	ResponseTimeMS int       `json:"response_time_ms"`
	Rating         *int      `json:"rating"` // 1-5 star rating
	
	// Relationships
	User User  `json:"user,omitempty"`
	Book *Book `json:"book,omitempty"`
}

// ReadingGroup represents reading clubs/groups
type ReadingGroup struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;size:255"`
	Description string `json:"description" gorm:"type:text"`
	GroupType   string `json:"group_type" gorm:"default:'book_club';size:50"` // book_club, study_group, discussion
	IsPublic    bool   `json:"is_public" gorm:"default:true"`
	MaxMembers  int    `json:"max_members" gorm:"default:50"`
	CreatedBy   uuid.UUID `json:"created_by" gorm:"not null"`
	
	// Relationships
	Members     []GroupMember `json:"members,omitempty"`
	Discussions []Discussion  `json:"discussions,omitempty"`
}

// GroupMember represents membership in reading groups
type GroupMember struct {
	BaseModel
	GroupID  uuid.UUID `json:"group_id" gorm:"not null;index"`
	UserID   uuid.UUID `json:"user_id" gorm:"not null;index"`
	Role     string    `json:"role" gorm:"default:'member';size:20"` // admin, moderator, member
	JoinedAt time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	
	// Relationships
	Group ReadingGroup `json:"group,omitempty"`
	User  User         `json:"user,omitempty"`
}

// Discussion represents group discussions
type Discussion struct {
	BaseModel
	GroupID          *uuid.UUID `json:"group_id" gorm:"index"`
	BookID           *uuid.UUID `json:"book_id" gorm:"index"`
	UserID           uuid.UUID  `json:"user_id" gorm:"not null;index"`
	Title            string     `json:"title" gorm:"size:255"`
	Content          string     `json:"content" gorm:"not null;type:text"`
	PassageReference string     `json:"passage_reference" gorm:"type:text"` // Reference to specific passage
	ParentID         *uuid.UUID `json:"parent_id" gorm:"index"`             // For threaded replies
	Upvotes          int        `json:"upvotes" gorm:"default:0"`
	
	// Relationships
	Group    *ReadingGroup `json:"group,omitempty"`
	Book     *Book         `json:"book,omitempty"`
	User     User          `json:"user,omitempty"`
	Replies  []Discussion  `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

// PublishedNote represents published annotation collections
type PublishedNote struct {
	BaseModel
	AuthorID       uuid.UUID `json:"author_id" gorm:"not null;index"`
	BookID         uuid.UUID `json:"book_id" gorm:"not null;index"`
	Title          string    `json:"title" gorm:"not null;size:255"`
	Description    string    `json:"description" gorm:"type:text"`
	DifficultyLevel string   `json:"difficulty_level" gorm:"default:'intermediate';size:20"`
	PriceCents     int       `json:"price_cents" gorm:"default:0"` // 0 for free
	IsVerified     bool      `json:"is_verified" gorm:"default:false"` // For verified scholars
	Rating         float64   `json:"rating" gorm:"default:0.0"`
	DownloadsCount int       `json:"downloads_count" gorm:"default:0"`
	
	// Relationships
	Author   User         `json:"author,omitempty"`
	Book     Book         `json:"book,omitempty"`
	Overlays []NoteOverlay `json:"overlays,omitempty"`
}

// NoteOverlay represents expert annotations/overlays
type NoteOverlay struct {
	BaseModel
	PublishedNoteID uuid.UUID `json:"published_note_id" gorm:"not null;index"`
	PageNumber      int       `json:"page_number"`
	StartPosition   int       `json:"start_position"`
	EndPosition     int       `json:"end_position"`
	NoteType        string    `json:"note_type" gorm:"size:50"` // explanation, context, cross_reference
	Content         string    `json:"content" gorm:"not null;type:text"`
	DisplayOrder    int       `json:"display_order" gorm:"default:0"`
	
	// Relationships
	PublishedNote PublishedNote `json:"published_note,omitempty"`
}

// ReadingSession tracks individual reading sessions
type ReadingSession struct {
	BaseModel
	UserID          uuid.UUID  `json:"user_id" gorm:"not null;index"`
	BookID          uuid.UUID  `json:"book_id" gorm:"not null;index"`
	StartedAt       time.Time  `json:"started_at" gorm:"not null"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationMinutes int        `json:"duration_minutes" gorm:"default:0"`
	PagesRead       int        `json:"pages_read" gorm:"default:0"`
	StartPage       *int       `json:"start_page"`
	EndPage         *int       `json:"end_page"`
	StartPosition   int        `json:"start_position" gorm:"default:0"`
	EndPosition     int        `json:"end_position" gorm:"default:0"`
	DeviceType      *string    `json:"device_type" gorm:"size:50"`
	
	// Relationships
	User User `json:"user,omitempty"`
	Book Book `json:"book,omitempty"`
}

// UserSession manages user authentication sessions
type UserSession struct {
	BaseModel
	UserID       uuid.UUID `json:"user_id" gorm:"not null;index"`
	DeviceID     string    `json:"device_id" gorm:"not null;size:255"`
	DeviceName   string    `json:"device_name" gorm:"size:255"`
	AccessToken  string    `json:"access_token" gorm:"not null;size:255"`
	RefreshToken string    `json:"refresh_token" gorm:"not null;size:255"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastUsed     time.Time `json:"last_used" gorm:"default:CURRENT_TIMESTAMP"`
	
	// Relationships
	User User `json:"user,omitempty"`
}
