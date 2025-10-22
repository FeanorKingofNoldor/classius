'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface SearchResult {
  id: string;
  type: 'book' | 'annotation' | 'highlight' | 'note';
  title: string;
  content: string;
  book_title?: string;
  book_author?: string;
  book_id?: string;
  author?: string;
  language?: string;
  snippet: string;
  score: number;
  created_at: string;
}

interface SearchStats {
  total_results: number;
  books: number;
  annotations: number;
  highlights: number;
  notes: number;
  search_time: string;
}

export default function SearchPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get('q') || '';

  const [query, setQuery] = useState(initialQuery);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [stats, setStats] = useState<SearchStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [activeFilter, setActiveFilter] = useState<'all' | 'books' | 'annotations' | 'highlights' | 'notes'>('all');
  const [sortBy, setSortBy] = useState<'relevance' | 'date' | 'title'>('relevance');
  
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const perPage = 20;

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    if (initialQuery) {
      performSearch(initialQuery);
    }
  }, [isAuthenticated, router, initialQuery]);

  useEffect(() => {
    if (query && query !== initialQuery) {
      const delayedSearch = setTimeout(() => {
        performSearch(query);
      }, 500); // Debounce search
      
      return () => clearTimeout(delayedSearch);
    }
  }, [query, activeFilter, sortBy, currentPage]);

  const performSearch = async (searchQuery: string) => {
    if (!searchQuery.trim()) {
      setResults([]);
      setStats(null);
      return;
    }

    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) return;

      const params = new URLSearchParams({
        q: searchQuery.trim(),
        type: activeFilter === 'all' ? '' : activeFilter,
        sort_by: sortBy,
        page: currentPage.toString(),
        per_page: perPage.toString(),
      });

      // Remove empty params
      Object.keys(params).forEach(key => {
        if (!params.get(key)) {
          params.delete(key);
        }
      });

      const response = await fetch(`http://localhost:8080/api/search?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setResults(result.data.results || []);
        setStats(result.data.stats);
        setTotalPages(Math.ceil(result.data.stats.total_results / perPage));
        
        // Update URL without causing a page refresh
        const newUrl = `/search?q=${encodeURIComponent(searchQuery)}`;
        window.history.replaceState({}, '', newUrl);
      } else {
        throw new Error('Search failed');
      }
    } catch (err) {
      console.error('Search error:', err);
      toast.error('Search failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    performSearch(query);
  };

  const highlightText = (text: string, searchTerm: string) => {
    if (!searchTerm) return text;
    
    const regex = new RegExp(`(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);
    
    return parts.map((part, index) =>
      regex.test(part) ? (
        <span key={index} className="bg-yellow-200 font-medium">
          {part}
        </span>
      ) : (
        part
      )
    );
  };

  const getResultIcon = (type: string) => {
    switch (type) {
      case 'book': return '📖';
      case 'annotation': return '📝';
      case 'highlight': return '✏️';
      case 'note': return '📄';
      default: return '📄';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (!isAuthenticated) {
    return null;
  }

  const filters = [
    { id: 'all', name: 'All Results', count: stats?.total_results || 0 },
    { id: 'books', name: 'Books', count: stats?.books || 0 },
    { id: 'annotations', name: 'Annotations', count: stats?.annotations || 0 },
    { id: 'highlights', name: 'Highlights', count: stats?.highlights || 0 },
    { id: 'notes', name: 'Notes', count: stats?.notes || 0 },
  ] as const;

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
                  <span className="text-2xl mr-2">🔍</span>
                  Search
                </h1>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Welcome, {user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Search Form */}
        <div className="mb-8">
          <form onSubmit={handleSearch} className="relative">
            <div className="flex">
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search across all your books, annotations, and notes..."
                className="flex-1 px-4 py-3 text-lg border border-gray-300 rounded-l-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                autoFocus
              />
              <button
                type="submit"
                className="px-8 py-3 bg-indigo-600 text-white rounded-r-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 flex items-center"
              >
                <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
                </svg>
                Search
              </button>
            </div>
          </form>
        </div>

        {/* Search Stats */}
        {stats && (
          <div className="mb-6 flex items-center justify-between text-sm text-gray-600">
            <div>
              {stats.total_results.toLocaleString()} results found in {stats.search_time}
            </div>
            <div className="flex items-center space-x-4">
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as any)}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="relevance">Sort by Relevance</option>
                <option value="date">Sort by Date</option>
                <option value="title">Sort by Title</option>
              </select>
            </div>
          </div>
        )}

        {/* Filter Tabs */}
        {stats && (
          <div className="mb-6 border-b border-gray-200">
            <nav className="-mb-px flex space-x-8">
              {filters.map((filter) => (
                <button
                  key={filter.id}
                  onClick={() => {
                    setActiveFilter(filter.id);
                    setCurrentPage(1);
                  }}
                  className={`py-2 px-1 border-b-2 font-medium text-sm ${
                    activeFilter === filter.id
                      ? 'border-indigo-500 text-indigo-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  {filter.name} ({filter.count.toLocaleString()})
                </button>
              ))}
            </nav>
          </div>
        )}

        {/* Search Results */}
        <div className="space-y-4">
          {loading && (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
              <p className="mt-2 text-gray-600">Searching...</p>
            </div>
          )}

          {!loading && !query && (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">🔍</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">Search your classical library</h3>
              <p className="text-gray-600">
                Find books, annotations, highlights, and notes across your entire collection
              </p>
            </div>
          )}

          {!loading && query && results.length === 0 && (
            <div className="text-center py-12 bg-white rounded-lg border">
              <div className="text-6xl mb-4">🔍</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">No results found</h3>
              <p className="text-gray-600 mb-4">
                Try different keywords or check your spelling
              </p>
              <div className="text-sm text-gray-500">
                <p>Search tips:</p>
                <ul className="mt-2 space-y-1">
                  <li>• Try broader or different keywords</li>
                  <li>• Check for typos in your search</li>
                  <li>• Use quotes for exact phrases: "virtue ethics"</li>
                  <li>• Search for author names or book titles</li>
                </ul>
              </div>
            </div>
          )}

          {!loading && results.length > 0 && (
            <>
              {results.map((result) => (
                <div key={`${result.type}-${result.id}`} className="bg-white rounded-lg border p-6 hover:shadow-md transition-shadow">
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center space-x-3">
                      <span className="text-2xl">{getResultIcon(result.type)}</span>
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium text-gray-500 capitalize">{result.type}</span>
                          <span className="text-xs text-gray-400">•</span>
                          <span className="text-xs text-gray-500">{formatDate(result.created_at)}</span>
                          {result.score && (
                            <>
                              <span className="text-xs text-gray-400">•</span>
                              <span className="text-xs text-gray-500">Score: {Math.round(result.score * 100)}%</span>
                            </>
                          )}
                        </div>
                        <h3 className="text-lg font-medium text-gray-900 mt-1">
                          {result.type === 'book' ? (
                            <Link
                              href={`/books/${result.id}`}
                              className="text-indigo-600 hover:text-indigo-800"
                            >
                              {highlightText(result.title, query)}
                            </Link>
                          ) : (
                            highlightText(result.title, query)
                          )}
                        </h3>
                      </div>
                    </div>
                  </div>

                  {/* Book Context */}
                  {result.book_title && result.book_author && (
                    <div className="mb-3">
                      <Link
                        href={`/books/${result.book_id}`}
                        className="text-sm text-indigo-600 hover:text-indigo-800"
                      >
                        📖 {result.book_title} by {result.book_author}
                      </Link>
                    </div>
                  )}

                  {/* Author for books */}
                  {result.type === 'book' && result.author && (
                    <p className="text-sm text-gray-600 mb-3">by {result.author}</p>
                  )}

                  {/* Content Snippet */}
                  <div className="text-gray-700">
                    <p className="line-clamp-3">
                      {highlightText(result.snippet, query)}
                    </p>
                  </div>

                  {/* Language for books */}
                  {result.type === 'book' && result.language && (
                    <div className="mt-3">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {result.language.toUpperCase()}
                      </span>
                    </div>
                  )}
                </div>
              ))}

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
                          {Math.min(currentPage * perPage, stats?.total_results || 0)}
                        </span> of <span className="font-medium">{stats?.total_results || 0}</span> results
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
            </>
          )}
        </div>
      </div>
    </div>
  );
}