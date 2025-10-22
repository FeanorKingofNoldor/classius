'use client';

import React, { useEffect, useRef, useState, useCallback } from 'react';
import ePub, { Book as EpubBook, Rendition } from 'epubjs';
import { Book } from '@/lib/api';

interface EpubReaderProps {
  book: Book;
  onLocationChange?: (cfi: string, percentage: number) => void;
  onAnnotationCreate?: (annotation: {
    type: 'highlight' | 'note' | 'bookmark';
    cfi: string;
    selectedText?: string;
    content?: string;
  }) => void;
  initialLocation?: string;
  theme?: 'light' | 'dark' | 'sepia';
  fontSize?: number;
  fontFamily?: string;
}

interface NavItem {
  id: string;
  label: string;
  href: string;
}

interface ReadingLocation {
  start: {
    cfi: string;
    percentage: number;
  };
  end: {
    cfi: string;
    percentage: number;
  };
}

export const EpubReader: React.FC<EpubReaderProps> = ({
  book,
  onLocationChange,
  onAnnotationCreate,
  initialLocation,
  theme = 'light',
  fontSize = 16,
  fontFamily = 'serif',
}) => {
  const viewerRef = useRef<HTMLDivElement>(null);
  const epubBookRef = useRef<EpubBook | null>(null);
  const renditionRef = useRef<Rendition | null>(null);
  
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toc, setToc] = useState<NavItem[]>([]);
  const [currentLocation, setCurrentLocation] = useState<ReadingLocation | null>(null);
  const [showToc, setShowToc] = useState(false);
  const [selectedText, setSelectedText] = useState<string>('');
  const [showAnnotationMenu, setShowAnnotationMenu] = useState(false);
  const [annotationPosition, setAnnotationPosition] = useState({ x: 0, y: 0 });

  // Load EPUB book
  const loadBook = useCallback(async () => {
    if (!viewerRef.current) return;

    try {
      setIsLoading(true);
      setError(null);

      // Get book content from API
      const response = await fetch(`/api/v1/books/${book.id}/content`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load book content');
      }

      const arrayBuffer = await response.arrayBuffer();
      const epubBook = ePub(arrayBuffer);
      epubBookRef.current = epubBook;

      // Create rendition
      const rendition = epubBook.renderTo(viewerRef.current, {
        width: '100%',
        height: '100%',
        spread: 'none',
      });
      renditionRef.current = rendition;

      // Apply theme
      await applyTheme(theme);

      // Apply font settings
      await applyFontSettings(fontSize, fontFamily);

      // Display book
      const displayed = initialLocation 
        ? await rendition.display(initialLocation)
        : await rendition.display();

      // Load table of contents
      const navigation = await epubBook.loaded.navigation;
      const tocItems: NavItem[] = navigation.toc.map((item: any) => ({
        id: item.id,
        label: item.label,
        href: item.href,
      }));
      setToc(tocItems);

      // Set up location tracking
      rendition.on('relocated', (location: any) => {
        const newLocation: ReadingLocation = {
          start: {
            cfi: location.start.cfi,
            percentage: location.start.percentage || 0,
          },
          end: {
            cfi: location.end.cfi,
            percentage: location.end.percentage || 0,
          }
        };
        setCurrentLocation(newLocation);
        onLocationChange?.(location.start.cfi, location.start.percentage || 0);
      });

      // Set up text selection handling
      rendition.on('selected', (cfiRange: string, contents: any) => {
        const selection = contents.window.getSelection();
        const text = selection.toString();
        
        if (text) {
          setSelectedText(text);
          
          // Get selection position for annotation menu
          const range = selection.getRangeAt(0);
          const rect = range.getBoundingClientRect();
          const viewerRect = viewerRef.current!.getBoundingClientRect();
          
          setAnnotationPosition({
            x: rect.left - viewerRect.left + rect.width / 2,
            y: rect.top - viewerRect.top - 10,
          });
          
          setShowAnnotationMenu(true);
        }
      });

      // Set up keyboard navigation
      rendition.on('keyup', (event: KeyboardEvent) => {
        if (event.key === 'ArrowLeft') {
          rendition.prev();
        } else if (event.key === 'ArrowRight') {
          rendition.next();
        }
      });

      setIsLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load book');
      setIsLoading(false);
    }
  }, [book.id, initialLocation, theme, fontSize, fontFamily, onLocationChange]);

  // Apply theme to rendition
  const applyTheme = async (themeName: string) => {
    if (!renditionRef.current) return;

    const themes = {
      light: {
        body: {
          'background-color': '#ffffff',
          'color': '#000000',
        }
      },
      dark: {
        body: {
          'background-color': '#1a1a1a',
          'color': '#e5e5e5',
        }
      },
      sepia: {
        body: {
          'background-color': '#f4f1ea',
          'color': '#5c4b37',
        }
      }
    };

    const themeStyles = themes[themeName];
    renditionRef.current.themes.default(themeStyles);
  };

  // Apply font settings
  const applyFontSettings = async (size: number, family: string) => {
    if (!renditionRef.current) return;

    renditionRef.current.themes.fontSize(`${size}px`);
    renditionRef.current.themes.font(family);
  };

  // Navigation functions
  const goToChapter = (href: string) => {
    renditionRef.current?.display(href);
    setShowToc(false);
  };

  const prevPage = () => {
    renditionRef.current?.prev();
  };

  const nextPage = () => {
    renditionRef.current?.next();
  };

  // Annotation functions
  const createAnnotation = (type: 'highlight' | 'note' | 'bookmark') => {
    if (!selectedText || !currentLocation) return;

    onAnnotationCreate?.({
      type,
      cfi: currentLocation.start.cfi,
      selectedText: type === 'bookmark' ? undefined : selectedText,
      content: type === 'note' ? '' : undefined, // Will be filled by user later
    });

    setShowAnnotationMenu(false);
    setSelectedText('');
  };

  // Initialize book on mount
  useEffect(() => {
    loadBook();

    return () => {
      // Cleanup
      if (renditionRef.current) {
        renditionRef.current.destroy();
      }
    };
  }, [loadBook]);

  // Update theme when it changes
  useEffect(() => {
    if (renditionRef.current) {
      applyTheme(theme);
    }
  }, [theme]);

  // Update font settings when they change
  useEffect(() => {
    if (renditionRef.current) {
      applyFontSettings(fontSize, fontFamily);
    }
  }, [fontSize, fontFamily]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading book...</p>
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
            onClick={loadBook}
            className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative h-full flex">
      {/* Table of Contents Sidebar */}
      {showToc && (
        <div className="absolute left-0 top-0 bottom-0 w-80 bg-white shadow-lg z-10 overflow-y-auto">
          <div className="p-4 border-b">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">Contents</h3>
              <button
                onClick={() => setShowToc(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>
          </div>
          <div className="p-4">
            {toc.map((item) => (
              <button
                key={item.id}
                onClick={() => goToChapter(item.href)}
                className="block w-full text-left p-2 hover:bg-gray-100 rounded"
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Main Reader */}
      <div className="flex-1 flex flex-col">
        {/* Reader Controls */}
        <div className="bg-white border-b px-4 py-2 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowToc(!showToc)}
              className="p-2 hover:bg-gray-100 rounded"
              title="Table of Contents"
            >
              📋
            </button>
            <button
              onClick={prevPage}
              className="p-2 hover:bg-gray-100 rounded"
              title="Previous Page"
            >
              ←
            </button>
            <button
              onClick={nextPage}
              className="p-2 hover:bg-gray-100 rounded"
              title="Next Page"
            >
              →
            </button>
          </div>

          <div className="flex-1 text-center">
            <h2 className="text-sm font-medium text-gray-900 truncate">
              {book.title}
            </h2>
            {currentLocation && (
              <p className="text-xs text-gray-500">
                {Math.round(currentLocation.start.percentage)}% complete
              </p>
            )}
          </div>

          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600">Progress:</span>
            <div className="w-24 bg-gray-200 rounded-full h-2">
              <div
                className="bg-indigo-600 h-2 rounded-full transition-all duration-300"
                style={{
                  width: `${currentLocation?.start.percentage || 0}%`
                }}
              />
            </div>
          </div>
        </div>

        {/* Book Content */}
        <div className="flex-1 relative">
          <div
            ref={viewerRef}
            className="w-full h-full focus:outline-none"
            tabIndex={0}
          />

          {/* Annotation Menu */}
          {showAnnotationMenu && (
            <div
              className="absolute bg-white border border-gray-300 rounded shadow-lg p-2 z-20"
              style={{
                left: annotationPosition.x,
                top: annotationPosition.y,
                transform: 'translateX(-50%) translateY(-100%)',
              }}
            >
              <div className="flex space-x-1">
                <button
                  onClick={() => createAnnotation('highlight')}
                  className="px-3 py-1 text-sm bg-yellow-200 hover:bg-yellow-300 rounded"
                  title="Highlight"
                >
                  🖍️
                </button>
                <button
                  onClick={() => createAnnotation('note')}
                  className="px-3 py-1 text-sm bg-blue-200 hover:bg-blue-300 rounded"
                  title="Add Note"
                >
                  📝
                </button>
                <button
                  onClick={() => createAnnotation('bookmark')}
                  className="px-3 py-1 text-sm bg-green-200 hover:bg-green-300 rounded"
                  title="Bookmark"
                >
                  🔖
                </button>
                <button
                  onClick={() => setShowAnnotationMenu(false)}
                  className="px-2 py-1 text-sm text-gray-500 hover:text-gray-700"
                >
                  ✕
                </button>
              </div>
            </div>
          )}

          {/* Click handlers for navigation */}
          <div
            className="absolute left-0 top-0 bottom-0 w-1/4 cursor-pointer"
            onClick={prevPage}
            title="Previous Page"
          />
          <div
            className="absolute right-0 top-0 bottom-0 w-1/4 cursor-pointer"
            onClick={nextPage}
            title="Next Page"
          />
        </div>
      </div>
    </div>
  );
};

export default EpubReader;