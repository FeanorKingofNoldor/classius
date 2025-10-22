-- Migration: 004_create_annotations_and_sage.sql
-- Description: Create tables for annotations, AI Sage conversations, and community features

-- Create annotations table
CREATE TABLE IF NOT EXISTS annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('highlight', 'note', 'bookmark')),
    page_number INTEGER,
    start_position INTEGER,
    end_position INTEGER,
    selected_text TEXT,
    content TEXT, -- Note content for notes
    color VARCHAR(20), -- For highlights (#ffff00, etc.)
    tags TEXT[], -- Array of tag strings
    is_private BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create sage_conversations table for AI interactions
CREATE TABLE IF NOT EXISTS sage_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE SET NULL,
    passage_text TEXT,
    question TEXT NOT NULL,
    response TEXT NOT NULL,
    context_data JSONB, -- Additional context as JSON
    response_time_ms INTEGER,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5), -- 1-5 star rating
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create reading_groups table for book clubs and study groups
CREATE TABLE IF NOT EXISTS reading_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    group_type VARCHAR(50) DEFAULT 'book_club' CHECK (group_type IN ('book_club', 'study_group', 'discussion')),
    is_public BOOLEAN DEFAULT TRUE,
    max_members INTEGER DEFAULT 50,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create group_members table
CREATE TABLE IF NOT EXISTS group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES reading_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'moderator', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    UNIQUE(group_id, user_id)
);

