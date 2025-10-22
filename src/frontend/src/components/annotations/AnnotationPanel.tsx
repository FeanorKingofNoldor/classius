'use client';

import React, { useState, useMemo } from 'react';
import { useAnnotations, Annotation, AnnotationFilter } from '@/hooks/useAnnotations';

interface AnnotationPanelProps {
  bookId: string;
  onNavigateToAnnotation?: (annotation: Annotation) => void;
  theme?: 'light' | 'dark' | 'sepia';
}

const ANNOTATION_COLORS = [
  { name: 'Yellow', value: '#fef3c7' },
  { name: 'Green', value: '#d1fae5' },
  { name: 'Blue', value: '#dbeafe' },
  { name: 'Purple', value: '#e9d5ff' },
  { name: 'Pink', value: '#fce7f3' },
  { name: 'Orange', value: '#fed7aa' },
  { name: 'Red', value: '#fee2e2' },
  { name: 'Gray', value: '#f3f4f6' },
];

export const AnnotationPanel: React.FC<AnnotationPanelProps> = ({
  bookId,
  onNavigateToAnnotation,
  theme = 'light',
}) => {
  const {
    filteredAnnotations,
    loading,
    filter,
    setFilter,
    clearFilter,
    updateAnnotation,
    deleteAnnotation,
    bulkDelete,
    bulkUpdateTags,
    bulkUpdateColor,
    exportAnnotations,
    searchAnnotations,
    getStats,
  } = useAnnotations(bookId);

  const [searchQuery, setSearchQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [selectedAnnotations, setSelectedAnnotations] = useState<string[]>([]);
  const [showBulkActions, setShowBulkActions] = useState(false);
  const [bulkTagInput, setBulkTagInput] = useState('');
  const [bulkColor, setBulkColor] = useState(ANNOTATION_COLORS[0].value);
  const [sortBy, setSortBy] = useState<'date' | 'type' | 'page'>('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  const stats = getStats();

  // Apply search and sorting
  const displayAnnotations = useMemo(() => {
    let annotations = searchQuery 
      ? searchAnnotations(searchQuery)
      : filteredAnnotations;

    // Sort annotations
    annotations.sort((a, b) => {
      let compareValue = 0;
      
      switch (sortBy) {
        case 'date':
          compareValue = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
        case 'type':
          compareValue = a.type.localeCompare(b.type);
          break;
        case 'page':
          compareValue = (a.page_number || 0) - (b.page_number || 0);
          break;
      }

      return sortOrder === 'asc' ? compareValue : -compareValue;
    });

    return annotations;
  }, [filteredAnnotations, searchQuery, searchAnnotations, sortBy, sortOrder]);

  const handleSelectAll = () => {
    if (selectedAnnotations.length === displayAnnotations.length) {
      setSelectedAnnotations([]);
    } else {
      setSelectedAnnotations(displayAnnotations.map(a => a.id));
    }
  };

  const handleSelectAnnotation = (id: string) => {
    setSelectedAnnotations(prev => 
      prev.includes(id) 
        ? prev.filter(selectedId => selectedId !== id)
        : [...prev, id]
    );
  };

  const handleBulkDelete = async () => {
    if (selectedAnnotations.length > 0 && 
        confirm(`Delete ${selectedAnnotations.length} selected annotations?`)) {
      await bulkDelete(selectedAnnotations);
      setSelectedAnnotations([]);
      setShowBulkActions(false);
    }
  };

  const handleBulkUpdateTags = async () => {
    if (selectedAnnotations.length > 0 && bulkTagInput.trim()) {
      const tags = bulkTagInput.split(',').map(tag => tag.trim()).filter(tag => tag);
      await bulkUpdateTags(selectedAnnotations, tags);
      setBulkTagInput('');
      setShowBulkActions(false);
    }
  };

  const handleBulkUpdateColor = async () => {
    if (selectedAnnotations.length > 0) {
      await bulkUpdateColor(selectedAnnotations, bulkColor);
      setShowBulkActions(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    });
  };

  const getThemeClasses = () => {
    const themes = {
      light: 'bg-white border-gray-300 text-gray-900',
      dark: 'bg-gray-800 border-gray-600 text-gray-100',
      sepia: 'bg-yellow-50 border-yellow-300 text-yellow-900',
    };
    return themes[theme];
  };

  if (loading) {
    return (
      <div className={`w-full h-96 flex items-center justify-center rounded-lg border ${getThemeClasses()}`}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto mb-2"></div>
          <p className="text-sm opacity-60">Loading annotations...</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`w-full rounded-lg border ${getThemeClasses()}`}>
      {/* Header */}
      <div className="p-4 border-b border-opacity-20 border-gray-400">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Annotations</h3>
          <div className="flex items-center space-x-2">
            <span className="text-sm opacity-60">{stats.total} total</span>
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`px-3 py-1 text-sm rounded border transition-colors ${
                showFilters ? 'bg-blue-100 text-blue-700 border-blue-300' : 'hover:bg-gray-100'
              }`}
            >
              Filters
            </button>
          </div>
        </div>

        {/* Search */}
        <div className="flex space-x-2 mb-3">
          <input
            type="text"
            placeholder="Search annotations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className={`flex-1 px-3 py-2 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              theme === 'dark' 
                ? 'bg-gray-700 border-gray-600 text-gray-100' 
                : 'bg-white border-gray-300 text-gray-900'
            }`}
          />
          <button
            onClick={() => setSearchQuery('')}
            disabled={!searchQuery}
            className="px-3 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
          >
            Clear
          </button>
        </div>

        {/* Quick Stats */}
        <div className="flex space-x-4 text-xs opacity-70">
          <span>🖍️ {stats.highlights} highlights</span>
          <span>📝 {stats.notes} notes</span>
          <span>🔖 {stats.bookmarks} bookmarks</span>
        </div>
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="p-4 border-b border-opacity-20 border-gray-400 bg-opacity-50 bg-gray-50 dark:bg-gray-900">
          <div className="space-y-4">
            {/* Type and Color filters */}
            <div className="flex flex-wrap gap-4">
              <div>
                <label className="block text-xs font-medium mb-1 opacity-70">Type</label>
                <select
                  value={filter.type || 'all'}
                  onChange={(e) => setFilter({ type: e.target.value as any })}
                  className={`px-2 py-1 text-sm border rounded ${
                    theme === 'dark' 
                      ? 'bg-gray-700 border-gray-600 text-gray-100' 
                      : 'bg-white border-gray-300 text-gray-900'
                  }`}
                >
                  <option value="all">All Types</option>
                  <option value="highlight">Highlights</option>
                  <option value="note">Notes</option>
                  <option value="bookmark">Bookmarks</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1 opacity-70">Color</label>
                <select
                  value={filter.color || ''}
                  onChange={(e) => setFilter({ color: e.target.value || undefined })}
                  className={`px-2 py-1 text-sm border rounded ${
                    theme === 'dark' 
                      ? 'bg-gray-700 border-gray-600 text-gray-100' 
                      : 'bg-white border-gray-300 text-gray-900'
                  }`}
                >
                  <option value="">All Colors</option>
                  {ANNOTATION_COLORS.map((color) => (
                    <option key={color.value} value={color.value}>
                      {color.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1 opacity-70">Privacy</label>
                <select
                  value={filter.isPrivate !== undefined ? filter.isPrivate.toString() : ''}
                  onChange={(e) => setFilter({ 
                    isPrivate: e.target.value === '' ? undefined : e.target.value === 'true' 
                  })}
                  className={`px-2 py-1 text-sm border rounded ${
                    theme === 'dark' 
                      ? 'bg-gray-700 border-gray-600 text-gray-100' 
                      : 'bg-white border-gray-300 text-gray-900'
                  }`}
                >
                  <option value="">All</option>
                  <option value="true">Private</option>
                  <option value="false">Public</option>
                </select>
              </div>
            </div>

            {/* Date filters */}
            <div className="flex space-x-4">
              <div>
                <label className="block text-xs font-medium mb-1 opacity-70">From Date</label>
                <input
                  type="date"
                  value={filter.dateFrom || ''}
                  onChange={(e) => setFilter({ dateFrom: e.target.value || undefined })}
                  className={`px-2 py-1 text-sm border rounded ${
                    theme === 'dark' 
                      ? 'bg-gray-700 border-gray-600 text-gray-100' 
                      : 'bg-white border-gray-300 text-gray-900'
                  }`}
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1 opacity-70">To Date</label>
                <input
                  type="date"
                  value={filter.dateTo || ''}
                  onChange={(e) => setFilter({ dateTo: e.target.value || undefined })}
                  className={`px-2 py-1 text-sm border rounded ${
                    theme === 'dark' 
                      ? 'bg-gray-700 border-gray-600 text-gray-100' 
                      : 'bg-white border-gray-300 text-gray-900'
                  }`}
                />
              </div>
            </div>

            <div className="flex justify-between">
              <button
                onClick={clearFilter}
                className="px-3 py-1 text-sm border border-gray-300 rounded hover:bg-gray-50"
              >
                Clear Filters
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Sorting and Bulk Actions */}
      <div className="p-3 border-b border-opacity-20 border-gray-400 flex items-center justify-between bg-opacity-30 bg-gray-100 dark:bg-gray-900">
        <div className="flex items-center space-x-3">
          {/* Selection */}
          <label className="flex items-center space-x-2 text-sm">
            <input
              type="checkbox"
              checked={selectedAnnotations.length === displayAnnotations.length && displayAnnotations.length > 0}
              onChange={handleSelectAll}
              className="rounded"
            />
            <span>Select all</span>
          </label>

          {selectedAnnotations.length > 0 && (
            <>
              <span className="text-sm opacity-60">
                {selectedAnnotations.length} selected
              </span>
              <button
                onClick={() => setShowBulkActions(!showBulkActions)}
                className="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
              >
                Bulk Actions
              </button>
            </>
          )}
        </div>

        <div className="flex items-center space-x-3">
          {/* Export */}
          <div className="flex items-center space-x-1">
            <span className="text-xs opacity-60">Export:</span>
            <button
              onClick={() => exportAnnotations('json')}
              className="text-xs px-2 py-1 border rounded hover:bg-gray-50"
            >
              JSON
            </button>
            <button
              onClick={() => exportAnnotations('csv')}
              className="text-xs px-2 py-1 border rounded hover:bg-gray-50"
            >
              CSV
            </button>
          </div>

          {/* Sorting */}
          <div className="flex items-center space-x-1 text-xs">
            <span className="opacity-60">Sort:</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="text-xs border rounded px-1 py-0.5"
            >
              <option value="date">Date</option>
              <option value="type">Type</option>
              <option value="page">Page</option>
            </select>
            <button
              onClick={() => setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')}
              className="text-xs p-1 hover:bg-gray-100 rounded"
              title={`Sort ${sortOrder === 'asc' ? 'descending' : 'ascending'}`}
            >
              {sortOrder === 'asc' ? '↑' : '↓'}
            </button>
          </div>
        </div>
      </div>

      {/* Bulk Actions Panel */}
      {showBulkActions && selectedAnnotations.length > 0 && (
        <div className="p-4 border-b border-opacity-20 border-gray-400 bg-blue-50 dark:bg-blue-900">
          <div className="space-y-3">
            <h4 className="text-sm font-medium">
              Bulk Actions ({selectedAnnotations.length} selected)
            </h4>
            
            <div className="flex flex-wrap gap-3">
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  placeholder="Tags (comma-separated)"
                  value={bulkTagInput}
                  onChange={(e) => setBulkTagInput(e.target.value)}
                  className="px-2 py-1 text-sm border rounded w-40"
                />
                <button
                  onClick={handleBulkUpdateTags}
                  disabled={!bulkTagInput.trim()}
                  className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                >
                  Update Tags
                </button>
              </div>

              <div className="flex items-center space-x-2">
                <select
                  value={bulkColor}
                  onChange={(e) => setBulkColor(e.target.value)}
                  className="px-2 py-1 text-sm border rounded"
                >
                  {ANNOTATION_COLORS.map((color) => (
                    <option key={color.value} value={color.value}>
                      {color.name}
                    </option>
                  ))}
                </select>
                <button
                  onClick={handleBulkUpdateColor}
                  className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                >
                  Update Color
                </button>
              </div>

              <button
                onClick={handleBulkDelete}
                className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
              >
                Delete Selected
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Annotations List */}
      <div className="max-h-96 overflow-y-auto">
        {displayAnnotations.length === 0 ? (
          <div className="p-8 text-center">
            <div className="text-4xl mb-2">📝</div>
            <p className="text-sm opacity-60 mb-2">
              {searchQuery || Object.keys(filter).length > 1 ? 'No matching annotations' : 'No annotations yet'}
            </p>
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Clear search
              </button>
            )}
          </div>
        ) : (
          <div className="divide-y divide-gray-200 divide-opacity-20">
            {displayAnnotations.map((annotation) => (
              <div key={annotation.id} className="p-4 hover:bg-gray-50 hover:bg-opacity-50">
                <div className="flex items-start space-x-3">
                  {/* Selection checkbox */}
                  <input
                    type="checkbox"
                    checked={selectedAnnotations.includes(annotation.id)}
                    onChange={() => handleSelectAnnotation(annotation.id)}
                    className="mt-1 rounded"
                  />

                  {/* Annotation content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2 mb-2">
                      <span className="text-sm">
                        {annotation.type === 'highlight' && '🖍️'}
                        {annotation.type === 'note' && '📝'}
                        {annotation.type === 'bookmark' && '🔖'}
                      </span>
                      <span className="text-xs font-medium capitalize opacity-70">
                        {annotation.type}
                      </span>
                      {annotation.page_number && (
                        <span className="text-xs opacity-60">
                          Page {annotation.page_number}
                        </span>
                      )}
                      <span className="text-xs opacity-60">
                        {formatDate(annotation.created_at)}
                      </span>
                      {annotation.is_private && (
                        <span className="text-xs bg-gray-200 px-1 rounded">Private</span>
                      )}
                    </div>

                    {annotation.selected_text && (
                      <p className="text-sm italic mb-2 p-2 rounded" 
                         style={{ 
                           backgroundColor: annotation.color || '#f3f4f6',
                           opacity: 0.8 
                         }}>
                        "{annotation.selected_text}"
                      </p>
                    )}

                    {annotation.content && (
                      <p className="text-sm mb-2 whitespace-pre-wrap">
                        {annotation.content}
                      </p>
                    )}

                    {annotation.tags && annotation.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1 mb-2">
                        {annotation.tags.map((tag) => (
                          <span
                            key={tag}
                            className="px-1.5 py-0.5 bg-blue-100 text-blue-700 text-xs rounded-full"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        {onNavigateToAnnotation && (
                          <button
                            onClick={() => onNavigateToAnnotation(annotation)}
                            className="text-xs text-blue-600 hover:text-blue-800"
                          >
                            Go to location
                          </button>
                        )}
                      </div>
                      
                      <div className="flex items-center space-x-1">
                        <button
                          onClick={() => updateAnnotation(annotation.id, {
                            is_private: !annotation.is_private
                          })}
                          className="text-xs p-1 hover:bg-gray-100 rounded"
                          title={annotation.is_private ? 'Make public' : 'Make private'}
                        >
                          {annotation.is_private ? '🔓' : '🔒'}
                        </button>
                        <button
                          onClick={() => deleteAnnotation(annotation.id)}
                          className="text-xs p-1 hover:bg-red-100 text-red-600 rounded"
                          title="Delete annotation"
                        >
                          🗑️
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};