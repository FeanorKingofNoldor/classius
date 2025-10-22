'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter, useParams } from 'next/navigation';
import { BookReader } from '@/components/readers';
import { Book, booksApi } from '@/lib/api';
import { toast } from 'react-hot-toast';


export default function ReadingPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  const params = useParams();
  const bookId = params.bookId as string;

  const [book, setBook] = useState<Book | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    if (bookId) {
      fetchBook();
    }
  }, [isAuthenticated, router, bookId]);

  const fetchBook = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await booksApi.getBook(bookId);
      setBook(response.data.data);
    } catch (err) {
      console.error('Error fetching book:', err);
      setError('Failed to load book');
      toast.error('Failed to load book');
    } finally {
      setLoading(false);
    }
  };

  const handleCloseReader = () => {
    router.push(`/books/${bookId}`);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading book...</p>
        </div>
      </div>
    );
  }

  if (error || !book) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">📖</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {error || 'Book not found'}
          </h3>
          <button
            onClick={() => router.push('/books')}
            className="text-indigo-600 hover:text-indigo-800 font-medium"
          >
            ← Back to Library
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <BookReader
        book={book}
        fullscreen={true}
        onClose={handleCloseReader}
      />
    </div>
  );
}