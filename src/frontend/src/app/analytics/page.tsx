'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import Layout from '@/components/Layout';
import { AnalyticsDashboard } from '@/components/AnalyticsDashboard';

export default function AnalyticsPage() {
  const { isAuthenticated, checkAuth, user } = useAuthStore();
  const router = useRouter();
  const [timeRange, setTimeRange] = useState<'week' | 'month' | 'year' | 'all'>('year');

  useEffect(() => {
    checkAuth();
    if (!isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, checkAuth, router]);

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  return (
    <Layout title="Reading Analytics">
      <div className="space-y-8">
        {/* Header */}
        <div className="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-lg px-6 py-8 text-white">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold mb-2 flex items-center">
                <span className="text-4xl mr-3">📈</span>
                Reading Analytics
              </h1>
              <p className="text-indigo-100">
                Discover insights about your reading journey with interactive visualizations.
              </p>
            </div>
            
            {/* Time Range Selector */}
            <div className="bg-white/10 backdrop-blur-sm rounded-lg p-3">
              <select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value as any)}
                className="bg-transparent text-white font-medium focus:outline-none cursor-pointer"
              >
                <option value="week" className="text-gray-900">This Week</option>
                <option value="month" className="text-gray-900">This Month</option>
                <option value="year" className="text-gray-900">This Year</option>
                <option value="all" className="text-gray-900">All Time</option>
              </select>
            </div>
          </div>
        </div>

        {/* Quick Navigation */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Link
            href="/stats"
            className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow text-center"
          >
            <div className="text-2xl mb-2">📊</div>
            <div className="text-sm font-medium text-gray-900">Detailed Stats</div>
          </Link>
          
          <Link
            href="/books"
            className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow text-center"
          >
            <div className="text-2xl mb-2">📚</div>
            <div className="text-sm font-medium text-gray-900">My Library</div>
          </Link>
          
          <Link
            href="/annotations"
            className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow text-center"
          >
            <div className="text-2xl mb-2">✍️</div>
            <div className="text-sm font-medium text-gray-900">Annotations</div>
          </Link>
          
          <Link
            href="/sage"
            className="bg-white p-4 rounded-lg shadow-sm border hover:shadow-md transition-shadow text-center"
          >
            <div className="text-2xl mb-2">🤖</div>
            <div className="text-sm font-medium text-gray-900">AI Sage</div>
          </Link>
        </div>

        {/* Analytics Dashboard */}
        <AnalyticsDashboard 
          timeRange={timeRange} 
          className="bg-gray-50 p-6 rounded-lg"
        />

        {/* Quick Insights */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <h3 className="text-lg font-medium text-gray-900 mb-4">💡 Quick Insights</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="p-4 bg-blue-50 rounded-lg">
              <h4 className="font-medium text-blue-900 mb-2">📖 Reading Habits</h4>
              <p className="text-sm text-blue-800">
                Track your daily reading patterns and discover your most productive reading hours.
              </p>
            </div>
            
            <div className="p-4 bg-green-50 rounded-lg">
              <h4 className="font-medium text-green-900 mb-2">🎯 Goals Progress</h4>
              <p className="text-sm text-green-800">
                Monitor your reading goals and stay motivated with progress tracking.
              </p>
            </div>
            
            <div className="p-4 bg-purple-50 rounded-lg">
              <h4 className="font-medium text-purple-900 mb-2">📊 Genre Analysis</h4>
              <p className="text-sm text-purple-800">
                Explore your reading preferences and discover new genres to try.
              </p>
            </div>
            
            <div className="p-4 bg-orange-50 rounded-lg">
              <h4 className="font-medium text-orange-900 mb-2">✍️ Engagement</h4>
              <p className="text-sm text-orange-800">
                See how actively you engage with texts through annotations and notes.
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center py-8 border-t">
          <p className="text-gray-600 text-sm">
            Reading analytics help you understand and improve your learning journey. 
            <Link href="/stats" className="text-indigo-600 hover:text-indigo-500 ml-1">
              View detailed statistics →
            </Link>
          </p>
        </div>
      </div>
    </Layout>
  );
}