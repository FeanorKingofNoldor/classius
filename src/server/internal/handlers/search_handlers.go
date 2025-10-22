package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/classius/server/internal/db"
	"github.com/classius/server/internal/utils"
)

// SearchRequest represents the search request parameters
type SearchRequest struct {
	Query      string   `json:"query" form:"query"`
	Types      []string `json:"types" form:"types"`         // books, annotations, highlights, notes
	BookIDs    []string `json:"book_ids" form:"book_ids"`
	Authors    []string `json:"authors" form:"authors"`
	Genres     []string `json:"genres" form:"genres"`
	Languages  []string `json:"languages" form:"languages"`
	Tags       []string `json:"tags" form:"tags"`
	DateFrom   string   `json:"date_from" form:"date_from"`
	DateTo     string   `json:"date_to" form:"date_to"`
	SortBy     string   `json:"sort_by" form:"sort_by"`     // relevance, date, title, author
	SortOrder  string   `json:"sort_order" form:"sort_order"` // asc, desc
	Page       int      `json:"page" form:"page"`
	PerPage    int      `json:"per_page" form:"per_page"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Results    SearchResults `json:"results"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
	Query      string        `json:"query"`
	Filters    SearchFilters `json:"filters"`
}

type SearchResults struct {
	Books       []BookResult       `json:"books"`
	Annotations []AnnotationResult `json:"annotations"`
	Highlights  []HighlightResult  `json:"highlights"`
	Notes       []NoteResult       `json:"notes"`
}

type BookResult struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Genre       string   `json:"genre"`
	Language    string   `json:"language"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	FileType    string   `json:"file_type"`
	PageCount   int      `json:"page_count"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	Relevance   float64  `json:"relevance"`
	Snippet     string   `json:"snippet"`
}

type AnnotationResult struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Content      string  `json:"content"`
	SelectedText string  `json:"selected_text"`
	PageNumber   int     `json:"page_number"`
	Color        string  `json:"color"`
	Tags         []string `json:"tags"`
	BookTitle    string  `json:"book_title"`
	BookAuthor   string  `json:"book_author"`
	BookID       string  `json:"book_id"`
	CreatedAt    string  `json:"created_at"`
	Relevance    float64 `json:"relevance"`
	Snippet      string  `json:"snippet"`
}

type HighlightResult struct {
	ID           string   `json:"id"`
	SelectedText string   `json:"selected_text"`
	PageNumber   int      `json:"page_number"`
	Color        string   `json:"color"`
	Tags         []string `json:"tags"`
	BookTitle    string   `json:"book_title"`
	BookAuthor   string   `json:"book_author"`
	BookID       string   `json:"book_id"`
	CreatedAt    string   `json:"created_at"`
	Relevance    float64  `json:"relevance"`
	Snippet      string   `json:"snippet"`
}

type NoteResult struct {
	ID           string   `json:"id"`
	Content      string   `json:"content"`
	SelectedText string   `json:"selected_text"`
	PageNumber   int      `json:"page_number"`
	Tags         []string `json:"tags"`
	BookTitle    string   `json:"book_title"`
	BookAuthor   string   `json:"book_author"`
	BookID       string   `json:"book_id"`
	CreatedAt    string   `json:"created_at"`
	Relevance    float64  `json:"relevance"`
	Snippet      string   `json:"snippet"`
}

type SearchFilters struct {
	AvailableTypes     []string `json:"available_types"`
	AvailableAuthors   []string `json:"available_authors"`
	AvailableGenres    []string `json:"available_genres"`
	AvailableLanguages []string `json:"available_languages"`
	AvailableTags      []string `json:"available_tags"`
}

// GlobalSearch performs search across books, annotations, highlights, and notes
func GlobalSearch(c *gin.Context) {
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

	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters", nil)
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}
	if len(req.Types) == 0 {
		req.Types = []string{"books", "annotations", "highlights", "notes"}
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	database := db.DB
	results, total, filters, err := performSearch(database, userUUID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Search failed", nil)
		return
	}

	totalPages := (total + req.PerPage - 1) / req.PerPage

	response := SearchResponse{
		Results:    *results,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
		Query:      req.Query,
		Filters:    *filters,
	}

	utils.SuccessResponse(c, "Search completed successfully", response)
}

