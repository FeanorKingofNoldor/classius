'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface Annotation {
  id: string;
  book_id: string;
  user_id: string;
  type: 'highlight' | 'note' | 'bookmark';
  page_number?: number;
  start_position: number;
  end_position: number;
  selected_text?: string;
  content?: string;
  color?: string;
  tags?: string[];
  is_private: boolean;
  created_at: string;
  updated_at: string;
  book?: {
    id: string;
    title: string;
    author: string;
    language: string;
  };
}

interface AnnotationStats {
  total_annotations: number;
  highlights: number;
  notes: number;
  bookmarks: number;
  books_with_annotations: number;
  favorite_books: { book_title: string; book_author: string; count: number }[];
}

export default function AnnotationsPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();

  const [annotations, setAnnotations] = useState<Annotation[]>([]);
  const [stats, setStats] = useState<AnnotationStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedAnnotations, setSelectedAnnotations] = useState<string[]>([]);
  const [activeTab, setActiveTab] = useState<'all' | 'highlights' | 'notes' | 'bookmarks'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedBook, setSelectedBook] = useState('');
  const [sortBy, setSortBy] = useState<'date' | 'book' | 'type'>('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [editingAnnotation, setEditingAnnotation] = useState<string | null>(null);
  const [editContent, setEditContent] = useState('');

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [totalAnnotations, setTotalAnnotations] = useState(0);
  const perPage = 20;

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    fetchAnnotations();
    fetchAnnotationStats();
  }, [isAuthenticated, router, currentPage, activeTab, searchQuery, selectedBook, sortBy, sortOrder]);

  const fetchAnnotations = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) return;

      const params = new URLSearchParams({
        page: currentPage.toString(),
        per_page: perPage.toString(),
        sort_by: sortBy,
        sort_order: sortOrder,
      });

      if (searchQuery) params.append('q', searchQuery);
      if (selectedBook) params.append('book_id', selectedBook);
      if (activeTab !== 'all') params.append('type', activeTab.slice(0, -1)); // Remove 's' from plural

      const response = await fetch(`http://localhost:8080/api/annotations?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setAnnotations(result.data.annotations || []);
        setTotalAnnotations(result.data.total || 0);
        setTotalPages(Math.ceil((result.data.total || 0) / perPage));
      }
    } catch (err) {
      console.error('Error fetching annotations:', err);
      toast.error('Failed to load annotations');
    } finally {
      setLoading(false);
    }
  };

  const fetchAnnotationStats = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch('http://localhost:8080/api/annotations/stats', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setStats(result.data);
      }
    } catch (err) {
      console.error('Error fetching annotation stats:', err);
    }
  };

  const updateAnnotation = async (annotationId: string, content: string) => {
    try {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch(`http://localhost:8080/api/annotations/${annotationId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content }),
      });

      if (!response.ok) {
        throw new Error('Failed to update annotation');
      }

      toast.success('Annotation updated successfully');
      setEditingAnnotation(null);
      fetchAnnotations();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update annotation');
    }
  };

  const deleteAnnotations = async (annotationIds: string[]) => {
    if (!confirm(`Are you sure you want to delete ${annotationIds.length} annotation(s)?`)) {
      return;
    }

    try {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const promises = annotationIds.map(id =>
        fetch(`http://localhost:8080/api/annotations/${id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        })
      );

      await Promise.all(promises);
      toast.success('Annotations deleted successfully');
      setSelectedAnnotations([]);
      fetchAnnotations();
      fetchAnnotationStats();
    } catch (err) {
      toast.error('Failed to delete annotations');
    }
  };

  const exportAnnotations = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch('http://localhost:8080/api/annotations/export', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to export annotations');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `classius-annotations-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success('Annotations exported successfully');
    } catch (err) {
      toast.error('Failed to export annotations');
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getAnnotationIcon = (type: string) => {
    switch (type) {
      case 'highlight': return '✏️';
      case 'note': return '📝';
      case 'bookmark': return '🔖';
      default: return '📄';
    }
  };

  const getAnnotationColor = (color?: string) => {
    const colors: { [key: string]: string } = {
      yellow: 'bg-yellow-100 border-yellow-300',
      blue: 'bg-blue-100 border-blue-300',
      green: 'bg-green-100 border-green-300',
      red: 'bg-red-100 border-red-300',
      purple: 'bg-purple-100 border-purple-300',
    };
    return colors[color || 'yellow'] || colors.yellow;
  };

  const filteredAnnotations = annotations.filter(annotation => {
    if (activeTab !== 'all' && !annotation.type.startsWith(activeTab.slice(0, -1))) {
      return false;
    }
    if (searchQuery && !annotation.selected_text?.toLowerCase().includes(searchQuery.toLowerCase()) &&
        !annotation.content?.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    return true;
  });

  const tabs = [
    { id: 'all', name: 'All', count: stats?.total_annotations || 0 },
    { id: 'highlights', name: 'Highlights', count: stats?.highlights || 0 },
    { id: 'notes', name: 'Notes', count: stats?.notes || 0 },
    { id: 'bookmarks', name: 'Bookmarks', count: stats?.bookmarks || 0 },
  ] as const;

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <Link href="/dashboard" className="text-indigo-600 hover:text-indigo-800">
                ← Back to Dashboard
              </Link>
              <div className="border-l border-gray-300 pl-4">
                <h1 className="text-xl font-semibold text-gray-900 flex items-center">
                  <span className="text-2xl mr-2">📝</span>
                  My Annotations
                </h1>
                <p className="text-sm text-gray-600">
                  {totalAnnotations} {totalAnnotations === 1 ? 'annotation' : 'annotations'} across {stats?.books_with_annotations || 0} books
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={exportAnnotations}
                className="px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md"
              >
                Export All
              </button>
              {selectedAnnotations.length > 0 && (
                <button
                  onClick={() => deleteAnnotations(selectedAnnotations)}
                  className="px-3 py-2 text-sm text-red-600 hover:text-red-800 hover:bg-red-50 rounded-md"
                >
                  Delete ({selectedAnnotations.length})
                </button>
              )}
              <span className="text-gray-700">Welcome, {user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Statistics Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <div className="text-3xl mr-4">📝</div>
                <div>
                  <div className="text-2xl font-bold text-gray-900">{stats.total_annotations}</div>
                  <div className="text-sm text-gray-600">Total Annotations</div>
                </div>
              </div>
            </div>
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <div className="text-3xl mr-4">📚</div>
                <div>
                  <div className="text-2xl font-bold text-gray-900">{stats.books_with_annotations}</div>
                  <div className="text-sm text-gray-600">Books Annotated</div>
                </div>
              </div>
            </div>
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <div className="text-3xl mr-4">✏️</div>
                <div>
                  <div className="text-2xl font-bold text-gray-900">{stats.highlights}</div>
                  <div className="text-sm text-gray-600">Highlights</div>
                </div>
              </div>
            </div>
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <div className="text-3xl mr-4">📄</div>
                <div>
                  <div className="text-2xl font-bold text-gray-900">{stats.notes}</div>
                  <div className="text-sm text-gray-600">Notes</div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Filters and Search */}
        <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
          {/* Tabs */}
          <div className="border-b border-gray-200 mb-4">
            <nav className="-mb-px flex space-x-8">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-2 px-1 border-b-2 font-medium text-sm ${
                    activeTab === tab.id
                      ? 'border-indigo-500 text-indigo-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  {tab.name} ({tab.count})
                </button>
              ))}
            </nav>
          </div>

          {/* Search and Sort */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="md:col-span-2">
              <input
                type="text"
                placeholder="Search annotations..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
            <div>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as any)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="date">Sort by Date</option>
                <option value="book">Sort by Book</option>
                <option value="type">Sort by Type</option>
              </select>
            </div>
            <div>
              <select
                value={sortOrder}
                onChange={(e) => setSortOrder(e.target.value as any)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="desc">Newest First</option>
                <option value="asc">Oldest First</option>
              </select>
            </div>
          </div>
        </div>

        {/* Annotations List */}
        <div className="space-y-4">
          {loading ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
              <p className="mt-2 text-gray-600">Loading annotations...</p>
            </div>
          ) : filteredAnnotations.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-lg shadow-sm border">
              <div className="text-6xl mb-4">📝</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">No annotations found</h3>
              <p className="text-gray-600">
                {searchQuery ? "Try adjusting your search criteria" : "Start reading and highlighting text to create annotations"}
              </p>
            </div>
          ) : (
            filteredAnnotations.map((annotation) => (
              <div
                key={annotation.id}
                className={`bg-white rounded-lg shadow-sm border p-6 ${getAnnotationColor(annotation.color)}`}
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-start space-x-3">
                    <input
                      type="checkbox"
                      checked={selectedAnnotations.includes(annotation.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedAnnotations([...selectedAnnotations, annotation.id]);
                        } else {
                          setSelectedAnnotations(selectedAnnotations.filter(id => id !== annotation.id));
                        }
                      }}
                      className="mt-1 h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                    />
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <span className="text-lg">{getAnnotationIcon(annotation.type)}</span>
                        <span className="text-sm font-medium text-gray-900 capitalize">{annotation.type}</span>
                        {annotation.book && (
                          <Link
                            href={`/books/${annotation.book.id}`}
                            className="text-sm text-indigo-600 hover:text-indigo-800"
                          >
                            {annotation.book.title} by {annotation.book.author}
                          </Link>
                        )}
                        <span className="text-xs text-gray-500">
                          {formatDate(annotation.created_at)}
                        </span>
                      </div>

                      {/* Selected Text */}
                      {annotation.selected_text && (
                        <div className="mb-3 p-3 bg-gray-50 rounded-md border-l-4 border-indigo-400">
                          <p className="text-sm text-gray-700 italic">
                            "{annotation.selected_text}"
                          </p>
                          {annotation.page_number && (
                            <p className="text-xs text-gray-500 mt-1">Page {annotation.page_number}</p>
                          )}
                        </div>
                      )}

                      {/* Note Content */}
                      {annotation.content && (
                        <div className="mb-3">
                          {editingAnnotation === annotation.id ? (
                            <div className="space-y-2">
                              <textarea
                                value={editContent}
                                onChange={(e) => setEditContent(e.target.value)}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
                                rows={3}
                              />
                              <div className="flex space-x-2">
                                <button
                                  onClick={() => updateAnnotation(annotation.id, editContent)}
                                  className="px-3 py-1 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700"
                                >
                                  Save
                                </button>
                                <button
                                  onClick={() => {
                                    setEditingAnnotation(null);
                                    setEditContent('');
                                  }}
                                  className="px-3 py-1 border border-gray-300 text-gray-700 text-sm rounded-md hover:bg-gray-50"
                                >
                                  Cancel
                                </button>
                              </div>
                            </div>
                          ) : (
                            <p className="text-sm text-gray-900">{annotation.content}</p>
                          )}
                        </div>
                      )}

                      {/* Tags */}
                      {annotation.tags && annotation.tags.length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-3">
                          {annotation.tags.map((tag, index) => (
                            <span
                              key={index}
                              className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Action Menu */}
                  <div className="flex items-center space-x-2">
                    {annotation.content && editingAnnotation !== annotation.id && (
                      <button
                        onClick={() => {
                          setEditingAnnotation(annotation.id);
                          setEditContent(annotation.content || '');
                        }}
                        className="text-gray-400 hover:text-gray-600"
                        title="Edit annotation"
                      >
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                        </svg>
                      </button>
                    )}
                    <button
                      onClick={() => deleteAnnotations([annotation.id])}
                      className="text-gray-400 hover:text-red-600"
                      title="Delete annotation"
                    >
                      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" clipRule="evenodd" />
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414L7.586 12l-1.293 1.293a1 1 0 101.414 1.414L9 13.414l2.293 2.293a1 1 0 001.414-1.414L11.414 12l1.293-1.293z" clipRule="evenodd" />
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between mt-8">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                disabled={currentPage === totalPages}
                className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
              >
                Next
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Showing <span className="font-medium">{(currentPage - 1) * perPage + 1}</span> to{' '}
                  <span className="font-medium">
                    {Math.min(currentPage * perPage, totalAnnotations)}
                  </span> of <span className="font-medium">{totalAnnotations}</span> annotations
                </p>
              </div>
              <div>
                <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    <span className="sr-only">Previous</span>
                    <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  </button>
                  
                  {[...Array(Math.min(5, totalPages))].map((_, i) => {
                    let pageNum;
                    if (totalPages <= 5) {
                      pageNum = i + 1;
                    } else if (currentPage <= 3) {
                      pageNum = i + 1;
                    } else if (currentPage >= totalPages - 2) {
                      pageNum = totalPages - 4 + i;
                    } else {
                      pageNum = currentPage - 2 + i;
                    }
                    
                    return (
                      <button
                        key={pageNum}
                        onClick={() => setCurrentPage(pageNum)}
                        className={`relative inline-flex items-center px-4 py-2 border text-sm font-medium ${
                          currentPage === pageNum
                            ? 'z-10 bg-indigo-50 border-indigo-500 text-indigo-600'
                            : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}
                  
                  <button
                    onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                    disabled={currentPage === totalPages}
                    className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                  >
                    <span className="sr-only">Next</span>
                    <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                    </svg>
                  </button>
                </nav>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}