'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { Book, annotationsApi } from '@/lib/api';
import { EpubReader } from './EpubReader';
import { PdfReader } from './PdfReader';
import { TextReader } from './TextReader';
import { toast } from 'react-hot-toast';

interface BookReaderProps {
  book: Book;
  initialLocation?: string | number;
  fullscreen?: boolean;
  onClose?: () => void;
}

interface ReadingSettings {
  theme: 'light' | 'dark' | 'sepia';
  fontSize: number;
  fontFamily: 'serif' | 'sans-serif' | 'monospace';
  lineHeight: number;
  columnWidth: number;
}

export const BookReader: React.FC<BookReaderProps> = ({
  book,
  initialLocation,
  fullscreen = false,
  onClose,
}) => {
  const [settings, setSettings] = useState<ReadingSettings>({
    theme: 'light',
    fontSize: 16,
    fontFamily: 'serif',
    lineHeight: 1.6,
    columnWidth: 70,
  });
  const [showSettings, setShowSettings] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(fullscreen);

  // Load reading preferences from localStorage
  useEffect(() => {
    const savedSettings = localStorage.getItem('reading_settings');
    if (savedSettings) {
      try {
        const parsed = JSON.parse(savedSettings);
        setSettings(prev => ({ ...prev, ...parsed }));
      } catch (error) {
        console.error('Failed to parse saved reading settings:', error);
      }
    }
  }, []);

  // Save reading preferences
  const updateSettings = useCallback((newSettings: Partial<ReadingSettings>) => {
    const updated = { ...settings, ...newSettings };
    setSettings(updated);
    localStorage.setItem('reading_settings', JSON.stringify(updated));
  }, [settings]);

  // Handle reading position changes
  const handleLocationChange = useCallback(async (location: string | number, percentage: number) => {
    try {
      // Save reading progress to backend
      await fetch('/api/v1/progress', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify({
          book_id: book.id,
          location: location.toString(),
          percentage: percentage,
          last_read_at: new Date().toISOString(),
        }),
      });
    } catch (error) {
      console.error('Failed to save reading progress:', error);
    }
  }, [book.id]);

  // Handle annotation creation
  const handleAnnotationCreate = useCallback(async (annotation: any) => {
    try {
      const newAnnotation = {
        book_id: book.id,
        type: annotation.type,
        page_number: annotation.pageNumber || 0,
        start_position: annotation.position || annotation.start || 0,
        end_position: annotation.position || annotation.end || 0,
        selected_text: annotation.selectedText,
        content: annotation.content || '',
        is_private: true,
      };

      await annotationsApi.createAnnotation(newAnnotation);
      toast.success(`${annotation.type === 'highlight' ? 'Highlight' : annotation.type === 'note' ? 'Note' : 'Bookmark'} created successfully`);
    } catch (error) {
      console.error('Failed to create annotation:', error);
      toast.error('Failed to create annotation');
    }
  }, [book.id]);

  // Fullscreen handling
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // Only handle shortcuts when not in an input field
      if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) {
        return;
      }

      switch (event.key) {
        case 'Escape':
          if (isFullscreen) {
            toggleFullscreen();
          } else if (onClose) {
            onClose();
          }
          break;
        case 'F11':
          event.preventDefault();
          toggleFullscreen();
          break;
        case 's':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            setShowSettings(!showSettings);
          }
          break;
        case '1':
          if (event.altKey) {
            event.preventDefault();
            updateSettings({ theme: 'light' });
          }
          break;
        case '2':
          if (event.altKey) {
            event.preventDefault();
            updateSettings({ theme: 'dark' });
          }
          break;
        case '3':
          if (event.altKey) {
            event.preventDefault();
            updateSettings({ theme: 'sepia' });
          }
          break;
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, [isFullscreen, showSettings, toggleFullscreen, onClose, updateSettings]);

  // Render appropriate reader based on file type
  const renderReader = () => {
    const readerProps = {
      book,
      onLocationChange: handleLocationChange,
      onAnnotationCreate: handleAnnotationCreate,
      theme: settings.theme,
      fontSize: settings.fontSize,
      fontFamily: settings.fontFamily,
    };

    switch (book.file_type.toLowerCase()) {
      case 'epub':
        return (
          <EpubReader
            {...readerProps}
            initialLocation={typeof initialLocation === 'string' ? initialLocation : undefined}
          />
        );
      case 'pdf':
        return (
          <PdfReader
            {...readerProps}
            initialPage={typeof initialLocation === 'number' ? initialLocation : 1}
          />
        );
      case 'txt':
      default:
        return (
          <TextReader
            {...readerProps}
            initialPosition={typeof initialLocation === 'number' ? initialLocation : 0}
            lineHeight={settings.lineHeight}
            columnWidth={settings.columnWidth}
          />
        );
    }
  };

  return (
    <div className={`${isFullscreen ? 'fixed inset-0 z-50' : 'h-full'} bg-white`}>
      {/* Header Controls */}
      <div className="absolute top-0 left-0 right-0 z-30 bg-white/90 backdrop-blur border-b border-gray-200 px-4 py-2 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          {onClose && (
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 rounded"
              title="Close Reader"
            >
              ←
            </button>
          )}
          <div className="text-sm font-medium text-gray-900 truncate max-w-xs">
            {book.title}
          </div>
          <div className="text-xs text-gray-500">
            by {book.author}
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <button
            onClick={() => setShowSettings(!showSettings)}
            className={`p-2 rounded hover:bg-gray-100 ${showSettings ? 'bg-gray-100' : ''}`}
            title="Reading Settings (Ctrl+S)"
          >
            ⚙️
          </button>
          <button
            onClick={toggleFullscreen}
            className="p-2 hover:bg-gray-100 rounded"
            title="Fullscreen (F11)"
          >
            {isFullscreen ? '🗗' : '🗖'}
          </button>
        </div>
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <div className="absolute top-0 right-0 w-80 h-full bg-white border-l border-gray-200 shadow-lg z-40 overflow-y-auto">
          <div className="p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Reading Settings</h3>
              <button
                onClick={() => setShowSettings(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            {/* Theme */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Theme
              </label>
              <div className="grid grid-cols-3 gap-2">
                {[
                  { key: 'light' as const, label: 'Light', bg: 'bg-white', text: 'text-gray-900' },
                  { key: 'dark' as const, label: 'Dark', bg: 'bg-gray-900', text: 'text-white' },
                  { key: 'sepia' as const, label: 'Sepia', bg: 'bg-yellow-50', text: 'text-yellow-900' }
                ].map(theme => (
                  <button
                    key={theme.key}
                    onClick={() => updateSettings({ theme: theme.key })}
                    className={`p-3 border rounded text-sm ${theme.bg} ${theme.text} ${
                      settings.theme === theme.key ? 'ring-2 ring-indigo-500' : 'border-gray-300'
                    }`}
                  >
                    {theme.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Font Size */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Font Size: {settings.fontSize}px
              </label>
              <input
                type="range"
                min="12"
                max="32"
                value={settings.fontSize}
                onChange={(e) => updateSettings({ fontSize: parseInt(e.target.value) })}
                className="w-full"
              />
              <div className="flex justify-between text-xs text-gray-500 mt-1">
                <span>12px</span>
                <span>32px</span>
              </div>
            </div>

            {/* Font Family */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Font Family
              </label>
              <select
                value={settings.fontFamily}
                onChange={(e) => updateSettings({ fontFamily: e.target.value as any })}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="serif">Serif</option>
                <option value="sans-serif">Sans Serif</option>
                <option value="monospace">Monospace</option>
              </select>
            </div>

            {/* Line Height (for text files) */}
            {book.file_type.toLowerCase() === 'txt' && (
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Line Height: {settings.lineHeight}
                </label>
                <input
                  type="range"
                  min="1.0"
                  max="2.5"
                  step="0.1"
                  value={settings.lineHeight}
                  onChange={(e) => updateSettings({ lineHeight: parseFloat(e.target.value) })}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-gray-500 mt-1">
                  <span>1.0</span>
                  <span>2.5</span>
                </div>
              </div>
            )}

            {/* Column Width (for text files) */}
            {book.file_type.toLowerCase() === 'txt' && (
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Column Width: {settings.columnWidth} characters
                </label>
                <input
                  type="range"
                  min="50"
                  max="120"
                  value={settings.columnWidth}
                  onChange={(e) => updateSettings({ columnWidth: parseInt(e.target.value) })}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-gray-500 mt-1">
                  <span>50</span>
                  <span>120</span>
                </div>
              </div>
            )}

            {/* Keyboard Shortcuts */}
            <div className="mt-8 pt-4 border-t border-gray-200">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Keyboard Shortcuts</h4>
              <div className="text-xs text-gray-600 space-y-1">
                <div className="flex justify-between">
                  <span>Settings</span>
                  <code>Ctrl+S</code>
                </div>
                <div className="flex justify-between">
                  <span>Fullscreen</span>
                  <code>F11</code>
                </div>
                <div className="flex justify-between">
                  <span>Light Theme</span>
                  <code>Alt+1</code>
                </div>
                <div className="flex justify-between">
                  <span>Dark Theme</span>
                  <code>Alt+2</code>
                </div>
                <div className="flex justify-between">
                  <span>Sepia Theme</span>
                  <code>Alt+3</code>
                </div>
                <div className="flex justify-between">
                  <span>Close</span>
                  <code>Escape</code>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Reader Content */}
      <div className={`h-full ${showSettings ? 'mr-80' : ''} ${isFullscreen ? '' : 'pt-12'}`}>
        {renderReader()}
      </div>
    </div>
  );
};

export default BookReader;