func performSearch(database *gorm.DB, userID uuid.UUID, req SearchRequest) (*SearchResults, int, *SearchFilters, error) {
	results := &SearchResults{}
	totalCount := 0

	// Search books
	if contains(req.Types, "books") {
		books, count, err := searchBooks(database, userID, req)
		if err != nil {
			return nil, 0, nil, err
		}
		results.Books = books
		totalCount += count
	}

	// Search annotations
	if contains(req.Types, "annotations") {
		annotations, count, err := searchAnnotations(database, userID, req)
		if err != nil {
			return nil, 0, nil, err
		}
		results.Annotations = annotations
		totalCount += count
	}

	// Search highlights
	if contains(req.Types, "highlights") {
		highlights, count, err := searchHighlights(database, userID, req)
		if err != nil {
			return nil, 0, nil, err
		}
		results.Highlights = highlights
		totalCount += count
	}

	// Search notes
	if contains(req.Types, "notes") {
		notes, count, err := searchNotes(database, userID, req)
		if err != nil {
			return nil, 0, nil, err
		}
		results.Notes = notes
		totalCount += count
	}

	// Get available filters
	filters, err := getSearchFilters(database, userID)
	if err != nil {
		return nil, 0, nil, err
	}

	return results, totalCount, filters, nil
}

