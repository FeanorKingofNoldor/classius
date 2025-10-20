'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import Layout from '@/components/Layout';
import { 
  BookOpenIcon, 
  ChatBubbleLeftRightIcon, 
  PlusIcon,
  ChartBarIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';

export default function DashboardPage() {
  const { isAuthenticated, checkAuth, user } = useAuthStore();
  const router = useRouter();

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

  const quickActions = [
    {
      name: 'Upload a Book',
      description: 'Add a new classical text to your library',
      icon: PlusIcon,
      href: '/library/upload',
      color: 'bg-blue-500',
    },
    {
      name: 'Ask the Sage',
      description: 'Get AI-powered insights on your readings',
      icon: ChatBubbleLeftRightIcon,
      href: '/sage',
      color: 'bg-purple-500',
    },
    {
      name: 'Continue Reading',
      description: 'Pick up where you left off',
      icon: BookOpenIcon,
      href: '/library',
      color: 'bg-green-500',
    },
  ];

  const recentActivity = [
    {
      id: 1,
      type: 'annotation',
      title: 'Added note to "The Republic"',
      time: '2 hours ago',
      icon: '📝',
    },
    {
      id: 2,
      type: 'sage',
      title: 'Asked Sage about Stoic philosophy',
      time: '1 day ago',
      icon: '🤖',
    },
    {
      id: 3,
      type: 'book',
      title: 'Started reading "Meditations"',
      time: '2 days ago',
      icon: '📖',
    },
  ];

  const stats = [
    { name: 'Books in Library', value: '12', icon: BookOpenIcon },
    { name: 'Pages Read', value: '1,247', icon: ChartBarIcon },
    { name: 'Notes Created', value: '89', icon: '📝' },
    { name: 'Hours Reading', value: '43', icon: ClockIcon },
  ];

  return (
    <Layout title="Dashboard">
      <div className="space-y-8">
        {/* Welcome Section */}
        <div className="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-lg px-6 py-8 text-white">
          <h1 className="text-3xl font-bold mb-2">
            Welcome back, {user?.full_name || user?.username}!
          </h1>
          <p className="text-indigo-100">
            Continue your journey through the great works of classical literature and philosophy.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {stats.map((stat) => (
            <div key={stat.name} className="bg-white rounded-lg p-6 shadow-sm border border-gray-200">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  {typeof stat.icon === 'string' ? (
                    <span className="text-2xl">{stat.icon}</span>
                  ) : (
                    <stat.icon className="h-8 w-8 text-indigo-600" />
                  )}
                </div>
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-500">{stat.name}</p>
                  <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Quick Actions */}
        <div>
          <h2 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {quickActions.map((action) => (
              <a
                key={action.name}
                href={action.href}
                className="relative group bg-white p-6 rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow duration-200"
              >
                <div>
                  <span className={`rounded-lg inline-flex p-3 ${action.color} text-white`}>
                    <action.icon className="h-6 w-6" aria-hidden="true" />
                  </span>
                </div>
                <div className="mt-4">
                  <h3 className="text-lg font-medium text-gray-900 group-hover:text-indigo-600">
                    {action.name}
                  </h3>
                  <p className="mt-2 text-sm text-gray-500">
                    {action.description}
                  </p>
                </div>
              </a>
            ))}
          </div>
        </div>

        {/* Recent Activity & Reading Progress */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Recent Activity */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Recent Activity</h3>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {recentActivity.map((activity) => (
                  <div key={activity.id} className="flex items-start space-x-3">
                    <span className="text-xl">{activity.icon}</span>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900">
                        {activity.title}
                      </p>
                      <p className="text-sm text-gray-500">{activity.time}</p>
                    </div>
                  </div>
                ))}
              </div>
              <div className="mt-6">
                <a
                  href="#"
                  className="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                >
                  View all activity →
                </a>
              </div>
            </div>
          </div>

          {/* Current Reading */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Currently Reading</h3>
            </div>
            <div className="p-6">
              <div className="text-center py-8">
                <BookOpenIcon className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No active reading</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Start reading a book to see your progress here.
                </p>
                <div className="mt-6">
                  <a
                    href="/library"
                    className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                  >
                    Browse Library
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Recommended Reading */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Recommended Classical Texts</h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {[
                { title: 'The Republic', author: 'Plato', description: 'Explore justice and the ideal state' },
                { title: 'Meditations', author: 'Marcus Aurelius', description: 'Stoic philosophy and self-reflection' },
                { title: 'The Aeneid', author: 'Virgil', description: 'Epic tale of destiny and duty' },
              ].map((book, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900">{book.title}</h4>
                  <p className="text-sm text-gray-600">by {book.author}</p>
                  <p className="text-sm text-gray-500 mt-2">{book.description}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}