'use client';

import { useState, useEffect, useCallback } from 'react';
import { annotationsApi } from '@/lib/api';
import { toast } from 'react-hot-toast';

export interface Annotation {
  id: string;
  user_id: string;
  book_id: string;
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
}

export interface AnnotationFilter {
  query?: string;
  type?: 'highlight' | 'note' | 'bookmark' | 'all';
  color?: string;
  tags?: string[];
  dateFrom?: string;
  dateTo?: string;
  isPrivate?: boolean;
}

interface UseAnnotationsReturn {
  annotations: Annotation[];
  filteredAnnotations: Annotation[];
  loading: boolean;
  creating: boolean;
  filter: AnnotationFilter;
  
  // Core operations
  createAnnotation: (annotation: Omit<Annotation, 'id' | 'user_id' | 'created_at' | 'updated_at'>) => Promise<Annotation | null>;
  updateAnnotation: (id: string, updates: Partial<Annotation>) => Promise<void>;
  deleteAnnotation: (id: string) => Promise<void>;
  
  // Filtering
  setFilter: (filter: Partial<AnnotationFilter>) => void;
  clearFilter: () => void;
  
  // Batch operations
  bulkDelete: (annotationIds: string[]) => Promise<void>;
  bulkUpdateTags: (annotationIds: string[], tags: string[]) => Promise<void>;
  bulkUpdateColor: (annotationIds: string[], color: string) => Promise<void>;
  
  // Export
  exportAnnotations: (format: 'json' | 'csv' | 'pdf') => Promise<void>;
  
  // Search
  searchAnnotations: (query: string) => Annotation[];
  
  // Stats
  getStats: () => {
    total: number;
    highlights: number;
    notes: number;
    bookmarks: number;
    byColor: Record<string, number>;
    byMonth: Record<string, number>;
  };
}