func searchBooks(database *gorm.DB, userID uuid.UUID, req SearchRequest) ([]BookResult, int, error) {
	query := database.Table("user_books ub").
		Select(`b.id, b.title, b.author, b.genre, b.language, b.description, 
		        b.cover_url, b.file_type, b.page_count, b.created_at,
		        ARRAY_AGG(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags`).
		Joins("JOIN books b ON ub.book_id = b.id").
		Joins("LEFT JOIN book_tags bt ON b.id = bt.book_id").
		Joins("LEFT JOIN tags t ON bt.tag_id = t.id AND t.deleted_at IS NULL").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Group("b.id, b.title, b.author, b.genre, b.language, b.description, b.cover_url, b.file_type, b.page_count, b.created_at")

	// Apply search query
	if req.Query != "" {
		searchTerms := strings.Fields(strings.ToLower(req.Query))
		for _, term := range searchTerms {
			query = query.Where("(LOWER(b.title) LIKE ? OR LOWER(b.author) LIKE ? OR LOWER(b.description) LIKE ?)",
				"%"+term+"%", "%"+term+"%", "%"+term+"%")
		}
	}

	// Apply filters
	if len(req.Authors) > 0 {
		query = query.Where("b.author IN ?", req.Authors)
	}
	if len(req.Genres) > 0 {
		query = query.Where("b.genre IN ?", req.Genres)
	}
	if len(req.Languages) > 0 {
		query = query.Where("b.language IN ?", req.Languages)
	}

	// Apply date filters
	if req.DateFrom != "" {
		query = query.Where("b.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("b.created_at <= ?", req.DateTo)
	}

	// Get count
	var count int64
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch req.SortBy {
	case "title":
		if req.SortOrder == "asc" {
			query = query.Order("b.title ASC")
		} else {
			query = query.Order("b.title DESC")
		}
	case "author":
		if req.SortOrder == "asc" {
			query = query.Order("b.author ASC")
		} else {
			query = query.Order("b.author DESC")
		}
	case "date":
		if req.SortOrder == "asc" {
			query = query.Order("b.created_at ASC")
		} else {
			query = query.Order("b.created_at DESC")
		}
	default: // relevance
		query = query.Order("b.created_at DESC")
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	query = query.Offset(offset).Limit(req.PerPage)

	type BookQueryResult struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		Author      string    `json:"author"`
		Genre       string    `json:"genre"`
		Language    string    `json:"language"`
		Description string    `json:"description"`
		CoverURL    string    `json:"cover_url"`
		FileType    string    `json:"file_type"`
		PageCount   int       `json:"page_count"`
		CreatedAt   string    `json:"created_at"`
		Tags        []string  `json:"tags"`
	}

	var queryResults []BookQueryResult
	if err := query.Find(&queryResults).Error; err != nil {
		return nil, 0, err
	}

	books := make([]BookResult, len(queryResults))
	for i, result := range queryResults {
		snippet := ""
		if req.Query != "" && result.Description != "" {
			snippet = generateSnippet(result.Description, req.Query)
		}

		books[i] = BookResult{
			ID:          result.ID,
			Title:       result.Title,
			Author:      result.Author,
			Genre:       result.Genre,
			Language:    result.Language,
			Description: result.Description,
			CoverURL:    result.CoverURL,
			FileType:    result.FileType,
			PageCount:   result.PageCount,
			Tags:        result.Tags,
			CreatedAt:   result.CreatedAt,
			Relevance:   calculateRelevance(req.Query, result.Title, result.Author, result.Description),
			Snippet:     snippet,
		}
	}

	return books, int(count), nil
}

func searchAnnotations(database *gorm.DB, userID uuid.UUID, req SearchRequest) ([]AnnotationResult, int, error) {
	query := database.Table("annotations a").
		Select(`a.id, a.type, a.content, a.selected_text, a.page_number, a.color,
		        a.tags, a.created_at, b.title as book_title, b.author as book_author, b.id as book_id`).
		Joins("JOIN books b ON a.book_id = b.id").
		Where("a.user_id = ? AND a.deleted_at IS NULL AND b.deleted_at IS NULL", userID)

	// Apply search query
	if req.Query != "" {
		searchTerms := strings.Fields(strings.ToLower(req.Query))
		for _, term := range searchTerms {
			query = query.Where("(LOWER(a.content) LIKE ? OR LOWER(a.selected_text) LIKE ?)",
				"%"+term+"%", "%"+term+"%")
		}
	}

	// Apply filters
	if len(req.BookIDs) > 0 {
		query = query.Where("a.book_id IN ?", req.BookIDs)
	}
	if len(req.Authors) > 0 {
		query = query.Where("b.author IN ?", req.Authors)
	}
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("? = ANY(a.tags)", tag)
		}
	}

	// Apply date filters
	if req.DateFrom != "" {
		query = query.Where("a.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("a.created_at <= ?", req.DateTo)
	}

	// Get count
	var count int64
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch req.SortBy {
	case "date":
		if req.SortOrder == "asc" {
			query = query.Order("a.created_at ASC")
		} else {
			query = query.Order("a.created_at DESC")
		}
	default: // relevance
		query = query.Order("a.created_at DESC")
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	query = query.Offset(offset).Limit(req.PerPage)

	type AnnotationQueryResult struct {
		ID           string   `json:"id"`
		Type         string   `json:"type"`
		Content      string   `json:"content"`
		SelectedText string   `json:"selected_text"`
		PageNumber   int      `json:"page_number"`
		Color        string   `json:"color"`
		Tags         []string `json:"tags"`
		CreatedAt    string   `json:"created_at"`
		BookTitle    string   `json:"book_title"`
		BookAuthor   string   `json:"book_author"`
		BookID       string   `json:"book_id"`
	}

	var queryResults []AnnotationQueryResult
	if err := query.Find(&queryResults).Error; err != nil {
		return nil, 0, err
	}

	annotations := make([]AnnotationResult, len(queryResults))
	for i, result := range queryResults {
		snippet := ""
		if req.Query != "" {
			searchText := result.Content + " " + result.SelectedText
			snippet = generateSnippet(searchText, req.Query)
		}

		annotations[i] = AnnotationResult{
			ID:           result.ID,
			Type:         result.Type,
			Content:      result.Content,
			SelectedText: result.SelectedText,
			PageNumber:   result.PageNumber,
			Color:        result.Color,
			Tags:         result.Tags,
			BookTitle:    result.BookTitle,
			BookAuthor:   result.BookAuthor,
			BookID:       result.BookID,
			CreatedAt:    result.CreatedAt,
			Relevance:    calculateRelevance(req.Query, result.Content, result.SelectedText, ""),
			Snippet:      snippet,
		}
	}

	return annotations, int(count), nil
}

func searchHighlights(database *gorm.DB, userID uuid.UUID, req SearchRequest) ([]HighlightResult, int, error) {
	query := database.Table("annotations a").
		Select(`a.id, a.selected_text, a.page_number, a.color, a.tags, a.created_at,
		        b.title as book_title, b.author as book_author, b.id as book_id`).
		Joins("JOIN books b ON a.book_id = b.id").
		Where("a.user_id = ? AND a.type = 'highlight' AND a.deleted_at IS NULL AND b.deleted_at IS NULL", userID)

	// Apply search query
	if req.Query != "" {
		searchTerms := strings.Fields(strings.ToLower(req.Query))
		for _, term := range searchTerms {
			query = query.Where("LOWER(a.selected_text) LIKE ?", "%"+term+"%")
		}
	}

	// Apply filters (similar to annotations)
	if len(req.BookIDs) > 0 {
		query = query.Where("a.book_id IN ?", req.BookIDs)
	}
	if len(req.Authors) > 0 {
		query = query.Where("b.author IN ?", req.Authors)
	}
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("? = ANY(a.tags)", tag)
		}
	}

	// Apply date filters
	if req.DateFrom != "" {
		query = query.Where("a.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("a.created_at <= ?", req.DateTo)
	}

	// Get count
	var count int64
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	switch req.SortBy {
	case "date":
		if req.SortOrder == "asc" {
			query = query.Order("a.created_at ASC")
		} else {
			query = query.Order("a.created_at DESC")
		}
	default:
		query = query.Order("a.created_at DESC")
	}

	offset := (req.Page - 1) * req.PerPage
	query = query.Offset(offset).Limit(req.PerPage)

	type HighlightQueryResult struct {
		ID           string   `json:"id"`
		SelectedText string   `json:"selected_text"`
		PageNumber   int      `json:"page_number"`
		Color        string   `json:"color"`
		Tags         []string `json:"tags"`
		CreatedAt    string   `json:"created_at"`
		BookTitle    string   `json:"book_title"`
		BookAuthor   string   `json:"book_author"`
		BookID       string   `json:"book_id"`
	}

	var queryResults []HighlightQueryResult
	if err := query.Find(&queryResults).Error; err != nil {
		return nil, 0, err
	}

	highlights := make([]HighlightResult, len(queryResults))
	for i, result := range queryResults {
		snippet := ""
		if req.Query != "" {
			snippet = generateSnippet(result.SelectedText, req.Query)
		}

		highlights[i] = HighlightResult{
			ID:           result.ID,
			SelectedText: result.SelectedText,
			PageNumber:   result.PageNumber,
			Color:        result.Color,
			Tags:         result.Tags,
			BookTitle:    result.BookTitle,
			BookAuthor:   result.BookAuthor,
			BookID:       result.BookID,
			CreatedAt:    result.CreatedAt,
			Relevance:    calculateRelevance(req.Query, result.SelectedText, "", ""),
			Snippet:      snippet,
		}
	}

	return highlights, int(count), nil
}

