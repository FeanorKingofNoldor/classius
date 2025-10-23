'use client';

import { useState, useRef, useCallback, useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface BookUploadForm {
  title: string;
  author: string;
  language: string;
  genre: string;
  publisher: string;
  isbn: string;
  description: string;
  published_at: string;
  tags: string;
  is_public: boolean;
}

export default function BookUploadPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const [dragActive, setDragActive] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [formData, setFormData] = useState<BookUploadForm>({
    title: '',
    author: '',
    language: 'en',
    genre: '',
    publisher: '',
    isbn: '',
    description: '',
    published_at: '',
    tags: '',
    is_public: false,
  });

  const supportedTypes = ['application/pdf', 'application/epub+zip', 'text/plain', 'application/x-mobipocket-ebook'];
  const maxFileSize = 100 * 1024 * 1024; // 100MB

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
    }
  }, [isAuthenticated, router]);

  // Handle drag events
  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  }, []);

  // Handle drop
  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const droppedFile = e.dataTransfer.files[0];
      handleFileSelection(droppedFile);
    }
  }, []);

  // Handle file input change
  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFileSelection(e.target.files[0]);
    }
  };

  // Process selected file
  const handleFileSelection = (selectedFile: File) => {
    // Validate file type
    if (!supportedTypes.includes(selectedFile.type)) {
      toast.error('Unsupported file type. Please upload PDF, EPUB, TXT, or MOBI files.');
      return;
    }

    // Validate file size
    if (selectedFile.size > maxFileSize) {
      toast.error('File size too large. Maximum size is 100MB.');
      return;
    }

    setFile(selectedFile);

    // Try to extract metadata from filename
    const fileName = selectedFile.name.replace(/\.[^/.]+$/, ""); // Remove extension
    const parts = fileName.split(' - ');
    
    if (parts.length >= 2) {
      setFormData(prev => ({
        ...prev,
        title: parts[1] || fileName,
        author: parts[0] || '',
      }));
    } else {
      setFormData(prev => ({
        ...prev,
        title: fileName,
      }));
    }

    toast.success('File selected successfully!');
  };

  // Handle form input changes
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value
    }));
  };

  // Upload the book
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!file) {
      toast.error('Please select a file to upload.');
      return;
    }

    if (!formData.title.trim()) {
      toast.error('Title is required.');
      return;
    }

    if (!formData.author.trim()) {
      toast.error('Author is required.');
      return;
    }

    try {
      setUploading(true);
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('No authentication token');
      }

      // Create FormData for multipart upload
      const uploadData = new FormData();
      uploadData.append('file', file);
      uploadData.append('title', formData.title.trim());
      uploadData.append('author', formData.author.trim());
      uploadData.append('language', formData.language);
      uploadData.append('genre', formData.genre.trim());
      uploadData.append('publisher', formData.publisher.trim());
      uploadData.append('isbn', formData.isbn.trim());
      uploadData.append('description', formData.description.trim());
      uploadData.append('is_public', formData.is_public.toString());

      if (formData.published_at) {
        uploadData.append('published_at', formData.published_at);
      }

      if (formData.tags.trim()) {
        uploadData.append('tags', formData.tags.trim());
      }

      const response = await fetch('http://localhost:8081/api/v1/books/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: uploadData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to upload book');
      }

      const result = await response.json();
      toast.success('Book uploaded successfully!');
      
      // Redirect to books page or the new book
      router.push('/books');
      
    } catch (err) {
      console.error('Upload error:', err);
      toast.error(err instanceof Error ? err.message : 'Failed to upload book');
    } finally {
      setUploading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Link href="/books" className="text-indigo-600 hover:text-indigo-800">
                ← Back to Library
              </Link>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Welcome, {user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Upload a Book</h1>
          <p className="mt-2 text-gray-600">
            Add a new book to your classical library. Supported formats: PDF, EPUB, TXT, MOBI (up to 100MB)
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-8">
          {/* File Upload Section */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">Select File</h2>
            
            {!file ? (
              <div
                className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                  dragActive
                    ? 'border-indigo-400 bg-indigo-50'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
                onDragEnter={handleDrag}
                onDragLeave={handleDrag}
                onDragOver={handleDrag}
                onDrop={handleDrop}
              >
                <div className="text-6xl text-gray-400 mb-4">📚</div>
                <div className="space-y-2">
                  <p className="text-xl text-gray-600">
                    Drag and drop your book file here, or{' '}
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="text-indigo-600 hover:text-indigo-500 underline"
                    >
                      browse files
                    </button>
                  </p>
                  <p className="text-sm text-gray-500">
                    Supports PDF, EPUB, TXT, MOBI files up to 100MB
                  </p>
                </div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,.epub,.txt,.mobi"
                  onChange={handleFileInputChange}
                  className="hidden"
                />
              </div>
            ) : (
              <div className="border rounded-lg p-4 bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="text-2xl">📖</div>
                    <div>
                      <p className="font-medium text-gray-900">{file.name}</p>
                      <p className="text-sm text-gray-500">{formatFileSize(file.size)}</p>
                    </div>
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      setFile(null);
                      setFormData(prev => ({ ...prev, title: '', author: '' }));
                    }}
                    className="text-red-600 hover:text-red-800"
                  >
                    Remove
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Book Metadata Section */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-6">Book Details</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Title - Required */}
              <div className="md:col-span-2">
                <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
                  Title *
                </label>
                <input
                  type="text"
                  id="title"
                  name="title"
                  required
                  value={formData.title}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Enter the book title"
                />
              </div>

              {/* Author - Required */}
              <div>
                <label htmlFor="author" className="block text-sm font-medium text-gray-700 mb-2">
                  Author *
                </label>
                <input
                  type="text"
                  id="author"
                  name="author"
                  required
                  value={formData.author}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Enter the author name"
                />
              </div>

              {/* Language */}
              <div>
                <label htmlFor="language" className="block text-sm font-medium text-gray-700 mb-2">
                  Language
                </label>
                <select
                  id="language"
                  name="language"
                  value={formData.language}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option value="en">English</option>
                  <option value="la">Latin</option>
                  <option value="grc">Ancient Greek</option>
                  <option value="de">German</option>
                  <option value="fr">French</option>
                  <option value="it">Italian</option>
                  <option value="es">Spanish</option>
                  <option value="ru">Russian</option>
                  <option value="zh">Chinese</option>
                  <option value="ar">Arabic</option>
                  <option value="he">Hebrew</option>
                  <option value="sa">Sanskrit</option>
                </select>
              </div>

              {/* Genre */}
              <div>
                <label htmlFor="genre" className="block text-sm font-medium text-gray-700 mb-2">
                  Genre
                </label>
                <input
                  type="text"
                  id="genre"
                  name="genre"
                  value={formData.genre}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Philosophy, Literature, History, etc."
                />
              </div>

              {/* Publisher */}
              <div>
                <label htmlFor="publisher" className="block text-sm font-medium text-gray-700 mb-2">
                  Publisher
                </label>
                <input
                  type="text"
                  id="publisher"
                  name="publisher"
                  value={formData.publisher}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Publisher name"
                />
              </div>

              {/* ISBN */}
              <div>
                <label htmlFor="isbn" className="block text-sm font-medium text-gray-700 mb-2">
                  ISBN
                </label>
                <input
                  type="text"
                  id="isbn"
                  name="isbn"
                  value={formData.isbn}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="978-0-123456-78-9"
                />
              </div>

              {/* Published Date */}
              <div>
                <label htmlFor="published_at" className="block text-sm font-medium text-gray-700 mb-2">
                  Published Date
                </label>
                <input
                  type="date"
                  id="published_at"
                  name="published_at"
                  value={formData.published_at}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>

              {/* Tags */}
              <div className="md:col-span-2">
                <label htmlFor="tags" className="block text-sm font-medium text-gray-700 mb-2">
                  Tags
                </label>
                <input
                  type="text"
                  id="tags"
                  name="tags"
                  value={formData.tags}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Stoicism, Ethics, Roman, Ancient (separate with commas)"
                />
                <p className="mt-1 text-sm text-gray-500">
                  Separate multiple tags with commas
                </p>
              </div>

              {/* Description */}
              <div className="md:col-span-2">
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                  Description
                </label>
                <textarea
                  id="description"
                  name="description"
                  rows={4}
                  value={formData.description}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Brief description of the book's content and significance..."
                />
              </div>

              {/* Public Visibility */}
              <div className="md:col-span-2">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="is_public"
                    name="is_public"
                    checked={formData.is_public}
                    onChange={handleInputChange}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <label htmlFor="is_public" className="ml-2 block text-sm text-gray-700">
                    Make this book publicly visible to other users
                  </label>
                </div>
                <p className="mt-1 text-sm text-gray-500">
                  Public books can be discovered and accessed by other Classius users
                </p>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end space-x-4">
            <Link
              href="/books"
              className="px-6 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={!file || uploading}
              className="px-6 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
            >
              {uploading && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              )}
              {uploading ? 'Uploading...' : 'Upload Book'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}