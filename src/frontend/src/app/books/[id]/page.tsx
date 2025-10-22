'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface Book {
  id: string;
  title: string;
  author: string;
  language: string;
  genre?: string;
  publisher?: string;
  published_at?: string;
  isbn?: string;
  description?: string;
  cover_url?: string;
  file_type: string;
  file_path: string;
  file_size: number;
  page_count?: number;
  word_count?: number;
  status: string;
  is_public: boolean;
  tags?: Tag[];
  created_at: string;
  updated_at: string;
}

interface Tag {
  id: string;
  name: string;
  color?: string;
}

interface Annotation {
  id: string;
  book_id: string;
  user_id: string;
  content: string;
  note?: string;
  start_position: number;
  end_position: number;
  page_number?: number;
  created_at: string;
  updated_at: string;
}

export default function BookViewerPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();
  const params = useParams();
  const bookId = params.id as string;

  const [book, setBook] = useState<Book | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [annotations, setAnnotations] = useState<Annotation[]>([]);
  const [showSidebar, setShowSidebar] = useState(false);
  const [sidebarTab, setSidebarTab] = useState<'info' | 'annotations' | 'sage'>('info');
  const [selectedText, setSelectedText] = useState('');
  const [showAnnotationForm, setShowAnnotationForm] = useState(false);
  const [annotationNote, setAnnotationNote] = useState('');
  const [sageQuery, setSageQuery] = useState('');
  const [sageResponse, setSageResponse] = useState('');
  const [sageLoading, setSageLoading] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    if (bookId) {
      fetchBook();
      fetchAnnotations();
    }
  }, [isAuthenticated, router, bookId]);

  const fetchBook = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('No authentication token');
      }

      const response = await fetch(`http://localhost:8080/api/books/${bookId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Book not found');
        }
        throw new Error('Failed to fetch book');
      }

      const result = await response.json();
      setBook(result.data);
      setError('');
    } catch (err) {
      console.error('Error fetching book:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch book');
      toast.error(err instanceof Error ? err.message : 'Failed to fetch book');
    } finally {
      setLoading(false);
    }
  };

  const fetchAnnotations = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch(`http://localhost:8080/api/annotations?book_id=${bookId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setAnnotations(result.data || []);
      }
    } catch (err) {
      console.error('Error fetching annotations:', err);
    }
  };

  const handleTextSelection = () => {
    const selection = window.getSelection();
    if (selection && selection.toString().length > 0) {
      setSelectedText(selection.toString());
      setShowAnnotationForm(true);
      setShowSidebar(true);
      setSidebarTab('annotations');
    }
  };

  const saveAnnotation = async () => {
    if (!selectedText.trim()) {
      toast.error('Please select some text first');
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const annotationData = {
        book_id: bookId,
        content: selectedText,
        note: annotationNote.trim() || undefined,
        start_position: 0, // This would need proper implementation
        end_position: selectedText.length,
      };

      const response = await fetch('http://localhost:8080/api/annotations', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(annotationData),
      });

      if (!response.ok) {
        throw new Error('Failed to save annotation');
      }

      toast.success('Annotation saved!');
      setShowAnnotationForm(false);
      setSelectedText('');
      setAnnotationNote('');
      fetchAnnotations(); // Refresh annotations
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to save annotation');
    }
  };

  const askSage = async () => {
    if (!sageQuery.trim()) {
      toast.error('Please enter a question for the AI Sage');
      return;
    }

    try {
      setSageLoading(true);
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const sageData = {
        book_id: bookId,
        query: sageQuery.trim(),
        context: selectedText || undefined,
      };

      const response = await fetch('http://localhost:8080/api/sage/query', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(sageData),
      });

      if (!response.ok) {
        throw new Error('Failed to get response from AI Sage');
      }

      const result = await response.json();
      setSageResponse(result.data.response);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to get AI Sage response');
    } finally {
      setSageLoading(false);
    }
  };

  const downloadBook = () => {
    if (!book) return;
    const downloadUrl = `http://localhost:8080/api/books/${bookId}/download`;
    const token = localStorage.getItem('token');
    
    // Create a temporary link with authentication
    fetch(downloadUrl, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })
    .then(response => response.blob())
    .then(blob => {
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${book.title} - ${book.author}.${book.file_type}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    })
    .catch(err => {
      toast.error('Failed to download book');
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading book...</p>
        </div>
      </div>
    );
  }

  if (error || !book) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">📖</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">{error || 'Book not found'}</h3>
          <Link
            href="/books"
            className="text-indigo-600 hover:text-indigo-800 font-medium"
          >
            ← Back to Library
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Main Content */}
      <div className={`flex-1 transition-all duration-300 ${showSidebar ? 'mr-96' : ''}`}>
        {/* Header */}
        <header className="bg-white shadow-sm border-b sticky top-0 z-10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center h-16">
              <div className="flex items-center space-x-4">
                <Link href="/books" className="text-indigo-600 hover:text-indigo-800">
                  ← Back to Library
                </Link>
                <div className="border-l border-gray-300 pl-4">
                  <h1 className="text-lg font-semibold text-gray-900 truncate max-w-md">
                    {book.title}
                  </h1>
                  <p className="text-sm text-gray-600">by {book.author}</p>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <button
                  onClick={() => {
                    setShowSidebar(!showSidebar);
                    setSidebarTab('info');
                  }}
                  className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded"
                  title="Book Info"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                  </svg>
                </button>
                <button
                  onClick={() => {
                    setShowSidebar(!showSidebar);
                    setSidebarTab('annotations');
                  }}
                  className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded"
                  title="Annotations"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
                  </svg>
                </button>
                <button
                  onClick={() => {
                    setShowSidebar(!showSidebar);
                    setSidebarTab('sage');
                  }}
                  className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded"
                  title="AI Sage"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M18 13V5a2 2 0 00-2-2H4a2 2 0 00-2 2v8a2 2 0 002 2h3l3 3 3-3h3a2 2 0 002-2zM5 7a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1zm1 3a1 1 0 100 2h3a1 1 0 100-2H6z" clipRule="evenodd" />
                  </svg>
                </button>
                <Link
                  href={`/reading/${bookId}`}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 text-sm font-medium"
                  title="Read Book"
                >
                  📖 Read
                </Link>
                <button
                  onClick={downloadBook}
                  className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded"
                  title="Download"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </header>

        {/* Book Content */}
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-white rounded-lg shadow-sm min-h-[600px]">
            {book.file_type === 'pdf' ? (
              <div className="p-8 text-center">
                <div className="text-6xl mb-4">📄</div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">PDF Viewer</h3>
                <p className="text-gray-600 mb-4">
                  PDF viewing functionality would be implemented here using libraries like PDF.js
                </p>
                <button
                  onClick={downloadBook}
                  className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
                >
                  Download to View
                </button>
              </div>
            ) : book.file_type === 'epub' ? (
              <div className="p-8 text-center">
                <div className="text-6xl mb-4">📖</div>
                <h3 className="text-lg font-medium text-gray-900 mb-2">EPUB Reader</h3>
                <p className="text-gray-600 mb-4">
                  EPUB reading interface would be implemented here using libraries like Epub.js
                </p>
                <button
                  onClick={downloadBook}
                  className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
                >
                  Download to View
                </button>
              </div>
            ) : (
              <div className="p-8">
                <div 
                  className="prose max-w-none text-gray-900 leading-relaxed"
                  onMouseUp={handleTextSelection}
                  style={{ fontFamily: 'Georgia, serif', fontSize: '18px', lineHeight: '1.8' }}
                >
                  <p className="mb-6 text-center italic text-gray-600">
                    [Text content would be loaded from the server and displayed here]
                  </p>
                  <p className="mb-4">
                    This is where the actual text content of the book would be displayed. 
                    For demonstration purposes, this shows how the reading interface would look.
                  </p>
                  <p className="mb-4">
                    Users can select text to create annotations, ask questions to the AI Sage, 
                    and navigate through the book content. The interface supports highlighting, 
                    note-taking, and intelligent assistance for understanding classical texts.
                  </p>
                  <p className="mb-4">
                    The actual implementation would parse different file formats and present 
                    them in a unified reading interface optimized for classical education.
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Sidebar */}
      {showSidebar && (
        <div className="fixed right-0 top-0 h-full w-96 bg-white shadow-lg border-l z-20 overflow-y-auto">
          {/* Sidebar Header */}
          <div className="flex items-center justify-between p-4 border-b">
            <div className="flex space-x-1">
              <button
                onClick={() => setSidebarTab('info')}
                className={`px-3 py-2 text-sm font-medium rounded ${
                  sidebarTab === 'info' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Info
              </button>
              <button
                onClick={() => setSidebarTab('annotations')}
                className={`px-3 py-2 text-sm font-medium rounded ${
                  sidebarTab === 'annotations' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Notes ({annotations.length})
              </button>
              <button
                onClick={() => setSidebarTab('sage')}
                className={`px-3 py-2 text-sm font-medium rounded ${
                  sidebarTab === 'sage' ? 'bg-indigo-100 text-indigo-700' : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                AI Sage
              </button>
            </div>
            <button
              onClick={() => setShowSidebar(false)}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Sidebar Content */}
          <div className="p-4">
            {sidebarTab === 'info' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Book Details</h3>
                  <div className="space-y-3">
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Title</dt>
                      <dd className="text-sm text-gray-900">{book.title}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Author</dt>
                      <dd className="text-sm text-gray-900">{book.author}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Language</dt>
                      <dd className="text-sm text-gray-900">{book.language.toUpperCase()}</dd>
                    </div>
                    {book.genre && (
                      <div>
                        <dt className="text-sm font-medium text-gray-600">Genre</dt>
                        <dd className="text-sm text-gray-900">{book.genre}</dd>
                      </div>
                    )}
                    {book.publisher && (
                      <div>
                        <dt className="text-sm font-medium text-gray-600">Publisher</dt>
                        <dd className="text-sm text-gray-900">{book.publisher}</dd>
                      </div>
                    )}
                    {book.isbn && (
                      <div>
                        <dt className="text-sm font-medium text-gray-600">ISBN</dt>
                        <dd className="text-sm text-gray-900">{book.isbn}</dd>
                      </div>
                    )}
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Format</dt>
                      <dd className="text-sm text-gray-900">{book.file_type.toUpperCase()}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Size</dt>
                      <dd className="text-sm text-gray-900">{formatFileSize(book.file_size)}</dd>
                    </div>
                    <div>
                      <dt className="text-sm font-medium text-gray-600">Added</dt>
                      <dd className="text-sm text-gray-900">{formatDate(book.created_at)}</dd>
                    </div>
                  </div>
                </div>

                {book.description && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-600 mb-2">Description</h4>
                    <p className="text-sm text-gray-900">{book.description}</p>
                  </div>
                )}

                {book.tags && book.tags.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-600 mb-2">Tags</h4>
                    <div className="flex flex-wrap gap-2">
                      {book.tags.map((tag) => (
                        <span
                          key={tag.id}
                          className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
                        >
                          {tag.name}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {sidebarTab === 'annotations' && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-900">Annotations</h3>

                {/* Annotation Form */}
                {showAnnotationForm && (
                  <div className="bg-gray-50 rounded-lg p-4 space-y-3">
                    <div>
                      <label className="text-sm font-medium text-gray-700">Selected Text:</label>
                      <p className="text-sm text-gray-900 bg-white p-2 rounded border italic">
                        "{selectedText}"
                      </p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Your Note:</label>
                      <textarea
                        value={annotationNote}
                        onChange={(e) => setAnnotationNote(e.target.value)}
                        className="w-full mt-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
                        placeholder="Add your thoughts or notes..."
                        rows={3}
                      />
                    </div>
                    <div className="flex space-x-2">
                      <button
                        onClick={saveAnnotation}
                        className="px-3 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700"
                      >
                        Save
                      </button>
                      <button
                        onClick={() => {
                          setShowAnnotationForm(false);
                          setSelectedText('');
                          setAnnotationNote('');
                        }}
                        className="px-3 py-2 border border-gray-300 text-gray-700 text-sm rounded-md hover:bg-gray-50"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                )}

                {/* Annotations List */}
                {annotations.length === 0 ? (
                  <div className="text-center py-8">
                    <div className="text-4xl mb-2">📝</div>
                    <p className="text-gray-600 text-sm">No annotations yet</p>
                    <p className="text-gray-500 text-xs mt-1">Select text to create your first annotation</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {annotations.map((annotation) => (
                      <div key={annotation.id} className="bg-gray-50 rounded-lg p-3">
                        <p className="text-sm text-gray-900 italic mb-2">"{annotation.content}"</p>
                        {annotation.note && (
                          <p className="text-sm text-gray-700">{annotation.note}</p>
                        )}
                        <p className="text-xs text-gray-500 mt-2">
                          {formatDate(annotation.created_at)}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {sidebarTab === 'sage' && (
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-gray-900">AI Sage</h3>
                <p className="text-sm text-gray-600">
                  Ask questions about the text, get explanations of difficult passages, 
                  or explore historical and cultural context.
                </p>

                {/* Query Form */}
                <div className="space-y-3">
                  {selectedText && (
                    <div>
                      <label className="text-sm font-medium text-gray-700">Selected Context:</label>
                      <p className="text-sm text-gray-900 bg-gray-50 p-2 rounded border italic">
                        "{selectedText}"
                      </p>
                    </div>
                  )}
                  <div>
                    <label className="text-sm font-medium text-gray-700">Your Question:</label>
                    <textarea
                      value={sageQuery}
                      onChange={(e) => setSageQuery(e.target.value)}
                      className="w-full mt-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
                      placeholder="What would you like to know about this text?"
                      rows={3}
                    />
                  </div>
                  <button
                    onClick={askSage}
                    disabled={sageLoading || !sageQuery.trim()}
                    className="w-full px-3 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 disabled:opacity-50 flex items-center justify-center"
                  >
                    {sageLoading && (
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    )}
                    {sageLoading ? 'Thinking...' : 'Ask the Sage'}
                  </button>
                </div>

                {/* Response */}
                {sageResponse && (
                  <div className="bg-indigo-50 rounded-lg p-4">
                    <h4 className="text-sm font-medium text-indigo-900 mb-2">AI Sage Response:</h4>
                    <div className="text-sm text-indigo-800 whitespace-pre-wrap">
                      {sageResponse}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}