export const useAnnotations = (bookId: string): UseAnnotationsReturn => {
  const [annotations, setAnnotations] = useState<Annotation[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [filter, setFilterState] = useState<AnnotationFilter>({ type: 'all' });

  // Load annotations for the book
  const loadAnnotations = useCallback(async () => {
    try {
      setLoading(true);
      const response = await annotationsApi.getAnnotations({ book_id: bookId });
      setAnnotations(response.data.data || []);
    } catch (error) {
      console.error('Failed to load annotations:', error);
      toast.error('Failed to load annotations');
    } finally {
      setLoading(false);
    }
  }, [bookId]);

  useEffect(() => {
    if (bookId) {
      loadAnnotations();
    }
  }, [bookId, loadAnnotations]);

  // Create annotation
  const createAnnotation = useCallback(async (
    annotationData: Omit<Annotation, 'id' | 'user_id' | 'created_at' | 'updated_at'>
  ): Promise<Annotation | null> => {
    try {
      setCreating(true);
      const response = await annotationsApi.createAnnotation(annotationData);
      const newAnnotation = response.data.data;
      setAnnotations(prev => [...prev, newAnnotation]);
      toast.success(`${annotationData.type} created successfully`);
      return newAnnotation;
    } catch (error) {
      console.error('Failed to create annotation:', error);
      toast.error('Failed to create annotation');
      return null;
    } finally {
      setCreating(false);
    }
  }, []);

  // Update annotation
  const updateAnnotation = useCallback(async (id: string, updates: Partial<Annotation>) => {
    try {
      await annotationsApi.updateAnnotation(id, updates);
      setAnnotations(prev =>
        prev.map(annotation =>
          annotation.id === id ? { ...annotation, ...updates } : annotation
        )
      );
      toast.success('Annotation updated');
    } catch (error) {
      console.error('Failed to update annotation:', error);
      toast.error('Failed to update annotation');
    }
  }, []);

  // Delete annotation
  const deleteAnnotation = useCallback(async (id: string) => {
    try {
      await annotationsApi.deleteAnnotation(id);
      setAnnotations(prev => prev.filter(annotation => annotation.id !== id));
      toast.success('Annotation deleted');
    } catch (error) {
      console.error('Failed to delete annotation:', error);
      toast.error('Failed to delete annotation');
    }
  }, []);

  // Filter annotations
  const filteredAnnotations = annotations.filter(annotation => {
    if (filter.query) {
      const query = filter.query.toLowerCase();
      const matchesText = annotation.selected_text?.toLowerCase().includes(query) ||
                         annotation.content?.toLowerCase().includes(query);
      if (!matchesText) return false;
    }

    if (filter.type && filter.type !== 'all' && annotation.type !== filter.type) {
      return false;
    }

    if (filter.color && annotation.color !== filter.color) {
      return false;
    }

    if (filter.tags && filter.tags.length > 0) {
      const annotationTags = annotation.tags || [];
      const hasAllTags = filter.tags.every(tag => annotationTags.includes(tag));
      if (!hasAllTags) return false;
    }

    if (filter.dateFrom) {
      const annotationDate = new Date(annotation.created_at);
      const fromDate = new Date(filter.dateFrom);
      if (annotationDate < fromDate) return false;
    }

    if (filter.dateTo) {
      const annotationDate = new Date(annotation.created_at);
      const toDate = new Date(filter.dateTo);
      if (annotationDate > toDate) return false;
    }

    if (filter.isPrivate !== undefined && annotation.is_private !== filter.isPrivate) {
      return false;
    }

    return true;
  });

  // Set filter
  const setFilter = useCallback((newFilter: Partial<AnnotationFilter>) => {
    setFilterState(prev => ({ ...prev, ...newFilter }));
  }, []);

  // Clear filter
  const clearFilter = useCallback(() => {
    setFilterState({ type: 'all' });
  }, []);

  // Bulk delete
  const bulkDelete = useCallback(async (annotationIds: string[]) => {
    try {
      await Promise.all(annotationIds.map(id => annotationsApi.deleteAnnotation(id)));
      setAnnotations(prev => prev.filter(annotation => !annotationIds.includes(annotation.id)));
      toast.success(`${annotationIds.length} annotations deleted`);
    } catch (error) {
      console.error('Failed to delete annotations:', error);
      toast.error('Failed to delete annotations');
    }
  }, []);

  // Bulk update tags
  const bulkUpdateTags = useCallback(async (annotationIds: string[], tags: string[]) => {
    try {
      await Promise.all(annotationIds.map(id => annotationsApi.updateAnnotation(id, { tags })));
      setAnnotations(prev =>
        prev.map(annotation =>
          annotationIds.includes(annotation.id) ? { ...annotation, tags } : annotation
        )
      );
      toast.success(`Updated tags for ${annotationIds.length} annotations`);
    } catch (error) {
      console.error('Failed to update tags:', error);
      toast.error('Failed to update tags');
    }
  }, []);

  // Bulk update color
  const bulkUpdateColor = useCallback(async (annotationIds: string[], color: string) => {
    try {
      await Promise.all(annotationIds.map(id => annotationsApi.updateAnnotation(id, { color })));
      setAnnotations(prev =>
        prev.map(annotation =>
          annotationIds.includes(annotation.id) ? { ...annotation, color } : annotation
        )
      );
      toast.success(`Updated color for ${annotationIds.length} annotations`);
    } catch (error) {
      console.error('Failed to update color:', error);
      toast.error('Failed to update color');
    }
  }, []);

  // Export annotations
  const exportAnnotations = useCallback(async (format: 'json' | 'csv' | 'pdf') => {
    try {
      // This would call a backend export endpoint
      const response = await fetch(`/api/v1/annotations/export?book_id=${bookId}&format=${format}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });

      if (!response.ok) throw new Error('Export failed');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `annotations_${bookId}.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success(`Annotations exported as ${format.toUpperCase()}`);
    } catch (error) {
      console.error('Failed to export annotations:', error);
      toast.error('Failed to export annotations');
    }
  }, [bookId]);

  // Search annotations
  const searchAnnotations = useCallback((query: string): Annotation[] => {
    const lowerQuery = query.toLowerCase();
    return annotations.filter(annotation =>
      annotation.selected_text?.toLowerCase().includes(lowerQuery) ||
      annotation.content?.toLowerCase().includes(lowerQuery) ||
      annotation.tags?.some(tag => tag.toLowerCase().includes(lowerQuery))
    );
  }, [annotations]);

  // Get stats
  const getStats = useCallback(() => {
    const stats = {
      total: annotations.length,
      highlights: 0,
      notes: 0,
      bookmarks: 0,
      byColor: {} as Record<string, number>,
      byMonth: {} as Record<string, number>,
    };

    annotations.forEach(annotation => {
      // Count by type
      if (annotation.type === 'highlight') stats.highlights++;
      else if (annotation.type === 'note') stats.notes++;
      else if (annotation.type === 'bookmark') stats.bookmarks++;

      // Count by color
      const color = annotation.color || 'default';
      stats.byColor[color] = (stats.byColor[color] || 0) + 1;

      // Count by month
      const month = new Date(annotation.created_at).toISOString().slice(0, 7);
      stats.byMonth[month] = (stats.byMonth[month] || 0) + 1;
    });

    return stats;
  }, [annotations]);

  return {
    annotations,
    filteredAnnotations,
    loading,
    creating,
    filter,
    createAnnotation,
    updateAnnotation,
    deleteAnnotation,
    setFilter,
    clearFilter,
    bulkDelete,
    bulkUpdateTags,
    bulkUpdateColor,
    exportAnnotations,
    searchAnnotations,
    getStats,
  };
};