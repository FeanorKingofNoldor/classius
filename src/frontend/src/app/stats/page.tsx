'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';
import {
  MonthlyProgressChart,
  ReadingHabitsChart,
  GenreDistributionChart,
  ReadingGoalChart,
  ReadingStreakChart,
  ContentAnalyticsChart,
  AnnotationEngagementChart,
} from '@/components/charts/ReadingCharts';

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

export default function BookStatsPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();

  const [stats, setStats] = useState<BookStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'overview' | 'habits' | 'content' | 'goals' | 'annotations'>('overview');
  const [timeRange, setTimeRange] = useState<'week' | 'month' | 'year' | 'all'>('year');

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    fetchBookStats();
  }, [isAuthenticated, router, timeRange]);

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

  const formatFileSize = (mb: number) => {
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(1)}GB`;
    }
    return `${mb.toFixed(1)}MB`;
  };

  if (!isAuthenticated) {
    return null;
  }

  const tabs = [
    { id: 'overview', name: 'Overview', icon: '📊' },
    { id: 'habits', name: 'Reading Habits', icon: '📈' },
    { id: 'content', name: 'Content Analysis', icon: '📚' },
    { id: 'goals', name: 'Goals & Progress', icon: '🎯' },
    { id: 'annotations', name: 'Engagement', icon: '✍️' },
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
                  <span className="text-2xl mr-2">📊</span>
                  Reading Statistics
                </h1>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value as any)}
                className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="week">This Week</option>
                <option value="month">This Month</option>
                <option value="year">This Year</option>
                <option value="all">All Time</option>
              </select>
              <span className="text-gray-700">Welcome, {user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-8">
          <nav className="-mb-px flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm flex items-center space-x-2 ${
                  activeTab === tab.id
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <span>{tab.icon}</span>
                <span>{tab.name}</span>
              </button>
            ))}
          </nav>
        </div>

        {loading ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading statistics...</p>
          </div>
        ) : !stats ? (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">📊</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No data available</h3>
            <p className="text-gray-600">Start reading books to see your statistics!</p>
          </div>
        ) : (
          <>
            {/* Overview Tab */}
            {activeTab === 'overview' && (
              <div className="space-y-8">
                {/* Key Metrics */}
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

                {/* Additional Metrics */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <h3 className="text-lg font-medium text-gray-900 mb-4">Reading Speed</h3>
                    <div className="text-3xl font-bold text-indigo-600 mb-2">
                      {stats.overview.average_reading_speed_pages_per_hour}
                    </div>
                    <div className="text-sm text-gray-600">Pages per hour average</div>
                  </div>

                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <h3 className="text-lg font-medium text-gray-900 mb-4">Current Streak</h3>
                    <div className="text-3xl font-bold text-green-600 mb-2">
                      {stats.overview.current_reading_streak_days}
                    </div>
                    <div className="text-sm text-gray-600">Days of continuous reading</div>
                  </div>

                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <h3 className="text-lg font-medium text-gray-900 mb-4">Best Streak</h3>
                    <div className="text-3xl font-bold text-orange-600 mb-2">
                      {stats.overview.longest_reading_streak_days}
                    </div>
                    <div className="text-sm text-gray-600">Days (personal record)</div>
                  </div>
                </div>

                {/* Progress Overview */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">Reading Progress</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <div className="flex justify-between text-sm text-gray-600 mb-2">
                        <span>Books Completed</span>
                        <span>{stats.overview.books_read} / {stats.overview.total_books}</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-3">
                        <div 
                          className="bg-green-600 h-3 rounded-full"
                          style={{ width: `${(stats.overview.books_read / stats.overview.total_books) * 100}%` }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="flex justify-between text-sm text-gray-600 mb-2">
                        <span>Books In Progress</span>
                        <span>{stats.overview.books_in_progress}</span>
                      </div>
                      <div className="text-2xl font-bold text-indigo-600">
                        {stats.overview.books_in_progress}
                      </div>
                      <div className="text-sm text-gray-500">Currently reading</div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Reading Habits Tab */}
            {activeTab === 'habits' && (
              <div className="space-y-8">
                {/* Daily Averages */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-6">Daily Reading Habits</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="text-center p-4 bg-gray-50 rounded-lg">
                      <div className="text-2xl font-bold text-indigo-600">{stats.reading_habits.daily_averages.pages_per_day}</div>
                      <div className="text-sm text-gray-600">Pages per day</div>
                    </div>
                    <div className="text-center p-4 bg-gray-50 rounded-lg">
                      <div className="text-2xl font-bold text-green-600">{stats.reading_habits.daily_averages.minutes_per_day}</div>
                      <div className="text-sm text-gray-600">Minutes per day</div>
                    </div>
                    <div className="text-center p-4 bg-gray-50 rounded-lg">
                      <div className="text-2xl font-bold text-purple-600">{stats.reading_habits.daily_averages.books_per_month}</div>
                      <div className="text-sm text-gray-600">Books per month</div>
                    </div>
                  </div>
                </div>

                {/* Monthly Progress Chart */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-6">Monthly Progress</h3>
                  <MonthlyProgressChart data={stats.reading_habits.monthly_progress} />
                </div>

                {/* Reading Habits Charts */}
                <ReadingHabitsChart 
                  weeklyData={stats.reading_habits.reading_by_day_of_week}
                  hourlyData={stats.reading_habits.reading_by_hour}
                />

                {/* Reading Streak */}
                <ReadingStreakChart 
                  currentStreak={stats.overview.current_reading_streak_days}
                  longestStreak={stats.overview.longest_reading_streak_days}
                />
              </div>
            )}

            {/* Content Analysis Tab */}
            {activeTab === 'content' && (
              <div className="space-y-8">
                {/* Genre Distribution Chart */}
                <GenreDistributionChart data={stats.content_analytics.genres} />

                {/* Languages and Authors Charts */}
                <ContentAnalyticsChart 
                  languages={stats.content_analytics.languages}
                  authors={stats.content_analytics.authors}
                />

                {/* File Formats */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-6">Library Composition</h3>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {stats.content_analytics.file_formats.map((format, index) => (
                      <div key={index} className="text-center p-4 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-indigo-600">{format.count}</div>
                        <div className="text-sm text-gray-600 uppercase font-medium">{format.format}</div>
                        <div className="text-xs text-gray-500 mt-1">{formatFileSize(format.total_size_mb)}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Goals & Progress Tab */}
            {activeTab === 'goals' && (
              <div className="space-y-8">
                {/* Goal Progress Chart */}
                <ReadingGoalChart 
                  yearlyGoal={stats.goals_and_progress.yearly_goal}
                  dailyGoal={stats.goals_and_progress.daily_goal}
                />

                {/* Individual Goal Details */}
                {stats.goals_and_progress.yearly_goal && (
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <h3 className="text-lg font-medium text-gray-900 mb-6">Yearly Reading Goal</h3>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                      <div className="text-center">
                        <div className="text-3xl font-bold text-indigo-600 mb-2">
                          {stats.goals_and_progress.yearly_goal.completed_books}
                        </div>
                        <div className="text-sm text-gray-600">Books Read</div>
                      </div>
                      <div className="text-center">
                        <div className="text-3xl font-bold text-gray-400 mb-2">
                          {stats.goals_and_progress.yearly_goal.target_books}
                        </div>
                        <div className="text-sm text-gray-600">Goal</div>
                      </div>
                      <div className="text-center">
                        <div className="text-3xl font-bold text-green-600 mb-2">
                          {Math.round(stats.goals_and_progress.yearly_goal.progress_percentage)}%
                        </div>
                        <div className="text-sm text-gray-600">Complete</div>
                      </div>
                    </div>
                    <div className="mt-6">
                      <div className="flex justify-between text-sm text-gray-600 mb-2">
                        <span>Progress to Goal</span>
                        <span>Projected completion: {stats.goals_and_progress.yearly_goal.projected_completion}</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-3">
                        <div 
                          className="bg-indigo-600 h-3 rounded-full"
                          style={{ width: `${stats.goals_and_progress.yearly_goal.progress_percentage}%` }}
                        />
                      </div>
                    </div>
                  </div>
                )}

                {stats.goals_and_progress.daily_goal && (
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <h3 className="text-lg font-medium text-gray-900 mb-6">Daily Reading Goal</h3>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                      <div className="text-center p-4 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-indigo-600">{stats.goals_and_progress.daily_goal.target_pages}</div>
                        <div className="text-sm text-gray-600">Daily Target</div>
                      </div>
                      <div className="text-center p-4 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-green-600">{stats.goals_and_progress.daily_goal.average_achieved}</div>
                        <div className="text-sm text-gray-600">Daily Average</div>
                      </div>
                      <div className="text-center p-4 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-purple-600">{Math.round(stats.goals_and_progress.daily_goal.success_rate)}%</div>
                        <div className="text-sm text-gray-600">Success Rate</div>
                      </div>
                    </div>
                  </div>
                )}

                {/* Milestones */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-6">Reading Milestones</h3>
                  <div className="space-y-4">
                    {stats.goals_and_progress.reading_milestones.map((milestone, index) => (
                      <div key={index} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div className="flex items-center space-x-3">
                          <div className={`w-3 h-3 rounded-full ${milestone.achieved_date ? 'bg-green-500' : 'bg-gray-300'}`}></div>
                          <span className="font-medium text-gray-900">{milestone.milestone}</span>
                        </div>
                        <div className="text-sm text-gray-500">
                          {milestone.achieved_date ? (
                            `Achieved ${new Date(milestone.achieved_date).toLocaleDateString()}`
                          ) : milestone.progress ? (
                            `${milestone.progress}% complete`
                          ) : (
                            'Not started'
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Annotations & Engagement Tab */}
            {activeTab === 'annotations' && (
              <div className="space-y-8">
                {/* Engagement Overview */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                  <div className="bg-white p-6 rounded-lg shadow-sm border text-center">
                    <div className="text-3xl font-bold text-indigo-600 mb-2">{stats.annotations_and_engagement.total_annotations}</div>
                    <div className="text-sm text-gray-600">Total Annotations</div>
                  </div>
                  <div className="bg-white p-6 rounded-lg shadow-sm border text-center">
                    <div className="text-3xl font-bold text-yellow-600 mb-2">{stats.annotations_and_engagement.highlights}</div>
                    <div className="text-sm text-gray-600">Highlights</div>
                  </div>
                  <div className="bg-white p-6 rounded-lg shadow-sm border text-center">
                    <div className="text-3xl font-bold text-green-600 mb-2">{stats.annotations_and_engagement.notes}</div>
                    <div className="text-sm text-gray-600">Notes</div>
                  </div>
                  <div className="bg-white p-6 rounded-lg shadow-sm border text-center">
                    <div className="text-3xl font-bold text-purple-600 mb-2">{stats.annotations_and_engagement.bookmarks}</div>
                    <div className="text-sm text-gray-600">Bookmarks</div>
                  </div>
                </div>

                {/* Annotation Engagement Charts */}
                <AnnotationEngagementChart 
                  totalAnnotations={stats.annotations_and_engagement.total_annotations}
                  highlights={stats.annotations_and_engagement.highlights}
                  notes={stats.annotations_and_engagement.notes}
                  bookmarks={stats.annotations_and_engagement.bookmarks}
                  mostAnnotatedBooks={stats.annotations_and_engagement.most_annotated_books}
                />

                {/* Engagement Insights */}
                <div className="bg-white p-6 rounded-lg shadow-sm border">
                  <h3 className="text-lg font-medium text-gray-900 mb-6">Reading Engagement</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <div className="text-center p-6 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-indigo-600 mb-2">
                          {stats.overview.total_books > 0 ? Math.round(stats.annotations_and_engagement.total_annotations / stats.overview.total_books) : 0}
                        </div>
                        <div className="text-sm text-gray-600">Average annotations per book</div>
                      </div>
                    </div>
                    <div>
                      <div className="text-center p-6 bg-gray-50 rounded-lg">
                        <div className="text-2xl font-bold text-green-600 mb-2">
                          {stats.overview.total_pages_read > 0 ? ((stats.annotations_and_engagement.total_annotations / stats.overview.total_pages_read) * 100).toFixed(1) : 0}%
                        </div>
                        <div className="text-sm text-gray-600">Pages with annotations</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}