func searchNotes(database *gorm.DB, userID uuid.UUID, req SearchRequest) ([]NoteResult, int, error) {
	query := database.Table("annotations a").
		Select(`a.id, a.content, a.selected_text, a.page_number, a.tags, a.created_at,
		        b.title as book_title, b.author as book_author, b.id as book_id`).
		Joins("JOIN books b ON a.book_id = b.id").
		Where("a.user_id = ? AND a.type = 'note' AND a.deleted_at IS NULL AND b.deleted_at IS NULL", userID)

	// Apply search query
	if req.Query != "" {
		searchTerms := strings.Fields(strings.ToLower(req.Query))
		for _, term := range searchTerms {
			query = query.Where("(LOWER(a.content) LIKE ? OR LOWER(a.selected_text) LIKE ?)",
				"%"+term+"%", "%"+term+"%")
		}
	}

	// Apply filters (similar to annotations)
	if len(req.BookIDs) > 0 {
		query = query.Where("a.book_id IN ?", req.BookIDs)
	}
	if len(req.Authors) > 0 {
		query = query.Where("b.author IN ?", req.Authors)
	}
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where("? = ANY(a.tags)", tag)
		}
	}

	// Apply date filters
	if req.DateFrom != "" {
		query = query.Where("a.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("a.created_at <= ?", req.DateTo)
	}

	// Get count
	var count int64
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	switch req.SortBy {
	case "date":
		if req.SortOrder == "asc" {
			query = query.Order("a.created_at ASC")
		} else {
			query = query.Order("a.created_at DESC")
		}
	default:
		query = query.Order("a.created_at DESC")
	}

	offset := (req.Page - 1) * req.PerPage
	query = query.Offset(offset).Limit(req.PerPage)

	type NoteQueryResult struct {
		ID           string   `json:"id"`
		Content      string   `json:"content"`
		SelectedText string   `json:"selected_text"`
		PageNumber   int      `json:"page_number"`
		Tags         []string `json:"tags"`
		CreatedAt    string   `json:"created_at"`
		BookTitle    string   `json:"book_title"`
		BookAuthor   string   `json:"book_author"`
		BookID       string   `json:"book_id"`
	}

	var queryResults []NoteQueryResult
	if err := query.Find(&queryResults).Error; err != nil {
		return nil, 0, err
	}

	notes := make([]NoteResult, len(queryResults))
	for i, result := range queryResults {
		snippet := ""
		if req.Query != "" {
			searchText := result.Content + " " + result.SelectedText
			snippet = generateSnippet(searchText, req.Query)
		}

		notes[i] = NoteResult{
			ID:           result.ID,
			Content:      result.Content,
			SelectedText: result.SelectedText,
			PageNumber:   result.PageNumber,
			Tags:         result.Tags,
			BookTitle:    result.BookTitle,
			BookAuthor:   result.BookAuthor,
			BookID:       result.BookID,
			CreatedAt:    result.CreatedAt,
			Relevance:    calculateRelevance(req.Query, result.Content, result.SelectedText, ""),
			Snippet:      snippet,
		}
	}

	return notes, int(count), nil
}

