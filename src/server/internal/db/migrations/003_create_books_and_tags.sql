-- Migration: 003_create_books_and_tags.sql
-- Description: Create tables for books, tags, and book-tag relationships

-- Create books table
CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    author VARCHAR(300) NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    genre VARCHAR(100),
    publisher VARCHAR(200),
    published_at TIMESTAMP,
    isbn VARCHAR(20),
    description TEXT,
    cover_url TEXT,
    file_url TEXT,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    page_count INTEGER,
    word_count INTEGER,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'archived', 'processing', 'error')),
    is_public BOOLEAN DEFAULT FALSE,
    
    -- Embedded metadata fields
    original_file_name VARCHAR(500),
    mime_type VARCHAR(100),
    encoding VARCHAR(50),
    has_images BOOLEAN DEFAULT FALSE,
    has_toc BOOLEAN DEFAULT FALSE,
    chapter_count INTEGER,
    
    -- Format information
    version VARCHAR(50),
    drm BOOLEAN DEFAULT FALSE,
    interactive BOOLEAN DEFAULT FALSE,
    rights TEXT,
    source VARCHAR(200),
    extra_data JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7), -- Hex color code
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create book_tags junction table
CREATE TABLE IF NOT EXISTS book_tags (
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, tag_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_genre ON books(genre);
CREATE INDEX IF NOT EXISTS idx_books_language ON books(language);
CREATE INDEX IF NOT EXISTS idx_books_status ON books(status);
CREATE INDEX IF NOT EXISTS idx_books_created_at ON books(created_at);
CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);
CREATE INDEX IF NOT EXISTS idx_books_deleted_at ON books(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_book_tags_book_id ON book_tags(book_id);
CREATE INDEX IF NOT EXISTS idx_book_tags_tag_id ON book_tags(tag_id);

-- Create unique constraint for tag names per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_user_name_unique ON tags(user_id, name) 
WHERE deleted_at IS NULL;

-- Create full-text search index for books (PostgreSQL specific)
CREATE INDEX IF NOT EXISTS idx_books_fulltext_search ON books 
USING gin(to_tsvector('english', title || ' ' || author || ' ' || COALESCE(description, '')));

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_books_updated_at 
    BEFORE UPDATE ON books
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tags_updated_at 
    BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add some useful views
CREATE OR REPLACE VIEW user_book_stats AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(b.id) as total_books,
    COALESCE(SUM(b.file_size), 0) as total_size_bytes,
    COUNT(CASE WHEN b.status = 'active' THEN 1 END) as active_books,
    COUNT(CASE WHEN b.status = 'archived' THEN 1 END) as archived_books,
    COUNT(CASE WHEN b.status = 'processing' THEN 1 END) as processing_books,
    COUNT(CASE WHEN b.status = 'error' THEN 1 END) as error_books,
    MAX(b.created_at) as last_book_added
FROM users u
LEFT JOIN books b ON u.id = b.user_id AND b.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username;

-- Create a view for book search with tag information
CREATE OR REPLACE VIEW books_with_tags AS
SELECT 
    b.*,
    ARRAY_AGG(t.name ORDER BY t.name) FILTER (WHERE t.name IS NOT NULL) as tag_names,
    ARRAY_AGG(t.id ORDER BY t.name) FILTER (WHERE t.id IS NOT NULL) as tag_ids
FROM books b
LEFT JOIN book_tags bt ON b.id = bt.book_id
LEFT JOIN tags t ON bt.tag_id = t.id AND t.deleted_at IS NULL
WHERE b.deleted_at IS NULL
GROUP BY b.id;