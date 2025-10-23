package services

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/taylorskalyo/goreader/epub"
)

// TextExtractor handles extraction of text from various file formats
type TextExtractor struct{}

// NewTextExtractor creates a new text extractor
func NewTextExtractor() *TextExtractor {
	return &TextExtractor{}
}

// ExtractedContent represents extracted book content
type ExtractedContent struct {
	Text      string
	PageCount int
	WordCount int
	HasImages bool
	HasTOC    bool
}

// ExtractText extracts text from a file based on its type
func (te *TextExtractor) ExtractText(filePath, fileType string) (*ExtractedContent, error) {
	switch fileType {
	case "pdf":
		return te.extractPDF(filePath)
	case "epub":
		return te.extractEPUB(filePath)
	case "txt":
		return te.extractTXT(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// extractPDF extracts text from PDF files
func (te *TextExtractor) extractPDF(filePath string) (*ExtractedContent, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var text strings.Builder
	pageCount := r.NumPage()
	hasImages := false

	// Extract text from each page
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Get text content
		content, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but continue with other pages
			fmt.Printf("Warning: failed to extract text from page %d: %v\n", pageNum, err)
			continue
		}

		text.WriteString(content)
		text.WriteString("\n\n") // Add page break

		// Note: Image detection would require more sophisticated PDF parsing
		// For now, we'll assume no images
	}

	extractedText := text.String()
	wordCount := countWords(extractedText)

	return &ExtractedContent{
		Text:      extractedText,
		PageCount: pageCount,
		WordCount: wordCount,
		HasImages: hasImages,
		HasTOC:    false, // TODO: Implement TOC detection for PDF
	}, nil
}

// extractEPUB extracts text from EPUB files
func (te *TextExtractor) extractEPUB(filePath string) (*ExtractedContent, error) {
	rc, err := epub.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer rc.Close()

	book := rc.Rootfiles[0]
	var text strings.Builder
	hasImages := false
	hasTOC := false // TODO: Check NCX
	pageCount := len(book.Spine.Itemrefs)

	// Extract text from spine items (reading order)
	for _, itemref := range book.Spine.Itemrefs {
		// Find the content file
		var item *epub.Item
		for i := range book.Manifest.Items {
			if book.Manifest.Items[i].ID == itemref.IDREF {
				item = &book.Manifest.Items[i]
				break
			}
		}

		if item == nil {
			continue
		}

		// Read the content
		content, err := item.Open()
		if err != nil {
			fmt.Printf("Warning: failed to open EPUB item %s: %v\n", item.HREF, err)
			continue
		}

		data, err := io.ReadAll(content)
		content.Close()
		if err != nil {
			fmt.Printf("Warning: failed to read EPUB item %s: %v\n", item.HREF, err)
			continue
		}

		// Strip HTML tags
		contentStr := string(data)
		plainText := stripHTMLTags(contentStr)
		text.WriteString(plainText)
		text.WriteString("\n\n")

		// Check for images
		if strings.Contains(strings.ToLower(contentStr), "<img") {
			hasImages = true
		}
	}

	extractedText := text.String()
	wordCount := countWords(extractedText)

	return &ExtractedContent{
		Text:      extractedText,
		PageCount: pageCount,
		WordCount: wordCount,
		HasImages: hasImages,
		HasTOC:    hasTOC,
	}, nil
}

// extractTXT extracts text from plain text files
func (te *TextExtractor) extractTXT(filePath string) (*ExtractedContent, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	text := string(content)
	wordCount := countWords(text)
	// Estimate pages (assuming 500 words per page)
	pageCount := (wordCount + 499) / 500

	return &ExtractedContent{
		Text:      text,
		PageCount: pageCount,
		WordCount: wordCount,
		HasImages: false,
		HasTOC:    false,
	}, nil
}

// stripHTMLTags removes HTML tags from text (basic implementation)
func stripHTMLTags(html string) string {
	// Remove script and style tags with their content
	html = removeTagContent(html, "script")
	html = removeTagContent(html, "style")

	// Simple regex-like removal of all tags
	var result strings.Builder
	inTag := false

	for _, char := range html {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// Clean up whitespace
	text := result.String()
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	text = strings.TrimSpace(text)

	return text
}

// removeTagContent removes specific tags and their content
func removeTagContent(html, tag string) string {
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	for {
		start := strings.Index(strings.ToLower(html), openTag)
		if start == -1 {
			break
		}

		end := strings.Index(strings.ToLower(html[start:]), closeTag)
		if end == -1 {
			break
		}

		html = html[:start] + html[start+end+len(closeTag):]
	}

	return html
}

// countWords counts the number of words in text
func countWords(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	return len(words)
}
