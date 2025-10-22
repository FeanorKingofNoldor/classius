'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Annotation } from '@/hooks/useAnnotations';

interface AnnotationToolbarProps {
  selectedText: string;
  position: { x: number; y: number };
  onCreateAnnotation: (annotation: {
    type: 'highlight' | 'note' | 'bookmark';
    selectedText: string;
    content?: string;
    color?: string;
    tags?: string[];
  }) => Promise<void>;
  onClose: () => void;
  existingAnnotations?: Annotation[];
  theme?: 'light' | 'dark' | 'sepia';
}

const HIGHLIGHT_COLORS = [
  { name: 'Yellow', value: '#fef3c7', textColor: '#92400e' },
  { name: 'Green', value: '#d1fae5', textColor: '#065f46' },
  { name: 'Blue', value: '#dbeafe', textColor: '#1e40af' },
  { name: 'Purple', value: '#e9d5ff', textColor: '#7c2d12' },
  { name: 'Pink', value: '#fce7f3', textColor: '#be185d' },
  { name: 'Orange', value: '#fed7aa', textColor: '#c2410c' },
  { name: 'Red', value: '#fee2e2', textColor: '#dc2626' },
  { name: 'Gray', value: '#f3f4f6', textColor: '#374151' },
];

export const AnnotationToolbar: React.FC<AnnotationToolbarProps> = ({
  selectedText,
  position,
  onCreateAnnotation,
  onClose,
  existingAnnotations = [],
  theme = 'light',
}) => {
  const [showColorPicker, setShowColorPicker] = useState(false);
  const [showNoteForm, setShowNoteForm] = useState(false);
  const [selectedColor, setSelectedColor] = useState(HIGHLIGHT_COLORS[0]);
  const [noteContent, setNoteContent] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [creating, setCreating] = useState(false);

  const toolbarRef = useRef<HTMLDivElement>(null);
  const noteTextareaRef = useRef<HTMLTextAreaElement>(null);

  // Check if this text is already annotated
  const existingHighlight = existingAnnotations.find(
    annotation => 
      annotation.selected_text === selectedText && 
      annotation.type === 'highlight'
  );

  const existingNote = existingAnnotations.find(
    annotation => 
      annotation.selected_text === selectedText && 
      annotation.type === 'note'
  );

  // Auto-focus note textarea when shown
  useEffect(() => {
    if (showNoteForm && noteTextareaRef.current) {
      noteTextareaRef.current.focus();
    }
  }, [showNoteForm]);

  // Handle click outside to close
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (toolbarRef.current && !toolbarRef.current.contains(event.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [onClose]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      } else if (event.key === 'Enter' && event.ctrlKey && noteContent.trim()) {
        handleCreateNote();
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, [onClose, noteContent]);

  const handleCreateHighlight = async (color = selectedColor) => {
    if (creating) return;
    
    try {
      setCreating(true);
      await onCreateAnnotation({
        type: 'highlight',
        selectedText,
        color: color.value,
      });
      onClose();
    } finally {
      setCreating(false);
    }
  };

  const handleCreateNote = async () => {
    if (creating || !noteContent.trim()) return;
    
    try {
      setCreating(true);
      await onCreateAnnotation({
        type: 'note',
        selectedText,
        content: noteContent.trim(),
        color: selectedColor.value,
        tags: tags.length > 0 ? tags : undefined,
      });
      onClose();
    } finally {
      setCreating(false);
    }
  };

  const handleCreateBookmark = async () => {
    if (creating) return;
    
    try {
      setCreating(true);
      await onCreateAnnotation({
        type: 'bookmark',
        selectedText,
      });
      onClose();
    } finally {
      setCreating(false);
    }
  };

  const handleAddTag = () => {
    const tag = tagInput.trim();
    if (tag && !tags.includes(tag)) {
      setTags(prev => [...prev, tag]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(prev => prev.filter(tag => tag !== tagToRemove));
  };

  const getThemeClasses = () => {
    const themes = {
      light: 'bg-white border-gray-300 text-gray-900',
      dark: 'bg-gray-800 border-gray-600 text-gray-100',
      sepia: 'bg-yellow-50 border-yellow-300 text-yellow-900',
    };
    return themes[theme];
  };

  return (
    <div
      ref={toolbarRef}
      className={`absolute z-50 rounded-lg shadow-lg border p-3 min-w-80 ${getThemeClasses()}`}
      style={{
        left: position.x,
        top: position.y,
        transform: 'translateX(-50%) translateY(-100%)',
        marginTop: '-10px',
      }}
    >
      {/* Selected Text Preview */}
      <div className="mb-3 pb-2 border-b border-opacity-20 border-gray-400">
        <p className="text-xs opacity-60 mb-1">Selected text:</p>
        <p className="text-sm italic line-clamp-2">
          "{selectedText.length > 100 ? selectedText.slice(0, 100) + '...' : selectedText}"
        </p>
      </div>

      {!showNoteForm ? (
        <div className="space-y-3">
          {/* Quick Actions */}
          <div className="flex space-x-2">
            {!existingHighlight && (
              <button
                onClick={() => handleCreateHighlight()}
                disabled={creating}
                className="flex items-center space-x-1 px-3 py-2 rounded-md text-sm font-medium transition-colors hover:opacity-80 disabled:opacity-50"
                style={{ 
                  backgroundColor: selectedColor.value, 
                  color: selectedColor.textColor 
                }}
                title="Highlight"
              >
                <span>🖍️</span>
                <span>Highlight</span>
              </button>
            )}

            <button
              onClick={() => setShowNoteForm(true)}
              disabled={creating}
              className="flex items-center space-x-1 px-3 py-2 bg-blue-100 text-blue-700 rounded-md text-sm font-medium hover:bg-blue-200 disabled:opacity-50 transition-colors"
              title="Add Note"
            >
              <span>📝</span>
              <span>Note</span>
            </button>

            <button
              onClick={handleCreateBookmark}
              disabled={creating}
              className="flex items-center space-x-1 px-3 py-2 bg-green-100 text-green-700 rounded-md text-sm font-medium hover:bg-green-200 disabled:opacity-50 transition-colors"
              title="Bookmark"
            >
              <span>🔖</span>
              <span>Mark</span>
            </button>
          </div>

          {/* Color Picker */}
          {!existingHighlight && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium opacity-70">Highlight Color:</span>
                <button
                  onClick={() => setShowColorPicker(!showColorPicker)}
                  className="text-xs hover:underline opacity-60"
                >
                  {showColorPicker ? 'Less' : 'More colors'}
                </button>
              </div>
              
              <div className={`grid gap-1 transition-all duration-200 ${
                showColorPicker ? 'grid-cols-4' : 'grid-cols-6'
              }`}>
                {(showColorPicker ? HIGHLIGHT_COLORS : HIGHLIGHT_COLORS.slice(0, 6)).map((color) => (
                  <button
                    key={color.name}
                    onClick={() => setSelectedColor(color)}
                    className={`w-8 h-8 rounded-full border-2 transition-all hover:scale-110 ${
                      selectedColor.value === color.value 
                        ? 'border-gray-900 ring-2 ring-gray-400' 
                        : 'border-gray-300'
                    }`}
                    style={{ backgroundColor: color.value }}
                    title={color.name}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Existing Annotations Info */}
          {(existingHighlight || existingNote) && (
            <div className="text-xs opacity-60 bg-opacity-50 bg-gray-100 dark:bg-gray-700 rounded p-2">
              {existingHighlight && <p>✅ Already highlighted</p>}
              {existingNote && <p>✅ Note exists</p>}
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {/* Note Form */}
          <div>
            <label className="block text-xs font-medium mb-2 opacity-70">
              Add your note:
            </label>
            <textarea
              ref={noteTextareaRef}
              value={noteContent}
              onChange={(e) => setNoteContent(e.target.value)}
              placeholder="What do you think about this passage?"
              className={`w-full px-3 py-2 text-sm border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                theme === 'dark' 
                  ? 'bg-gray-700 border-gray-600 text-gray-100' 
                  : 'bg-white border-gray-300 text-gray-900'
              }`}
              rows={3}
            />
          </div>

          {/* Tags */}
          <div>
            <label className="block text-xs font-medium mb-2 opacity-70">
              Tags (optional):
            </label>
            <div className="flex flex-wrap gap-1 mb-2">
              {tags.map((tag) => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded-full"
                >
                  {tag}
                  <button
                    onClick={() => handleRemoveTag(tag)}
                    className="hover:text-blue-900"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleAddTag()}
                placeholder="Add tag..."
                className={`flex-1 px-2 py-1 text-xs border rounded ${
                  theme === 'dark' 
                    ? 'bg-gray-700 border-gray-600 text-gray-100' 
                    : 'bg-white border-gray-300 text-gray-900'
                }`}
              />
              <button
                onClick={handleAddTag}
                disabled={!tagInput.trim()}
                className="px-2 py-1 bg-blue-500 text-white text-xs rounded hover:bg-blue-600 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-2">
            <button
              onClick={handleCreateNote}
              disabled={creating || !noteContent.trim()}
              className="flex-1 px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              {creating ? (
                <span className="flex items-center justify-center gap-2">
                  <div className="w-3 h-3 border border-white border-t-transparent rounded-full animate-spin" />
                  Saving...
                </span>
              ) : (
                'Save Note'
              )}
            </button>
            <button
              onClick={() => setShowNoteForm(false)}
              disabled={creating}
              className="px-3 py-2 border border-gray-300 text-gray-700 text-sm rounded hover:bg-gray-50 disabled:opacity-50 transition-colors"
            >
              Back
            </button>
          </div>

          <p className="text-xs opacity-60 text-center">
            Press Ctrl+Enter to save quickly
          </p>
        </div>
      )}

      {/* Close button */}
      <button
        onClick={onClose}
        className="absolute -top-2 -right-2 w-6 h-6 bg-gray-500 text-white rounded-full text-xs hover:bg-gray-600 transition-colors"
        title="Close"
      >
        ×
      </button>
    </div>
  );
};