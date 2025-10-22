-- Migration: 002_create_user_progress_tables.sql
-- Description: Create tables for user books, reading progress, and related structures

-- Create user_books table (many-to-many relationship between users and books)
CREATE TABLE IF NOT EXISTS user_books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    source VARCHAR(50) DEFAULT 'manual' CHECK (source IN ('manual', 'purchased', 'shared', 'imported')),
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    UNIQUE(user_id, book_id)
);

-- Create reading_progress table
CREATE TABLE IF NOT EXISTS reading_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    current_page INTEGER DEFAULT 0,
    total_pages INTEGER DEFAULT 0,
    current_position INTEGER DEFAULT 0,
    percentage DECIMAL(5,2) DEFAULT 0.0 CHECK (percentage >= 0.0 AND percentage <= 100.0),
    time_spent_minutes INTEGER DEFAULT 0,
    last_read TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reading_streak_days INTEGER DEFAULT 0,
    notes_count INTEGER DEFAULT 0,
    highlights_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    UNIQUE(user_id, book_id)
);

-- Create bookmarks table
CREATE TABLE IF NOT EXISTS bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    name VARCHAR(255),
    page_number INTEGER NOT NULL,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create reading_sessions table for tracking reading activity
CREATE TABLE IF NOT EXISTS reading_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    duration_minutes INTEGER DEFAULT 0,
    pages_read INTEGER DEFAULT 0,
    start_page INTEGER,
    end_page INTEGER,
    start_position INTEGER DEFAULT 0,
    end_position INTEGER DEFAULT 0,
    device_type VARCHAR(50), -- web, mobile, tablet, e-reader
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_books_user_id ON user_books(user_id);
CREATE INDEX IF NOT EXISTS idx_user_books_book_id ON user_books(book_id);
CREATE INDEX IF NOT EXISTS idx_user_books_added_at ON user_books(added_at);
CREATE INDEX IF NOT EXISTS idx_user_books_deleted_at ON user_books(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_book_id ON reading_progress(book_id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_last_read ON reading_progress(last_read);
CREATE INDEX IF NOT EXISTS idx_reading_progress_percentage ON reading_progress(percentage);
CREATE INDEX IF NOT EXISTS idx_reading_progress_deleted_at ON reading_progress(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_book_id ON bookmarks(book_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_page_number ON bookmarks(page_number);
CREATE INDEX IF NOT EXISTS idx_bookmarks_deleted_at ON bookmarks(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reading_sessions_user_id ON reading_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_book_id ON reading_sessions(book_id);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_started_at ON reading_sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_reading_sessions_ended_at ON reading_sessions(ended_at);

-- Add triggers for updated_at
CREATE TRIGGER update_user_books_updated_at 
    BEFORE UPDATE ON user_books
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reading_progress_updated_at 
    BEFORE UPDATE ON reading_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookmarks_updated_at 
    BEFORE UPDATE ON bookmarks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reading_sessions_updated_at 
    BEFORE UPDATE ON reading_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create useful views for statistics
CREATE OR REPLACE VIEW user_reading_stats AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(DISTINCT ub.book_id) as total_books,
    COUNT(DISTINCT CASE WHEN rp.percentage >= 100 THEN ub.book_id END) as books_completed,
    COUNT(DISTINCT CASE WHEN rp.percentage > 0 AND rp.percentage < 100 THEN ub.book_id END) as books_in_progress,
    COALESCE(SUM(rp.time_spent_minutes), 0) as total_reading_time_minutes,
    COALESCE(SUM(rp.current_page), 0) as total_pages_read,
    MAX(rp.reading_streak_days) as longest_reading_streak,
    AVG(CASE WHEN rp.time_spent_minutes > 0 THEN rp.time_spent_minutes END) as avg_reading_time_per_book,
    COUNT(DISTINCT rs.id) as total_reading_sessions,
    MAX(rp.last_read) as last_reading_activity
FROM users u
LEFT JOIN user_books ub ON u.id = ub.user_id AND ub.deleted_at IS NULL
LEFT JOIN reading_progress rp ON u.id = rp.user_id AND rp.deleted_at IS NULL
LEFT JOIN reading_sessions rs ON u.id = rs.user_id
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username;