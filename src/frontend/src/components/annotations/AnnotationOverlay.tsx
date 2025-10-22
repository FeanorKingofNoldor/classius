'use client';

import React, { useState } from 'react';
import { Annotation } from '@/hooks/useAnnotations';

interface AnnotationOverlayProps {
  annotation: Annotation;
  position: { x: number; y: number; width: number; height: number };
  onUpdate?: (annotation: Annotation) => void;
  onDelete?: (id: string) => void;
  theme?: 'light' | 'dark' | 'sepia';
}

export const AnnotationOverlay: React.FC<AnnotationOverlayProps> = ({
  annotation,
  position,
  onUpdate,
  onDelete,
  theme = 'light',
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(annotation.content || '');

  const handleSave = () => {
    if (onUpdate) {
      onUpdate({
        ...annotation,
        content: editContent.trim(),
        updated_at: new Date().toISOString(),
      });
    }
    setIsEditing(false);
  };

  const handleDelete = () => {
    if (onDelete && confirm('Are you sure you want to delete this annotation?')) {
      onDelete(annotation.id);
    }
  };

  const getAnnotationIcon = () => {
    switch (annotation.type) {
      case 'highlight':
        return '🖍️';
      case 'note':
        return '📝';
      case 'bookmark':
        return '🔖';
      default:
        return '📌';
    }
  };

  const getAnnotationColor = () => {
    return annotation.color || '#fef3c7';
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
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

  return (
    <>
      {/* Highlight Background (for highlight type) */}
      {annotation.type === 'highlight' && (
        <div
          className="absolute pointer-events-none rounded-sm opacity-30"
          style={{
            left: position.x,
            top: position.y,
            width: position.width,
            height: position.height,
            backgroundColor: getAnnotationColor(),
          }}
        />
      )}

      {/* Annotation Indicator */}
      <div
        className="absolute cursor-pointer group"
        style={{
          left: annotation.type === 'bookmark' ? position.x - 20 : position.x + position.width + 2,
          top: position.y,
        }}
        onClick={() => setShowDetails(!showDetails)}
      >
        <div
          className={`w-6 h-6 rounded-full flex items-center justify-center text-xs border-2 border-white shadow-md transition-transform group-hover:scale-110 ${
            annotation.type === 'bookmark' ? 'bg-green-500' : 'bg-blue-500'
          }`}
          style={{
            backgroundColor: annotation.type === 'highlight' ? getAnnotationColor() : undefined,
          }}
        >
          {getAnnotationIcon()}
        </div>

        {/* Tooltip on hover */}
        <div className="absolute left-8 top-0 opacity-0 group-hover:opacity-100 transition-opacity delay-300 pointer-events-none z-50">
          <div className={`px-2 py-1 text-xs rounded shadow-lg whitespace-nowrap max-w-xs ${getThemeClasses()}`}>
            <div className="font-medium">
              {annotation.type.charAt(0).toUpperCase() + annotation.type.slice(1)}
            </div>
            {annotation.selected_text && (
              <div className="opacity-70 truncate max-w-40">
                "{annotation.selected_text.slice(0, 50)}..."
              </div>
            )}
            {annotation.content && (
              <div className="opacity-70 mt-1 truncate max-w-40">
                {annotation.content.slice(0, 50)}...
              </div>
            )}
            <div className="text-xs opacity-50 mt-1">
              {formatDate(annotation.created_at)}
            </div>
          </div>
        </div>
      </div>

      {/* Detailed Popup */}
      {showDetails && (
        <div
          className={`absolute z-50 w-80 rounded-lg shadow-xl border p-4 ${getThemeClasses()}`}
          style={{
            left: Math.max(10, Math.min(position.x, window.innerWidth - 330)),
            top: position.y + position.height + 10,
          }}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center space-x-2">
              <span className="text-lg">{getAnnotationIcon()}</span>
              <span className="font-medium text-sm capitalize">
                {annotation.type}
              </span>
              {annotation.tags && annotation.tags.length > 0 && (
                <div className="flex flex-wrap gap-1">
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
            </div>
            <button
              onClick={() => setShowDetails(false)}
              className="text-gray-400 hover:text-gray-600 text-lg"
            >
              ×
            </button>
          </div>

          {/* Selected Text */}
          {annotation.selected_text && (
            <div className="mb-3 p-3 bg-opacity-20 rounded" style={{ backgroundColor: getAnnotationColor() }}>
              <p className="text-sm italic">
                "{annotation.selected_text}"
              </p>
            </div>
          )}

          {/* Content/Note */}
          {annotation.type !== 'bookmark' && (
            <div className="mb-3">
              {isEditing ? (
                <div className="space-y-2">
                  <textarea
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    className={`w-full px-3 py-2 text-sm border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                      theme === 'dark' 
                        ? 'bg-gray-700 border-gray-600 text-gray-100' 
                        : 'bg-white border-gray-300 text-gray-900'
                    }`}
                    rows={3}
                    placeholder={annotation.type === 'note' ? 'Add your note...' : 'Add description...'}
                  />
                  <div className="flex gap-2">
                    <button
                      onClick={handleSave}
                      className="px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                    >
                      Save
                    </button>
                    <button
                      onClick={() => {
                        setIsEditing(false);
                        setEditContent(annotation.content || '');
                      }}
                      className="px-3 py-1 border border-gray-300 text-gray-700 text-sm rounded hover:bg-gray-50"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div>
                  {annotation.content ? (
                    <p className="text-sm whitespace-pre-wrap">{annotation.content}</p>
                  ) : (
                    <p className="text-sm italic opacity-60">
                      {annotation.type === 'note' ? 'No note added' : 'No description'}
                    </p>
                  )}
                  {onUpdate && (
                    <button
                      onClick={() => setIsEditing(true)}
                      className="mt-2 text-xs text-blue-600 hover:text-blue-800"
                    >
                      {annotation.content ? 'Edit' : 'Add note'}
                    </button>
                  )}
                </div>
              )}
            </div>
          )}

          {/* Metadata */}
          <div className="text-xs opacity-60 space-y-1">
            <p>Created: {formatDate(annotation.created_at)}</p>
            {annotation.updated_at !== annotation.created_at && (
              <p>Updated: {formatDate(annotation.updated_at)}</p>
            )}
            {annotation.page_number && (
              <p>Page: {annotation.page_number}</p>
            )}
            <p>Privacy: {annotation.is_private ? 'Private' : 'Public'}</p>
          </div>

          {/* Actions */}
          <div className="flex gap-2 mt-4 pt-3 border-t border-opacity-20 border-gray-400">
            {onUpdate && (
              <button
                onClick={() => onUpdate({
                  ...annotation,
                  is_private: !annotation.is_private,
                  updated_at: new Date().toISOString(),
                })}
                className="flex-1 px-3 py-2 text-xs border border-gray-300 rounded hover:bg-gray-50 transition-colors"
              >
                {annotation.is_private ? '🔓 Make Public' : '🔒 Make Private'}
              </button>
            )}
            
            {onDelete && (
              <button
                onClick={handleDelete}
                className="px-3 py-2 text-xs text-red-600 border border-red-300 rounded hover:bg-red-50 transition-colors"
              >
                🗑️ Delete
              </button>
            )}
          </div>
        </div>
      )}

      {/* Click outside handler for popup */}
      {showDetails && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setShowDetails(false)}
        />
      )}
    </>
  );
};