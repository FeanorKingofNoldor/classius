'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import Link from 'next/link';

export default function Home() {
  const { isAuthenticated, checkAuth } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    checkAuth();
    if (isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, checkAuth, router]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col justify-center min-h-screen py-12">
          <div className="text-center">
            <div className="mb-8">
              <span className="text-8xl">📚</span>
            </div>
            <h1 className="text-6xl font-bold text-gray-900 mb-6">
              Classius
            </h1>
            <p className="text-xl text-gray-600 mb-12 max-w-2xl mx-auto">
              Study classical literature, philosophy, and historical texts with 
              AI-powered educational assistance. Discover the wisdom of the ages
              in the digital era.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
              <Link
                href="/auth/register"
                className="inline-flex items-center justify-center px-8 py-3 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 transition duration-200 shadow-lg hover:shadow-xl"
              >
                Start Your Journey
              </Link>
              <Link
                href="/auth/login"
                className="inline-flex items-center justify-center px-8 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition duration-200 shadow-lg hover:shadow-xl"
              >
                Sign In
              </Link>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-4xl mx-auto">
              <div className="text-center p-6 bg-white rounded-lg shadow-md">
                <div className="text-4xl mb-4">📖</div>
                <h3 className="text-lg font-semibold mb-2">Classical Library</h3>
                <p className="text-gray-600">Upload and organize your collection of classical texts, from Plato to Augustine.</p>
              </div>
              
              <div className="text-center p-6 bg-white rounded-lg shadow-md">
                <div className="text-4xl mb-4">🤖</div>
                <h3 className="text-lg font-semibold mb-2">AI Sage</h3>
                <p className="text-gray-600">Get educational insights and analysis powered by AI tuned for classical education.</p>
              </div>
              
              <div className="text-center p-6 bg-white rounded-lg shadow-md">
                <div className="text-4xl mb-4">✍️</div>
                <h3 className="text-lg font-semibold mb-2">Smart Annotations</h3>
                <p className="text-gray-600">Take notes, highlight passages, and sync your insights across all devices.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
