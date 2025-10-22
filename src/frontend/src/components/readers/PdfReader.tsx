'use client';

import React, { useState, useCallback, useRef } from 'react';
import { Document, Page, pdfjs } from 'react-pdf';
import { Book } from '@/lib/api';

import 'react-pdf/dist/Page/AnnotationLayer.css';
import 'react-pdf/dist/Page/TextLayer.css';

// Configure PDF.js worker
pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.js`;

interface PdfReaderProps {
  book: Book;
  onPageChange?: (pageNumber: number, totalPages: number) => void;
  onAnnotationCreate?: (annotation: {
    type: 'highlight' | 'note' | 'bookmark';
    pageNumber: number;
    position: { x: number; y: number; width: number; height: number };
    selectedText?: string;
    content?: string;
  }) => void;
  initialPage?: number;
  theme?: 'light' | 'dark';
  zoomLevel?: number;
}

interface TextSelection {
  text: string;
  pageNumber: number;
  position: { x: number; y: number; width: number; height: number };
}

export const PdfReader: React.FC<PdfReaderProps> = ({
  book,
  onPageChange,
  onAnnotationCreate,
  initialPage = 1,
  theme = 'light',
  zoomLevel = 1.0,
}) => {
  const [numPages, setNumPages] = useState<number>(0);
  const [pageNumber, setPageNumber] = useState(initialPage);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [scale, setScale] = useState(zoomLevel);
  const [selectedText, setSelectedText] = useState<TextSelection | null>(null);
  const [showAnnotationMenu, setShowAnnotationMenu] = useState(false);
  const [file, setFile] = useState<ArrayBuffer | null>(null);

  const pageRef = useRef<HTMLDivElement>(null);

  // Load PDF document
  const onDocumentLoadSuccess = useCallback(({ numPages }: { numPages: number }) => {
    setNumPages(numPages);
    setLoading(false);
    onPageChange?.(pageNumber, numPages);
  }, [pageNumber, onPageChange]);

  const onDocumentLoadError = useCallback((error: Error) => {
    console.error('Error loading PDF:', error);
    setError('Failed to load PDF document');
    setLoading(false);
  }, []);

  // Load PDF file from API
  const loadPdfFile = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/v1/books/${book.id}/content`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to load PDF content');
      }

      const arrayBuffer = await response.arrayBuffer();
      setFile(arrayBuffer);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load PDF');
      setLoading(false);
    }
  }, [book.id]);

  // Initialize PDF loading
  React.useEffect(() => {
    loadPdfFile();
  }, [loadPdfFile]);

  // Navigation functions
  const goToPrevPage = () => {
    if (pageNumber > 1) {
      const newPage = pageNumber - 1;
      setPageNumber(newPage);
      onPageChange?.(newPage, numPages);
    }
  };

  const goToNextPage = () => {
    if (pageNumber < numPages) {
      const newPage = pageNumber + 1;
      setPageNumber(newPage);
      onPageChange?.(newPage, numPages);
    }
  };

  const goToPage = (page: number) => {
    if (page >= 1 && page <= numPages) {
      setPageNumber(page);
      onPageChange?.(page, numPages);
    }
  };

  // Zoom functions
  const zoomIn = () => {
    setScale(prev => Math.min(prev + 0.2, 3.0));
  };

  const zoomOut = () => {
    setScale(prev => Math.max(prev - 0.2, 0.5));
  };

  const resetZoom = () => {
    setScale(1.0);
  };

  // Text selection handling
  const handleTextSelection = () => {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    const text = selection.toString().trim();

    if (text && pageRef.current) {
      const rect = range.getBoundingClientRect();
      const pageRect = pageRef.current.getBoundingClientRect();

      const selectionData: TextSelection = {
        text,
        pageNumber,
        position: {
          x: rect.left - pageRect.left,
          y: rect.top - pageRect.top,
          width: rect.width,
          height: rect.height,
        }
      };

      setSelectedText(selectionData);
      setShowAnnotationMenu(true);
    }
  };

  // Annotation functions
  const createAnnotation = (type: 'highlight' | 'note' | 'bookmark') => {
    if (!selectedText && type !== 'bookmark') return;

    onAnnotationCreate?.({
      type,
      pageNumber,
      position: selectedText?.position || { x: 0, y: 0, width: 0, height: 0 },
      selectedText: type === 'bookmark' ? undefined : selectedText?.text,
      content: type === 'note' ? '' : undefined,
    });

    setShowAnnotationMenu(false);
    setSelectedText(null);
    
    // Clear selection
    window.getSelection()?.removeAllRanges();
  };

  // Keyboard navigation
  React.useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      switch (event.key) {
        case 'ArrowLeft':
          goToPrevPage();
          break;
        case 'ArrowRight':
          goToNextPage();
          break;
        case '+':
        case '=':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            zoomIn();
          }
          break;
        case '-':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            zoomOut();
          }
          break;
        case '0':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault();
            resetZoom();
          }
          break;
      }
    };

    document.addEventListener('keydown', handleKeyPress);
    return () => document.removeEventListener('keydown', handleKeyPress);
  }, [pageNumber, numPages]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading PDF...</p>
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
            onClick={loadPdfFile}
            className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className={`h-full flex flex-col ${theme === 'dark' ? 'bg-gray-900' : 'bg-gray-100'}`}>
      {/* Controls */}
      <div className={`border-b px-4 py-2 flex items-center justify-between ${
        theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'
      }`}>
        <div className="flex items-center space-x-2">
          <button
            onClick={goToPrevPage}
            disabled={pageNumber <= 1}
            className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            title="Previous Page"
          >
            ←
          </button>
          
          <div className="flex items-center space-x-2">
            <span className={`text-sm ${theme === 'dark' ? 'text-gray-300' : 'text-gray-600'}`}>
              Page
            </span>
            <input
              type="number"
              min={1}
              max={numPages}
              value={pageNumber}
              onChange={(e) => goToPage(parseInt(e.target.value))}
              className={`w-16 px-2 py-1 text-sm border rounded ${
                theme === 'dark' 
                  ? 'bg-gray-700 border-gray-600 text-gray-300' 
                  : 'bg-white border-gray-300'
              }`}
            />
            <span className={`text-sm ${theme === 'dark' ? 'text-gray-300' : 'text-gray-600'}`}>
              of {numPages}
            </span>
          </div>

          <button
            onClick={goToNextPage}
            disabled={pageNumber >= numPages}
            className="p-2 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            title="Next Page"
          >
            →
          </button>
        </div>

        <div className="flex-1 text-center">
          <h2 className={`text-sm font-medium truncate ${
            theme === 'dark' ? 'text-gray-300' : 'text-gray-900'
          }`}>
            {book.title}
          </h2>
          <p className={`text-xs ${theme === 'dark' ? 'text-gray-400' : 'text-gray-500'}`}>
            {Math.round((pageNumber / numPages) * 100)}% complete
          </p>
        </div>

        <div className="flex items-center space-x-2">
          <button
            onClick={zoomOut}
            disabled={scale <= 0.5}
            className="p-2 rounded hover:bg-gray-100 disabled:opacity-50"
            title="Zoom Out"
          >
            🔍-
          </button>
          
          <span className={`text-sm min-w-[4rem] text-center ${
            theme === 'dark' ? 'text-gray-300' : 'text-gray-600'
          }`}>
            {Math.round(scale * 100)}%
          </span>
          
          <button
            onClick={zoomIn}
            disabled={scale >= 3.0}
            className="p-2 rounded hover:bg-gray-100 disabled:opacity-50"
            title="Zoom In"
          >
            🔍+
          </button>
          
          <button
            onClick={resetZoom}
            className="p-2 rounded hover:bg-gray-100"
            title="Reset Zoom"
          >
            🔄
          </button>

          <button
            onClick={() => createAnnotation('bookmark')}
            className="p-2 rounded hover:bg-gray-100"
            title="Add Bookmark"
          >
            🔖
          </button>
        </div>
      </div>

      {/* Progress Bar */}
      <div className={`h-1 ${theme === 'dark' ? 'bg-gray-700' : 'bg-gray-200'}`}>
        <div
          className="h-full bg-indigo-600 transition-all duration-300"
          style={{ width: `${(pageNumber / numPages) * 100}%` }}
        />
      </div>

      {/* PDF Viewer */}
      <div className="flex-1 overflow-auto p-4">
        <div className="flex justify-center" ref={pageRef}>
          {file && (
            <div className="relative">
              <Document
                file={file}
                onLoadSuccess={onDocumentLoadSuccess}
                onLoadError={onDocumentLoadError}
                loading={<div>Loading page...</div>}
              >
                <Page
                  pageNumber={pageNumber}
                  scale={scale}
                  onLoadSuccess={() => {}}
                  onLoadError={() => {}}
                  className={`shadow-lg ${theme === 'dark' ? 'shadow-gray-800' : 'shadow-gray-400'}`}
                  onMouseUp={handleTextSelection}
                />
              </Document>

              {/* Annotation Menu */}
              {showAnnotationMenu && selectedText && (
                <div
                  className="absolute bg-white border border-gray-300 rounded shadow-lg p-2 z-20"
                  style={{
                    left: selectedText.position.x,
                    top: selectedText.position.y - 60,
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
                      onClick={() => setShowAnnotationMenu(false)}
                      className="px-2 py-1 text-sm text-gray-500 hover:text-gray-700"
                    >
                      ✕
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Page Navigation */}
      <div className={`border-t px-4 py-2 flex items-center justify-center space-x-2 ${
        theme === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'
      }`}>
        <button
          onClick={() => goToPage(1)}
          disabled={pageNumber === 1}
          className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50"
        >
          First
        </button>
        <button
          onClick={goToPrevPage}
          disabled={pageNumber <= 1}
          className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50"
        >
          Previous
        </button>
        <button
          onClick={goToNextPage}
          disabled={pageNumber >= numPages}
          className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50"
        >
          Next
        </button>
        <button
          onClick={() => goToPage(numPages)}
          disabled={pageNumber === numPages}
          className="px-3 py-1 text-sm border rounded hover:bg-gray-100 disabled:opacity-50"
        >
          Last
        </button>
      </div>
    </div>
  );
};

export default PdfReader;