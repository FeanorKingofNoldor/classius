'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface UserStats {
  total_books: number;
  books_read: number;
  pages_read: number;
  total_annotations: number;
  reading_streak: number;
  favorite_genres: { genre: string; count: number }[];
  favorite_authors: { author: string; count: number }[];
  reading_time_minutes: number;
  books_by_language: { language: string; count: number }[];
}

interface UserPreferences {
  reading_goal_books_per_year: number;
  reading_goal_pages_per_day: number;
  preferred_font_size: 'small' | 'medium' | 'large';
  preferred_theme: 'light' | 'dark' | 'sepia';
  enable_reading_reminders: boolean;
  enable_progress_tracking: boolean;
  default_book_privacy: boolean;
  sage_interaction_style: 'concise' | 'detailed' | 'socratic';
}

export default function ProfilePage() {
  const { user, isAuthenticated, updateUser, logout } = useAuthStore();
  const router = useRouter();

  const [activeTab, setActiveTab] = useState<'overview' | 'statistics' | 'preferences' | 'account'>('overview');
  const [stats, setStats] = useState<UserStats | null>(null);
  const [preferences, setPreferences] = useState<UserPreferences>({
    reading_goal_books_per_year: 12,
    reading_goal_pages_per_day: 10,
    preferred_font_size: 'medium',
    preferred_theme: 'light',
    enable_reading_reminders: true,
    enable_progress_tracking: true,
    default_book_privacy: false,
    sage_interaction_style: 'detailed',
  });
  const [loading, setLoading] = useState(false);
  const [editingProfile, setEditingProfile] = useState(false);
  const [profileForm, setProfileForm] = useState({
    full_name: user?.full_name || '',
    username: user?.username || '',
    email: user?.email || '',
  });

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    fetchUserStats();
    fetchUserPreferences();
  }, [isAuthenticated, router]);

  const fetchUserStats = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch('http://localhost:8080/api/user/stats', {
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
      console.error('Error fetching user stats:', err);
    }
  };

  const fetchUserPreferences = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch('http://localhost:8080/api/user/preferences', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setPreferences(result.data);
      }
    } catch (err) {
      console.error('Error fetching user preferences:', err);
    }
  };

  const updateProfile = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch('http://localhost:8080/api/user/profile', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(profileForm),
      });

      if (!response.ok) {
        throw new Error('Failed to update profile');
      }

      const result = await response.json();
      updateUser(result.data);
      setEditingProfile(false);
      toast.success('Profile updated successfully!');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const updatePreferences = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch('http://localhost:8080/api/user/preferences', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(preferences),
      });

      if (!response.ok) {
        throw new Error('Failed to update preferences');
      }

      toast.success('Preferences updated successfully!');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update preferences');
    } finally {
      setLoading(false);
    }
  };

  const formatReadingTime = (minutes: number) => {
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
  };

  if (!isAuthenticated || !user) {
    return null;
  }

  const tabs = [
    { id: 'overview', name: 'Overview', icon: '👤' },
    { id: 'statistics', name: 'Statistics', icon: '📊' },
    { id: 'preferences', name: 'Preferences', icon: '⚙️' },
    { id: 'account', name: 'Account', icon: '🔐' },
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
                <h1 className="text-xl font-semibold text-gray-900">My Profile</h1>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Welcome, {user.username || user.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar Navigation */}
          <div className="lg:col-span-1">
            <nav className="space-y-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md ${
                    activeTab === tab.id
                      ? 'bg-indigo-100 text-indigo-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                  }`}
                >
                  <span className="mr-3">{tab.icon}</span>
                  {tab.name}
                </button>
              ))}
            </nav>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-3">
            <div className="bg-white shadow-sm rounded-lg">
              {/* Overview Tab */}
              {activeTab === 'overview' && (
                <div className="p-6">
                  <div className="flex justify-between items-start mb-6">
                    <h2 className="text-2xl font-bold text-gray-900">Profile Overview</h2>
                    <button
                      onClick={() => setEditingProfile(!editingProfile)}
                      className="px-4 py-2 text-sm text-indigo-600 hover:text-indigo-800 hover:bg-indigo-50 rounded-md"
                    >
                      {editingProfile ? 'Cancel' : 'Edit Profile'}
                    </button>
                  </div>

                  {editingProfile ? (
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">Full Name</label>
                        <input
                          type="text"
                          value={profileForm.full_name}
                          onChange={(e) => setProfileForm(prev => ({ ...prev, full_name: e.target.value }))}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">Username</label>
                        <input
                          type="text"
                          value={profileForm.username}
                          onChange={(e) => setProfileForm(prev => ({ ...prev, username: e.target.value }))}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">Email</label>
                        <input
                          type="email"
                          value={profileForm.email}
                          onChange={(e) => setProfileForm(prev => ({ ...prev, email: e.target.value }))}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        />
                      </div>
                      <div className="flex space-x-3">
                        <button
                          onClick={updateProfile}
                          disabled={loading}
                          className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
                        >
                          {loading ? 'Saving...' : 'Save Changes'}
                        </button>
                        <button
                          onClick={() => setEditingProfile(false)}
                          className="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                        >
                          Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <h3 className="text-lg font-medium text-gray-900 mb-4">Personal Information</h3>
                        <dl className="space-y-3">
                          <div>
                            <dt className="text-sm font-medium text-gray-500">Full Name</dt>
                            <dd className="text-sm text-gray-900">{user.full_name || 'Not provided'}</dd>
                          </div>
                          <div>
                            <dt className="text-sm font-medium text-gray-500">Username</dt>
                            <dd className="text-sm text-gray-900">{user.username}</dd>
                          </div>
                          <div>
                            <dt className="text-sm font-medium text-gray-500">Email</dt>
                            <dd className="text-sm text-gray-900">{user.email}</dd>
                          </div>
                          <div>
                            <dt className="text-sm font-medium text-gray-500">Member Since</dt>
                            <dd className="text-sm text-gray-900">
                              {new Date(user.created_at).toLocaleDateString()}
                            </dd>
                          </div>
                        </dl>
                      </div>

                      <div>
                        <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Stats</h3>
                        <div className="grid grid-cols-2 gap-4">
                          <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="text-2xl font-bold text-indigo-600">{stats?.total_books || 0}</div>
                            <div className="text-sm text-gray-600">Books in Library</div>
                          </div>
                          <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="text-2xl font-bold text-green-600">{stats?.books_read || 0}</div>
                            <div className="text-sm text-gray-600">Books Read</div>
                          </div>
                          <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="text-2xl font-bold text-purple-600">{stats?.total_annotations || 0}</div>
                            <div className="text-sm text-gray-600">Annotations</div>
                          </div>
                          <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="text-2xl font-bold text-orange-600">{stats?.reading_streak || 0}</div>
                            <div className="text-sm text-gray-600">Day Streak</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Statistics Tab */}
              {activeTab === 'statistics' && (
                <div className="p-6">
                  <h2 className="text-2xl font-bold text-gray-900 mb-6">Reading Statistics</h2>
                  
                  {stats ? (
                    <div className="space-y-8">
                      {/* Reading Overview */}
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                        <div className="bg-gradient-to-r from-blue-500 to-indigo-600 p-6 rounded-lg text-white">
                          <div className="text-3xl font-bold">{stats.total_books}</div>
                          <div className="text-blue-100">Total Books</div>
                        </div>
                        <div className="bg-gradient-to-r from-green-500 to-emerald-600 p-6 rounded-lg text-white">
                          <div className="text-3xl font-bold">{stats.books_read}</div>
                          <div className="text-green-100">Books Completed</div>
                        </div>
                        <div className="bg-gradient-to-r from-purple-500 to-violet-600 p-6 rounded-lg text-white">
                          <div className="text-3xl font-bold">{stats.pages_read.toLocaleString()}</div>
                          <div className="text-purple-100">Pages Read</div>
                        </div>
                        <div className="bg-gradient-to-r from-orange-500 to-red-600 p-6 rounded-lg text-white">
                          <div className="text-3xl font-bold">{formatReadingTime(stats.reading_time_minutes)}</div>
                          <div className="text-orange-100">Reading Time</div>
                        </div>
                      </div>

                      {/* Favorite Genres */}
                      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                        <div>
                          <h3 className="text-lg font-medium text-gray-900 mb-4">Favorite Genres</h3>
                          <div className="space-y-3">
                            {stats.favorite_genres.slice(0, 5).map((genre) => (
                              <div key={genre.genre} className="flex justify-between items-center">
                                <span className="text-sm text-gray-700">{genre.genre}</span>
                                <div className="flex items-center">
                                  <div className="w-20 bg-gray-200 rounded-full h-2 mr-2">
                                    <div 
                                      className="bg-indigo-600 h-2 rounded-full" 
                                      style={{ width: `${(genre.count / stats.favorite_genres[0]?.count * 100) || 0}%` }}
                                    />
                                  </div>
                                  <span className="text-sm text-gray-500">{genre.count}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>

                        <div>
                          <h3 className="text-lg font-medium text-gray-900 mb-4">Favorite Authors</h3>
                          <div className="space-y-3">
                            {stats.favorite_authors.slice(0, 5).map((author) => (
                              <div key={author.author} className="flex justify-between items-center">
                                <span className="text-sm text-gray-700">{author.author}</span>
                                <div className="flex items-center">
                                  <div className="w-20 bg-gray-200 rounded-full h-2 mr-2">
                                    <div 
                                      className="bg-green-600 h-2 rounded-full" 
                                      style={{ width: `${(author.count / stats.favorite_authors[0]?.count * 100) || 0}%` }}
                                    />
                                  </div>
                                  <span className="text-sm text-gray-500">{author.count}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>

                      {/* Languages */}
                      <div>
                        <h3 className="text-lg font-medium text-gray-900 mb-4">Books by Language</h3>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                          {stats.books_by_language.map((lang) => (
                            <div key={lang.language} className="text-center p-4 bg-gray-50 rounded-lg">
                              <div className="text-2xl font-bold text-indigo-600">{lang.count}</div>
                              <div className="text-sm text-gray-600">{lang.language.toUpperCase()}</div>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
                      <p className="mt-2 text-gray-600">Loading statistics...</p>
                    </div>
                  )}
                </div>
              )}

              {/* Preferences Tab */}
              {activeTab === 'preferences' && (
                <div className="p-6">
                  <div className="flex justify-between items-start mb-6">
                    <h2 className="text-2xl font-bold text-gray-900">Preferences</h2>
                    <button
                      onClick={updatePreferences}
                      disabled={loading}
                      className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
                    >
                      {loading ? 'Saving...' : 'Save Changes'}
                    </button>
                  </div>

                  <div className="space-y-8">
                    {/* Reading Goals */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Reading Goals</h3>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Books per Year</label>
                          <input
                            type="number"
                            min="1"
                            max="365"
                            value={preferences.reading_goal_books_per_year}
                            onChange={(e) => setPreferences(prev => ({ ...prev, reading_goal_books_per_year: parseInt(e.target.value) }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Pages per Day</label>
                          <input
                            type="number"
                            min="1"
                            max="1000"
                            value={preferences.reading_goal_pages_per_day}
                            onChange={(e) => setPreferences(prev => ({ ...prev, reading_goal_pages_per_day: parseInt(e.target.value) }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </div>
                      </div>
                    </div>

                    {/* Reading Experience */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Reading Experience</h3>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Font Size</label>
                          <select
                            value={preferences.preferred_font_size}
                            onChange={(e) => setPreferences(prev => ({ ...prev, preferred_font_size: e.target.value as any }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                          >
                            <option value="small">Small</option>
                            <option value="medium">Medium</option>
                            <option value="large">Large</option>
                          </select>
                        </div>
                        <div>
                          <label className="block text-sm font-medium text-gray-700">Theme</label>
                          <select
                            value={preferences.preferred_theme}
                            onChange={(e) => setPreferences(prev => ({ ...prev, preferred_theme: e.target.value as any }))}
                            className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                          >
                            <option value="light">Light</option>
                            <option value="dark">Dark</option>
                            <option value="sepia">Sepia</option>
                          </select>
                        </div>
                      </div>
                    </div>

                    {/* AI Sage Preferences */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">AI Sage</h3>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">Interaction Style</label>
                        <select
                          value={preferences.sage_interaction_style}
                          onChange={(e) => setPreferences(prev => ({ ...prev, sage_interaction_style: e.target.value as any }))}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        >
                          <option value="concise">Concise</option>
                          <option value="detailed">Detailed</option>
                          <option value="socratic">Socratic</option>
                        </select>
                        <p className="mt-1 text-sm text-gray-500">
                          Choose how the AI Sage responds to your questions
                        </p>
                      </div>
                    </div>

                    {/* Privacy & Tracking */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Privacy & Tracking</h3>
                      <div className="space-y-4">
                        <div className="flex items-center">
                          <input
                            id="reading-reminders"
                            type="checkbox"
                            checked={preferences.enable_reading_reminders}
                            onChange={(e) => setPreferences(prev => ({ ...prev, enable_reading_reminders: e.target.checked }))}
                            className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                          />
                          <label htmlFor="reading-reminders" className="ml-2 block text-sm text-gray-900">
                            Enable reading reminders
                          </label>
                        </div>
                        <div className="flex items-center">
                          <input
                            id="progress-tracking"
                            type="checkbox"
                            checked={preferences.enable_progress_tracking}
                            onChange={(e) => setPreferences(prev => ({ ...prev, enable_progress_tracking: e.target.checked }))}
                            className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                          />
                          <label htmlFor="progress-tracking" className="ml-2 block text-sm text-gray-900">
                            Track reading progress
                          </label>
                        </div>
                        <div className="flex items-center">
                          <input
                            id="default-privacy"
                            type="checkbox"
                            checked={preferences.default_book_privacy}
                            onChange={(e) => setPreferences(prev => ({ ...prev, default_book_privacy: e.target.checked }))}
                            className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                          />
                          <label htmlFor="default-privacy" className="ml-2 block text-sm text-gray-900">
                            Make uploaded books public by default
                          </label>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Account Tab */}
              {activeTab === 'account' && (
                <div className="p-6">
                  <h2 className="text-2xl font-bold text-gray-900 mb-6">Account Settings</h2>
                  
                  <div className="space-y-8">
                    {/* Account Info */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Account Information</h3>
                      <dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <dt className="text-sm font-medium text-gray-500">Account Type</dt>
                          <dd className="text-sm text-gray-900 capitalize">{user.subscription_tier || 'Free'}</dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">Member Since</dt>
                          <dd className="text-sm text-gray-900">
                            {new Date(user.created_at).toLocaleDateString()}
                          </dd>
                        </div>
                      </dl>
                    </div>

                    {/* Security */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Security</h3>
                      <div className="space-y-4">
                        <button className="w-full md:w-auto px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50">
                          Change Password
                        </button>
                        <div className="text-sm text-gray-500">
                          Last password change: Never
                        </div>
                      </div>
                    </div>

                    {/* Data Export */}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900 mb-4">Data Management</h3>
                      <div className="space-y-4">
                        <div>
                          <button className="w-full md:w-auto px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 mr-3">
                            Export My Data
                          </button>
                          <p className="text-sm text-gray-500 mt-2">
                            Download all your books, annotations, and reading data
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* Danger Zone */}
                    <div className="border-t border-red-200 pt-8">
                      <h3 className="text-lg font-medium text-red-900 mb-4">Danger Zone</h3>
                      <div className="space-y-4">
                        <button
                          onClick={logout}
                          className="w-full md:w-auto px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
                        >
                          Sign Out
                        </button>
                        <button className="w-full md:w-auto px-4 py-2 border border-red-300 text-red-700 rounded-md hover:bg-red-50 ml-3">
                          Delete Account
                        </button>
                        <p className="text-sm text-red-600">
                          These actions cannot be undone. Please be careful.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}