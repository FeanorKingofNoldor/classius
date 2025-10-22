'use client';

import React, { useState, useEffect } from 'react';
import { toast } from 'react-hot-toast';
import {
  MonthlyProgressChart,
  ReadingHabitsChart,
  GenreDistributionChart,
  ReadingGoalChart,
  ReadingStreakChart,
  ContentAnalyticsChart,
  AnnotationEngagementChart,
} from './charts/ReadingCharts';

interface BookStats {
  overview: {
    total_books: number;
    books_read: number;
    books_in_progress: number;
    total_pages_read: number;
    total_reading_time_minutes: number;
    average_reading_speed_pages_per_hour: number;
    current_reading_streak_days: number;
    longest_reading_streak_days: number;
  };
  
  reading_habits: {
    daily_averages: {
      pages_per_day: number;
      minutes_per_day: number;
      books_per_month: number;
    };
    monthly_progress: Array<{
      month: string;
      pages_read: number;
      books_completed: number;
      reading_time_minutes: number;
    }>;
    reading_by_day_of_week: Array<{
      day: string;
      pages_read: number;
      sessions: number;
    }>;
    reading_by_hour: Array<{
      hour: number;
      pages_read: number;
      sessions: number;
    }>;
  };

  content_analytics: {
    genres: Array<{
      genre: string;
      book_count: number;
      pages_read: number;
      completion_rate: number;
    }>;
    languages: Array<{
      language: string;
      book_count: number;
      pages_read: number;
      reading_time_minutes: number;
    }>;
    authors: Array<{
      author: string;
      book_count: number;
      pages_read: number;
      favorite_book?: string;
    }>;
    file_formats: Array<{
      format: string;
      count: number;
      total_size_mb: number;
    }>;
  };

  goals_and_progress: {
    yearly_goal?: {
      target_books: number;
      completed_books: number;
      progress_percentage: number;
      projected_completion: string;
    };
    daily_goal?: {
      target_pages: number;
      average_achieved: number;
      success_rate: number;
    };
    reading_milestones: Array<{
      milestone: string;
      achieved_date?: string;
      progress?: number;
    }>;
  };

  annotations_and_engagement: {
    total_annotations: number;
    highlights: number;
    notes: number;
    bookmarks: number;
    most_annotated_books: Array<{
      title: string;
      author: string;
      annotation_count: number;
    }>;
  };
}

interface AnalyticsDashboardProps {
  timeRange?: 'week' | 'month' | 'year' | 'all';
  className?: string;
}

export const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = ({ 
  timeRange = 'year',
  className = '' 
}) => {
  const [stats, setStats] = useState<BookStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchBookStats();
  }, [timeRange]);

  const fetchBookStats = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch(`http://localhost:8082/api/stats/books?range=${timeRange}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch statistics');
      }

      const result = await response.json();
      setStats(result.data);
    } catch (err) {
      console.error('Error fetching book stats:', err);
      toast.error('Failed to load statistics');
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (minutes: number) => {
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
  };

  if (loading) {
    return (
      <div className={`${className}`}>
        <div className="text-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading analytics...</p>
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className={`${className}`}>
        <div className="text-center py-12">
          <div className="text-6xl mb-4">📊</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No data available</h3>
          <p className="text-gray-600">Start reading books to see your analytics!</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-8 ${className}`}>
      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-gradient-to-r from-blue-500 to-indigo-600 p-6 rounded-lg text-white">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-3xl font-bold">{stats.overview.total_books}</div>
              <div className="text-blue-100">Total Books</div>
            </div>
            <div className="text-4xl">📚</div>
          </div>
        </div>

        <div className="bg-gradient-to-r from-green-500 to-emerald-600 p-6 rounded-lg text-white">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-3xl font-bold">{stats.overview.books_read}</div>
              <div className="text-green-100">Books Completed</div>
            </div>
            <div className="text-4xl">✅</div>
          </div>
        </div>

        <div className="bg-gradient-to-r from-purple-500 to-violet-600 p-6 rounded-lg text-white">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-3xl font-bold">{stats.overview.total_pages_read.toLocaleString()}</div>
              <div className="text-purple-100">Pages Read</div>
            </div>
            <div className="text-4xl">📄</div>
          </div>
        </div>

        <div className="bg-gradient-to-r from-orange-500 to-red-600 p-6 rounded-lg text-white">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-3xl font-bold">{formatTime(stats.overview.total_reading_time_minutes)}</div>
              <div className="text-orange-100">Reading Time</div>
            </div>
            <div className="text-4xl">⏱️</div>
          </div>
        </div>
      </div>

      {/* Monthly Progress Chart */}
      <div className="bg-white p-6 rounded-lg shadow-sm border">
        <h3 className="text-lg font-medium text-gray-900 mb-6">Monthly Reading Progress</h3>
        <MonthlyProgressChart data={stats.reading_habits.monthly_progress} />
      </div>

      {/* Reading Goals */}
      {(stats.goals_and_progress.yearly_goal || stats.goals_and_progress.daily_goal) && (
        <ReadingGoalChart 
          yearlyGoal={stats.goals_and_progress.yearly_goal}
          dailyGoal={stats.goals_and_progress.daily_goal}
        />
      )}

      {/* Reading Habits and Streak */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <ReadingStreakChart 
          currentStreak={stats.overview.current_reading_streak_days}
          longestStreak={stats.overview.longest_reading_streak_days}
        />
        
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <h3 className="text-lg font-medium text-gray-900 mb-6">Reading Habits Summary</h3>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-indigo-600">{stats.reading_habits.daily_averages.pages_per_day}</div>
              <div className="text-sm text-gray-600">Pages/day</div>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-green-600">{stats.reading_habits.daily_averages.minutes_per_day}</div>
              <div className="text-sm text-gray-600">Minutes/day</div>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-purple-600">{stats.reading_habits.daily_averages.books_per_month}</div>
              <div className="text-sm text-gray-600">Books/month</div>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-orange-600">{stats.overview.average_reading_speed_pages_per_hour}</div>
              <div className="text-sm text-gray-600">Pages/hour</div>
            </div>
          </div>
        </div>
      </div>

      {/* Genre Distribution */}
      <GenreDistributionChart data={stats.content_analytics.genres} />

      {/* Content Analytics */}
      <ContentAnalyticsChart 
        languages={stats.content_analytics.languages}
        authors={stats.content_analytics.authors}
      />

      {/* Annotation Engagement */}
      <AnnotationEngagementChart 
        totalAnnotations={stats.annotations_and_engagement.total_annotations}
        highlights={stats.annotations_and_engagement.highlights}
        notes={stats.annotations_and_engagement.notes}
        bookmarks={stats.annotations_and_engagement.bookmarks}
        mostAnnotatedBooks={stats.annotations_and_engagement.most_annotated_books}
      />

      {/* Reading Habits Detailed Charts */}
      <ReadingHabitsChart 
        weeklyData={stats.reading_habits.reading_by_day_of_week}
        hourlyData={stats.reading_habits.reading_by_hour}
      />
    </div>
  );
};