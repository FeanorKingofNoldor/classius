'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Book } from '@/lib/api';

interface TextReaderProps {
  book: Book;
  onLocationChange?: (position: number, percentage: number) => void;
  onAnnotationCreate?: (annotation: {
    type: 'highlight' | 'note' | 'bookmark';
    position: number;
    selectedText?: string;
    content?: string;
  }) => void;
  initialPosition?: number;
  theme?: 'light' | 'dark' | 'sepia';
  fontSize?: number;
  fontFamily?: string;
  lineHeight?: number;
  columnWidth?: number;
}

interface TextSelection {
  text: string;
  start: number;
  end: number;
  position: { x: number; y: number };
}

export const TextReader: React.FC<TextReaderProps> = ({
  book,
  onLocationChange,
  onAnnotationCreate,
  initialPosition = 0,
  theme = 'light',
  fontSize = 16,
  fontFamily = 'serif',
  lineHeight = 1.6,
  columnWidth = 70,
}) => {
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [scrollPosition, setScrollPosition] = useState(0);
  const [selectedText, setSelectedText] = useState<TextSelection | null>(null);
  const [showAnnotationMenu, setShowAnnotationMenu] = useState(false);
  const [bookmarks, setBookmarks] = useState<number[]>([]);
  const [readingMode, setReadingMode] = useState<'normal' | 'column' | 'focus'>('normal');

  const contentRef = useRef<HTMLDivElement>(null);
  const textAreaRef = useRef<HTMLDivElement>(null);

  // Load text content
  const loadTextContent = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/v1/books/${book.id}/content`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load text content');
      }

      const text = await response.text();
      setContent(text);
      setLoading(false);

      // Restore position
      if (initialPosition && textAreaRef.current) {
        setTimeout(() => {
          const element = textAreaRef.current;
          if (element) {
            const targetScroll = (initialPosition / 100) * (element.scrollHeight - element.clientHeight);
            element.scrollTop = targetScroll;
            setScrollPosition(initialPosition);
          }
        }, 100);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load text');
      setLoading(false);
    }
  }, [book.id, initialPosition]);

  // Handle scroll and track reading position
  const handleScroll = useCallback(() => {
    if (!textAreaRef.current) return;

    const element = textAreaRef.current;
    const scrollTop = element.scrollTop;
    const scrollHeight = element.scrollHeight - element.clientHeight;
    const percentage = scrollHeight > 0 ? (scrollTop / scrollHeight) * 100 : 0;

    setScrollPosition(percentage);
    onLocationChange?.(scrollTop, percentage);
  }, [onLocationChange]);

  // Handle text selection
  const handleTextSelection = useCallback(() => {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    const selectedText = selection.toString().trim();

    if (selectedText && textAreaRef.current) {
      // Calculate relative position within the text
      const preSelectionRange = range.cloneRange();
      preSelectionRange.selectNodeContents(textAreaRef.current);
      preSelectionRange.setEnd(range.startContainer, range.startOffset);
      const start = preSelectionRange.toString().length;
      const end = start + selectedText.length;

      // Get selection position for menu
      const rect = range.getBoundingClientRect();
      const containerRect = textAreaRef.current.getBoundingClientRect();

      const selectionData: TextSelection = {
        text: selectedText,
        start,
        end,
        position: {
          x: rect.left - containerRect.left + rect.width / 2,
          y: rect.top - containerRect.top - 10,
        },
      };

      setSelectedText(selectionData);
      setShowAnnotationMenu(true);
    }
  }, []);

  // Annotation functions
  const createAnnotation = (type: 'highlight' | 'note' | 'bookmark') => {
    if (!selectedText && type !== 'bookmark') return;

    const position = selectedText ? selectedText.start : scrollPosition;

    onAnnotationCreate?.({
      type,
      position,
      selectedText: type === 'bookmark' ? undefined : selectedText?.text,
      content: type === 'note' ? '' : undefined,
    });

    if (type === 'bookmark') {
      setBookmarks(prev => [...prev, Math.round(scrollPosition)]);
    }

    setShowAnnotationMenu(false);
    setSelectedText(null);
    window.getSelection()?.removeAllRanges();
  };

  // Navigation functions
  const goToPosition = (percentage: number) => {
    if (!textAreaRef.current) return;

    const element = textAreaRef.current;
    const targetScroll = (percentage / 100) * (element.scrollHeight - element.clientHeight);
    element.scrollTo({ top: targetScroll, behavior: 'smooth' });
  };

  const goToBookmark = (position: number) => {
    goToPosition(position);
  };

  // Theme configurations
  const getThemeClasses = () => {
    const themes = {
      light: 'bg-white text-gray-900',
      dark: 'bg-gray-900 text-gray-100',
      sepia: 'bg-yellow-50 text-yellow-900',
    };
    return themes[theme];
  };

  const getReadingModeClasses = () => {
    switch (readingMode) {
      case 'column':
        return 'columns-2 gap-8';
      case 'focus':
        return 'max-w-2xl mx-auto';
      default:
        return '';
    }
  };

  // Initialize component
  useEffect(() => {
    loadTextContent();
  }, [loadTextContent]);

  // Add scroll listener
  useEffect(() => {
    const element = textAreaRef.current;
    if (element) {
      element.addEventListener('scroll', handleScroll);
      return () => element.removeEventListener('scroll', handleScroll);
    }
  }, [handleScroll]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (!textAreaRef.current) return;

      switch (event.key) {
        case 'ArrowUp':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            textAreaRef.current.scrollBy({ top: -100, behavior: 'smooth' });
          }
          break;
        case 'ArrowDown':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            textAreaRef.current.scrollBy({ top: 100, behavior: 'smooth' });
          }
          break;
        case 'Home':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            goToPosition(0);
          }
          break;
        case 'End':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            goToPosition(100);
          }
          break;
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading text...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-red-500 mb-4">❌</div>
          <p className="text-red-600">{error}</p>
          <button
            onClick={loadTextContent}
            className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className={`h-full flex flex-col ${getThemeClasses()}`}>
      {/* Controls */}
      <div className={`border-b px-4 py-2 flex items-center justify-between ${
        theme === 'dark' ? 'border-gray-700' : 'border-gray-200'
      }`}>
        <div className="flex items-center space-x-2">
          <select
            value={readingMode}
            onChange={(e) => setReadingMode(e.target.value as 'normal' | 'column' | 'focus')}
            className={`px-2 py-1 text-sm border rounded ${
              theme === 'dark' 
                ? 'bg-gray-800 border-gray-600 text-gray-300' 
                : 'bg-white border-gray-300'
            }`}
          >
            <option value="normal">Normal</option>
            <option value="column">Two Columns</option>
            <option value="focus">Focus Mode</option>
          </select>

          <button
            onClick={() => createAnnotation('bookmark')}
            className={`p-2 rounded hover:opacity-80 ${
              theme === 'dark' ? 'hover:bg-gray-800' : 'hover:bg-gray-100'
            }`}
            title="Add Bookmark"
          >
            🔖
          </button>

          {bookmarks.length > 0 && (
            <select
              onChange={(e) => goToBookmark(Number(e.target.value))}
              className={`px-2 py-1 text-sm border rounded ${
                theme === 'dark' 
                  ? 'bg-gray-800 border-gray-600 text-gray-300' 
                  : 'bg-white border-gray-300'
              }`}
              defaultValue=""
            >
              <option value="" disabled>Go to bookmark</option>
              {bookmarks.map((bookmark, index) => (
                <option key={index} value={bookmark}>
                  Bookmark {index + 1} ({Math.round(bookmark)}%)
                </option>
              ))}
            </select>
          )}
        </div>

        <div className="flex-1 text-center">
          <h2 className="text-sm font-medium truncate">
            {book.title}
          </h2>
          <p className={`text-xs ${theme === 'dark' ? 'text-gray-400' : 'text-gray-500'}`}>
            {Math.round(scrollPosition)}% complete
          </p>
        </div>

        <div className="flex items-center space-x-2">
          <span className={`text-sm ${theme === 'dark' ? 'text-gray-400' : 'text-gray-600'}`}>
            Progress:
          </span>
          <div className={`w-24 rounded-full h-2 ${
            theme === 'dark' ? 'bg-gray-700' : 'bg-gray-200'
          }`}>
            <div
              className="bg-indigo-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${scrollPosition}%` }}
            />
          </div>
          <button
            onClick={() => goToPosition(0)}
            className={`text-xs px-2 py-1 rounded ${
              theme === 'dark' ? 'hover:bg-gray-800' : 'hover:bg-gray-100'
            }`}
          >
            Top
          </button>
        </div>
      </div>

      {/* Text Content */}
      <div className="flex-1 relative">
        <div
          ref={textAreaRef}
          className="h-full overflow-auto px-8 py-6 focus:outline-none"
          tabIndex={0}
          onMouseUp={handleTextSelection}
        >
          <div
            ref={contentRef}
            className={`${getReadingModeClasses()} prose prose-lg max-w-none`}
            style={{
              fontSize: `${fontSize}px`,
              fontFamily,
              lineHeight,
              maxWidth: readingMode === 'focus' ? `${columnWidth}ch` : 'none',
            }}
          >
            <pre className="whitespace-pre-wrap font-inherit leading-inherit">
              {content}
            </pre>
          </div>
        </div>

        {/* Annotation Menu */}
        {showAnnotationMenu && selectedText && (
          <div
            className={`absolute border rounded shadow-lg p-2 z-20 ${
              theme === 'dark' 
                ? 'bg-gray-800 border-gray-600' 
                : 'bg-white border-gray-300'
            }`}
            style={{
              left: selectedText.position.x,
              top: selectedText.position.y,
              transform: 'translateX(-50%) translateY(-100%)',
            }}
          >
            <div className="flex space-x-1">
              <button
                onClick={() => createAnnotation('highlight')}
                className="px-3 py-1 text-sm bg-yellow-200 hover:bg-yellow-300 rounded text-black"
                title="Highlight"
              >
                🖍️
              </button>
              <button
                onClick={() => createAnnotation('note')}
                className="px-3 py-1 text-sm bg-blue-200 hover:bg-blue-300 rounded text-black"
                title="Add Note"
              >
                📝
              </button>
              <button
                onClick={() => createAnnotation('bookmark')}
                className="px-3 py-1 text-sm bg-green-200 hover:bg-green-300 rounded text-black"
                title="Bookmark"
              >
                🔖
              </button>
              <button
                onClick={() => setShowAnnotationMenu(false)}
                className={`px-2 py-1 text-sm hover:opacity-70 ${
                  theme === 'dark' ? 'text-gray-300' : 'text-gray-500'
                }`}
              >
                ✕
              </button>
            </div>
          </div>
        )}

        {/* Scroll Progress Indicator */}
        <div className={`absolute right-0 top-0 w-1 h-full ${
          theme === 'dark' ? 'bg-gray-700' : 'bg-gray-200'
        }`}>
          <div
            className="w-full bg-indigo-600 transition-all duration-300"
            style={{ height: `${scrollPosition}%` }}
          />
        </div>
      </div>
    </div>
  );
};

export default TextReader;