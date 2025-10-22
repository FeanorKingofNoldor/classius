-- Classius Database Initialization Script
-- Creates all necessary tables for the Classius server

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    subscription_tier VARCHAR(20) DEFAULT 'free',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP WITH TIME ZONE
);

-- Books table (master catalog)
CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    author VARCHAR(255),
    language VARCHAR(10),
    publisher VARCHAR(255),
    isbn VARCHAR(20),
    description TEXT,
    cover_url TEXT,
    file_path TEXT,
    file_format VARCHAR(20),
    page_count INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    difficulty_level VARCHAR(20),
    tradition VARCHAR(50), -- Western, Eastern, Islamic, etc.
    time_period VARCHAR(50), -- Ancient, Medieval, Modern, etc.
    genre VARCHAR(100),
    tags TEXT[], -- Array of tags
    is_public BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User library (books owned/accessed by users)
CREATE TABLE IF NOT EXISTS user_books (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    source VARCHAR(50) DEFAULT 'manual', -- manual, purchased, shared
    UNIQUE(user_id, book_id)
);

-- Reading progress
CREATE TABLE IF NOT EXISTS reading_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE CASCADE,
    current_page INTEGER DEFAULT 0,
    total_pages INTEGER DEFAULT 0,
    current_position INTEGER DEFAULT 0,
    percentage DECIMAL(5,4) DEFAULT 0.0,
    time_spent_minutes INTEGER DEFAULT 0,
    last_read TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    reading_streak_days INTEGER DEFAULT 0,
    notes_count INTEGER DEFAULT 0,
    highlights_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, book_id)
);

-- Annotations (highlights, notes, bookmarks)
CREATE TABLE IF NOT EXISTS annotations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- highlight, note, bookmark
    page_number INTEGER,
    start_position INTEGER,
    end_position INTEGER,
    selected_text TEXT,
    content TEXT, -- Note content
    color VARCHAR(20), -- For highlights
    tags TEXT[], -- Array of tags
    is_private BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bookmarks (special type of annotation)
CREATE TABLE IF NOT EXISTS bookmarks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE CASCADE,
    name VARCHAR(255),
    page_number INTEGER NOT NULL,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- AI Sage conversations
CREATE TABLE IF NOT EXISTS sage_conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id),
    passage_text TEXT,
    question TEXT NOT NULL,
    response TEXT NOT NULL,
    context_data JSONB, -- Additional context (reading history, etc.)
    response_time_ms INTEGER,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Reading groups/clubs
CREATE TABLE IF NOT EXISTS reading_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    group_type VARCHAR(50) DEFAULT 'book_club', -- book_club, study_group, discussion
    is_public BOOLEAN DEFAULT true,
    max_members INTEGER DEFAULT 50,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Group memberships
CREATE TABLE IF NOT EXISTS group_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID REFERENCES reading_groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- admin, moderator, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

-- Group discussions
CREATE TABLE IF NOT EXISTS discussions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID REFERENCES reading_groups(id),
    book_id UUID REFERENCES books(id),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255),
    content TEXT NOT NULL,
    passage_reference TEXT, -- Reference to specific passage
    parent_id UUID REFERENCES discussions(id), -- For threaded replies
    upvotes INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Note publishing system
CREATE TABLE IF NOT EXISTS published_notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id UUID REFERENCES users(id),
    book_id UUID REFERENCES books(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    difficulty_level VARCHAR(20) DEFAULT 'intermediate',
    price_cents INTEGER DEFAULT 0, -- 0 for free
    is_verified BOOLEAN DEFAULT false, -- For verified scholars
    rating DECIMAL(3,2) DEFAULT 0.0,
    downloads_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Expert note overlays
CREATE TABLE IF NOT EXISTS note_overlays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    published_note_id UUID REFERENCES published_notes(id) ON DELETE CASCADE,
    page_number INTEGER,
    start_position INTEGER,
    end_position INTEGER,
    note_type VARCHAR(50), -- explanation, context, cross_reference
    content TEXT NOT NULL,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User sessions (for device sync)
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(255),
    access_token VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_tradition ON books(tradition);
CREATE INDEX IF NOT EXISTS idx_user_books_user_id ON user_books(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_progress_book_id ON reading_progress(book_id);
CREATE INDEX IF NOT EXISTS idx_annotations_user_book ON annotations(user_id, book_id);
CREATE INDEX IF NOT EXISTS idx_sage_conversations_user_id ON sage_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_discussions_group_id ON discussions(group_id);
CREATE INDEX IF NOT EXISTS idx_discussions_book_id ON discussions(book_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_device_id ON user_sessions(device_id);

-- Insert some sample data for development
INSERT INTO books (title, author, language, tradition, time_period, genre, is_public) VALUES
('The Republic', 'Plato', 'en', 'Western', 'Ancient', 'Philosophy', true),
('Meditations', 'Marcus Aurelius', 'en', 'Western', 'Ancient', 'Philosophy', true),
('The Iliad', 'Homer', 'en', 'Western', 'Ancient', 'Epic Poetry', true),
('The Odyssey', 'Homer', 'en', 'Western', 'Ancient', 'Epic Poetry', true),
('Nicomachean Ethics', 'Aristotle', 'en', 'Western', 'Ancient', 'Philosophy', true),
('Tao Te Ching', 'Laozi', 'en', 'Eastern', 'Ancient', 'Philosophy', true),
('The Analects', 'Confucius', 'en', 'Eastern', 'Ancient', 'Philosophy', true),
('The Quran', 'Various', 'en', 'Islamic', 'Medieval', 'Religious Text', true);

-- Create a demo user
INSERT INTO users (username, email, password_hash, full_name) VALUES
('demo_user', 'demo@classius.com', '$2b$12$demo_hash_placeholder', 'Demo User');

-- Add demo books to demo user's library
INSERT INTO user_books (user_id, book_id) 
SELECT u.id, b.id 
FROM users u, books b 
WHERE u.username = 'demo_user' AND b.title IN ('The Republic', 'Meditations', 'The Iliad');

-- Add some sample reading progress
INSERT INTO reading_progress (user_id, book_id, current_page, total_pages, percentage) 
SELECT u.id, b.id, 25, 100, 0.25
FROM users u, books b 
WHERE u.username = 'demo_user' AND b.title = 'The Republic';

-- Success message
SELECT 'Classius database initialized successfully!' as message;