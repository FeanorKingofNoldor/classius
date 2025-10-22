package models

import (
	"testing"
	"github.com/google/uuid"
)

func TestBookModel(t *testing.T) {
	// Test Book creation
	book := &Book{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Title:    "The Republic",
		Author:   "Plato",
		Language: "en",
		FileType: "epub",
		FileSize: 1024000,
		FilePath: "/uploads/test.epub",
		Status:   BookStatusActive,
	}

	if book.Title != "The Republic" {
		t.Errorf("Expected title 'The Republic', got %s", book.Title)
	}

	if book.Author != "Plato" {
		t.Errorf("Expected author 'Plato', got %s", book.Author)
	}

	if book.Status != BookStatusActive {
		t.Errorf("Expected status 'active', got %s", book.Status)
	}
}

func TestBookFileExtension(t *testing.T) {
	testCases := []struct {
		fileType string
		expected string
	}{
		{"epub", ".epub"},
		{"pdf", ".pdf"},
		{"txt", ".txt"},
		{"mobi", ".mobi"},
		{"azw3", ".azw3"},
		{"unknown", ""},
	}

	for _, tc := range testCases {
		book := &Book{FileType: tc.fileType}
		result := book.GetFileExtension()
		if result != tc.expected {
			t.Errorf("For file type %s, expected %s, got %s", tc.fileType, tc.expected, result)
		}
	}
}

func TestIsValidFileType(t *testing.T) {
	validTypes := []string{"epub", "pdf", "txt", "mobi", "azw", "azw3"}
	invalidTypes := []string{"doc", "docx", "unknown", ""}

	for _, fileType := range validTypes {
		if !IsValidFileType(fileType) {
			t.Errorf("Expected %s to be a valid file type", fileType)
		}
	}

	for _, fileType := range invalidTypes {
		if IsValidFileType(fileType) {
			t.Errorf("Expected %s to be an invalid file type", fileType)
		}
	}
}

func TestGetMimeType(t *testing.T) {
	testCases := []struct {
		fileType string
		expected string
	}{
		{"epub", "application/epub+zip"},
		{"pdf", "application/pdf"},
		{"txt", "text/plain"},
		{"unknown", "application/octet-stream"},
	}

	for _, tc := range testCases {
		result := GetMimeType(tc.fileType)
		if result != tc.expected {
			t.Errorf("For file type %s, expected %s, got %s", tc.fileType, tc.expected, result)
		}
	}
}

func TestBookFilter(t *testing.T) {
	userID := uuid.New()
	filter := &BookFilter{
		UserID:    userID,
		Query:     "plato",
		Author:    "Plato", 
		Genre:     "Philosophy",
		Language:  "en",
		Status:    BookStatusActive,
		FileType:  "epub",
		Page:      1,
		PerPage:   20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if filter.UserID != userID {
		t.Errorf("Expected UserID to be %s, got %s", userID, filter.UserID)
	}

	if filter.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", filter.Page)
	}

	if filter.PerPage != 20 {
		t.Errorf("Expected PerPage to be 20, got %d", filter.PerPage)
	}
}