func getSearchFilters(database *gorm.DB, userID uuid.UUID) (*SearchFilters, error) {
	filters := &SearchFilters{
		AvailableTypes: []string{"books", "annotations", "highlights", "notes"},
	}

	// Get available authors
	var authors []string
	database.Table("user_books ub").
		Select("DISTINCT b.author").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Pluck("author", &authors)
	filters.AvailableAuthors = authors

	// Get available genres
	var genres []string
	database.Table("user_books ub").
		Select("DISTINCT b.genre").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL AND b.genre != ''", userID).
		Pluck("genre", &genres)
	filters.AvailableGenres = genres

	// Get available languages
	var languages []string
	database.Table("user_books ub").
		Select("DISTINCT b.language").
		Joins("JOIN books b ON ub.book_id = b.id").
		Where("ub.user_id = ? AND ub.deleted_at IS NULL AND b.deleted_at IS NULL", userID).
		Pluck("language", &languages)
	filters.AvailableLanguages = languages

	// Get available tags
	var tags []string
	database.Table("tags").
		Select("DISTINCT name").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Pluck("name", &tags)
	filters.AvailableTags = tags

	return filters, nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateSnippet(text, query string) string {
	if text == "" || query == "" {
		return ""
	}

	searchTerms := strings.Fields(strings.ToLower(query))
	lowerText := strings.ToLower(text)
	
	for _, term := range searchTerms {
		index := strings.Index(lowerText, term)
		if index != -1 {
			start := index - 50
			if start < 0 {
				start = 0
			}
			end := index + len(term) + 50
			if end > len(text) {
				end = len(text)
			}
			
			snippet := text[start:end]
			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(text) {
				snippet = snippet + "..."
			}
			
			return snippet
		}
	}
	
	// Return first 100 characters if no match found
	if len(text) > 100 {
		return text[:100] + "..."
	}
	return text
}

func calculateRelevance(query string, texts ...string) float64 {
	if query == "" {
		return 0.0
	}
	
	searchTerms := strings.Fields(strings.ToLower(query))
	totalScore := 0.0
	
	for _, text := range texts {
		if text == "" {
			continue
		}
		lowerText := strings.ToLower(text)
		for _, term := range searchTerms {
			count := strings.Count(lowerText, term)
			totalScore += float64(count)
		}
	}
	
	return totalScore
}