'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Book } from '@/lib/api';
import { BookReader } from './BookReader';
import { AnnotationToolbar, AnnotationPanel, AnnotationOverlay } from '@/components/annotations';
import { useAnnotations, Annotation } from '@/hooks/useAnnotations';

interface EnhancedBookReaderProps {
  book: Book;
  initialLocation?: string | number;
  fullscreen?: boolean;
  onClose?: () => void;
}

interface TextSelection {
  text: string;
  position: { x: number; y: number };
}

export const EnhancedBookReader: React.FC<EnhancedBookReaderProps> = ({
  book,
  initialLocation,
  fullscreen = false,
  onClose,
}) => {
  const [selectedText, setSelectedText] = useState<TextSelection | null>(null);
  const [showAnnotationPanel, setShowAnnotationPanel] = useState(false);
  const [showAnnotationOverlays, setShowAnnotationOverlays] = useState(true);

  const readerContainerRef = useRef<HTMLDivElement>(null);

  // Use the annotations hook
  const {
    annotations,
    createAnnotation,
    updateAnnotation,
    deleteAnnotation,
  } = useAnnotations(book.id);

  // Handle text selection in the reader
  const handleTextSelection = useCallback(() => {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    const text = selection.toString().trim();

    if (text && readerContainerRef.current) {
      const rect = range.getBoundingClientRect();
      const containerRect = readerContainerRef.current.getBoundingClientRect();

      setSelectedText({
        text,
        position: {
          x: rect.left - containerRect.left + rect.width / 2,
          y: rect.top - containerRect.top - 10,
        },
      });
    } else {
      setSelectedText(null);
    }
  }, []);

  // Handle annotation creation from toolbar
  const handleCreateAnnotation = useCallback(async (annotationData: {
    type: 'highlight' | 'note' | 'bookmark';
    selectedText: string;
    content?: string;
    color?: string;
    tags?: string[];
  }) => {
    // Calculate position based on selection
    const position = selectedText?.position || { x: 0, y: 0 };

    const newAnnotation = {
      book_id: book.id,
      type: annotationData.type,
      page_number: 0, // This would be calculated based on the reader
      start_position: 0, // This would be the character position
      end_position: annotationData.selectedText.length,
      selected_text: annotationData.selectedText,
      content: annotationData.content,
      color: annotationData.color,
      tags: annotationData.tags,
      is_private: true,
    };

    await createAnnotation(newAnnotation);
    setSelectedText(null);
  }, [createAnnotation, book.id, selectedText]);

  // Handle navigation to annotation
  const handleNavigateToAnnotation = useCallback((annotation: Annotation) => {
    // This would scroll/navigate to the annotation position in the reader
    console.log('Navigate to annotation:', annotation);
    // Implementation depends on the specific reader type
  }, []);

  // Set up text selection listener
  useEffect(() => {
    const container = readerContainerRef.current;
    if (container) {
      container.addEventListener('mouseup', handleTextSelection);
      container.addEventListener('selectionchange', handleTextSelection);
      
      return () => {
        container.removeEventListener('mouseup', handleTextSelection);
        container.removeEventListener('selectionchange', handleTextSelection);
      };
    }
  }, [handleTextSelection]);

  // Get annotations for current position (this would be more sophisticated in real implementation)
  const visibleAnnotations = annotations.filter(annotation => 
    // For demo purposes, show all annotations
    // In real implementation, this would filter by current page/position
    true
  );

  return (
    <div className="h-full flex">
      {/* Main Reader Area */}
      <div 
        ref={readerContainerRef}
        className={`flex-1 relative ${showAnnotationPanel ? 'mr-96' : ''} transition-all duration-300`}
      >
        {/* Enhanced Header Controls */}
        <div className="absolute top-0 left-0 right-0 z-30 bg-white/90 backdrop-blur border-b border-gray-200 px-4 py-2 flex items-center justify-between">
          <div className="flex items-center space-x-4">
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
              onClick={() => setShowAnnotationOverlays(!showAnnotationOverlays)}
              className={`p-2 rounded hover:bg-gray-100 ${showAnnotationOverlays ? 'bg-gray-100' : ''}`}
              title="Toggle Annotation Overlays"
            >
              👁️
            </button>
            <button
              onClick={() => setShowAnnotationPanel(!showAnnotationPanel)}
              className={`p-2 rounded hover:bg-gray-100 ${showAnnotationPanel ? 'bg-gray-100' : ''}`}
              title="Annotation Panel"
            >
              📝
            </button>
          </div>
        </div>

        {/* BookReader with annotation support */}
        <div className="h-full pt-12 relative">
          <BookReader
            book={book}
            initialLocation={initialLocation}
            fullscreen={false}
            onClose={undefined} // We handle close in the enhanced version
          />

          {/* Annotation Overlays */}
          {showAnnotationOverlays && visibleAnnotations.map((annotation) => (
            <AnnotationOverlay
              key={annotation.id}
              annotation={annotation}
              position={{
                // These would be calculated based on annotation position in the text
                x: Math.random() * 400 + 50, // Demo positioning
                y: Math.random() * 300 + 100,
                width: 200,
                height: 20,
              }}
              onUpdate={updateAnnotation}
              onDelete={deleteAnnotation}
              theme="light"
            />
          ))}

          {/* Annotation Toolbar */}
          {selectedText && (
            <AnnotationToolbar
              selectedText={selectedText.text}
              position={selectedText.position}
              onCreateAnnotation={handleCreateAnnotation}
              onClose={() => setSelectedText(null)}
              existingAnnotations={annotations}
              theme="light"
            />
          )}
        </div>
      </div>

      {/* Annotation Panel */}
      {showAnnotationPanel && (
        <div className="fixed right-0 top-0 h-full w-96 bg-white shadow-lg border-l z-40 overflow-hidden">
          <div className="h-full flex flex-col">
            {/* Panel Header */}
            <div className="p-4 border-b bg-gray-50 flex items-center justify-between">
              <h3 className="font-semibold text-gray-900">Annotations</h3>
              <button
                onClick={() => setShowAnnotationPanel(false)}
                className="text-gray-500 hover:text-gray-700 p-1"
              >
                ✕
              </button>
            </div>

            {/* Panel Content */}
            <div className="flex-1 overflow-hidden">
              <AnnotationPanel
                bookId={book.id}
                onNavigateToAnnotation={handleNavigateToAnnotation}
                theme="light"
              />
            </div>
          </div>
        </div>
      )}

      {/* Click outside handler for annotation panel */}
      {showAnnotationPanel && (
        <div
          className="fixed inset-0 z-30"
          onClick={() => setShowAnnotationPanel(false)}
        />
      )}
    </div>
  );
};

export default EnhancedBookReader;