-- Create discussions table for group discussions and comments
CREATE TABLE IF NOT EXISTS discussions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES reading_groups(id) ON DELETE CASCADE,
    book_id UUID REFERENCES books(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    content TEXT NOT NULL,
    passage_reference TEXT, -- Reference to specific passage
    parent_id UUID REFERENCES discussions(id) ON DELETE CASCADE, -- For threaded replies
    upvotes INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create published_notes table for expert annotations
CREATE TABLE IF NOT EXISTS published_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    difficulty_level VARCHAR(20) DEFAULT 'intermediate' CHECK (difficulty_level IN ('beginner', 'intermediate', 'advanced', 'expert')),
    price_cents INTEGER DEFAULT 0, -- 0 for free
    is_verified BOOLEAN DEFAULT FALSE, -- For verified scholars
    rating DECIMAL(3,2) DEFAULT 0.0 CHECK (rating >= 0.0 AND rating <= 5.0),
    downloads_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create note_overlays table for expert annotation overlays
CREATE TABLE IF NOT EXISTS note_overlays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published_note_id UUID NOT NULL REFERENCES published_notes(id) ON DELETE CASCADE,
    page_number INTEGER,
    start_position INTEGER,
    end_position INTEGER,
    note_type VARCHAR(50) CHECK (note_type IN ('explanation', 'context', 'cross_reference', 'translation', 'historical_note')),
    content TEXT NOT NULL,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_annotations_user_id ON annotations(user_id);
CREATE INDEX IF NOT EXISTS idx_annotations_book_id ON annotations(book_id);
CREATE INDEX IF NOT EXISTS idx_annotations_type ON annotations(type);
CREATE INDEX IF NOT EXISTS idx_annotations_page_number ON annotations(page_number);
CREATE INDEX IF NOT EXISTS idx_annotations_is_private ON annotations(is_private);
CREATE INDEX IF NOT EXISTS idx_annotations_created_at ON annotations(created_at);
CREATE INDEX IF NOT EXISTS idx_annotations_deleted_at ON annotations(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sage_conversations_user_id ON sage_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_sage_conversations_book_id ON sage_conversations(book_id);
CREATE INDEX IF NOT EXISTS idx_sage_conversations_rating ON sage_conversations(rating);
CREATE INDEX IF NOT EXISTS idx_sage_conversations_created_at ON sage_conversations(created_at);
CREATE INDEX IF NOT EXISTS idx_sage_conversations_deleted_at ON sage_conversations(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reading_groups_created_by ON reading_groups(created_by);
CREATE INDEX IF NOT EXISTS idx_reading_groups_is_public ON reading_groups(is_public);
CREATE INDEX IF NOT EXISTS idx_reading_groups_group_type ON reading_groups(group_type);
CREATE INDEX IF NOT EXISTS idx_reading_groups_deleted_at ON reading_groups(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_group_members_group_id ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_role ON group_members(role);
CREATE INDEX IF NOT EXISTS idx_group_members_deleted_at ON group_members(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_discussions_group_id ON discussions(group_id);
CREATE INDEX IF NOT EXISTS idx_discussions_book_id ON discussions(book_id);
CREATE INDEX IF NOT EXISTS idx_discussions_user_id ON discussions(user_id);
CREATE INDEX IF NOT EXISTS idx_discussions_parent_id ON discussions(parent_id);
CREATE INDEX IF NOT EXISTS idx_discussions_upvotes ON discussions(upvotes);
CREATE INDEX IF NOT EXISTS idx_discussions_created_at ON discussions(created_at);
CREATE INDEX IF NOT EXISTS idx_discussions_deleted_at ON discussions(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_published_notes_author_id ON published_notes(author_id);
CREATE INDEX IF NOT EXISTS idx_published_notes_book_id ON published_notes(book_id);
CREATE INDEX IF NOT EXISTS idx_published_notes_is_verified ON published_notes(is_verified);
CREATE INDEX IF NOT EXISTS idx_published_notes_rating ON published_notes(rating);
CREATE INDEX IF NOT EXISTS idx_published_notes_difficulty_level ON published_notes(difficulty_level);
CREATE INDEX IF NOT EXISTS idx_published_notes_deleted_at ON published_notes(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_note_overlays_published_note_id ON note_overlays(published_note_id);
CREATE INDEX IF NOT EXISTS idx_note_overlays_page_number ON note_overlays(page_number);
CREATE INDEX IF NOT EXISTS idx_note_overlays_note_type ON note_overlays(note_type);
CREATE INDEX IF NOT EXISTS idx_note_overlays_display_order ON note_overlays(display_order);
CREATE INDEX IF NOT EXISTS idx_note_overlays_deleted_at ON note_overlays(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add GIN index for tags array search
CREATE INDEX IF NOT EXISTS idx_annotations_tags ON annotations USING gin(tags);

-- Add full-text search index for annotations
CREATE INDEX IF NOT EXISTS idx_annotations_fulltext ON annotations 
USING gin(to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(selected_text, '')));

-- Add full-text search index for sage conversations
CREATE INDEX IF NOT EXISTS idx_sage_conversations_fulltext ON sage_conversations 
USING gin(to_tsvector('english', question || ' ' || response));

-- Add triggers for updated_at
CREATE TRIGGER update_annotations_updated_at 
    BEFORE UPDATE ON annotations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sage_conversations_updated_at 
    BEFORE UPDATE ON sage_conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reading_groups_updated_at 
    BEFORE UPDATE ON reading_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_group_members_updated_at 
    BEFORE UPDATE ON group_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_discussions_updated_at 
    BEFORE UPDATE ON discussions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_published_notes_updated_at 
    BEFORE UPDATE ON published_notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_note_overlays_updated_at 
    BEFORE UPDATE ON note_overlays
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create useful views for annotations
CREATE OR REPLACE VIEW user_annotation_stats AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(a.id) as total_annotations,
    COUNT(CASE WHEN a.type = 'highlight' THEN 1 END) as highlights,
    COUNT(CASE WHEN a.type = 'note' THEN 1 END) as notes,
    COUNT(CASE WHEN a.type = 'bookmark' THEN 1 END) as bookmarks,
    COUNT(DISTINCT a.book_id) as books_with_annotations,
    MAX(a.created_at) as last_annotation_date
FROM users u
LEFT JOIN annotations a ON u.id = a.user_id AND a.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username;