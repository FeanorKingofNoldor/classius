import axios, { AxiosInstance, AxiosResponse } from 'axios';
import Cookies from 'js-cookie';
import { toast } from 'react-hot-toast';

// API Base Configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = Cookies.get('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Handle 401 errors (unauthorized)
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Try to refresh token
        const refreshToken = Cookies.get('refresh_token');
        if (refreshToken) {
          const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token } = response.data.data;
          Cookies.set('access_token', access_token);

          // Retry original request with new token
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        Cookies.remove('access_token');
        Cookies.remove('refresh_token');
        window.location.href = '/auth/login';
      }
    }

    // Handle other errors
    if (error.response?.data?.message) {
      toast.error(error.response.data.message);
    } else if (error.message) {
      toast.error(error.message);
    }

    return Promise.reject(error);
  }
);

// API Response Types
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data: T;
  error?: string;
}

export interface PaginatedResponse<T = any> {
  data: T[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

// Auth Types
export interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
  avatar_url?: string;
  subscription_tier: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  full_name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

// Book Types
export interface Book {
  id: string;
  user_id: string;
  title: string;
  author: string;
  language: string;
  genre?: string;
  publisher?: string;
  published_at?: string;
  isbn?: string;
  description?: string;
  cover_url?: string;
  file_path: string;
  file_size: number;
  file_type: string;
  page_count?: number;
  word_count?: number;
  status: 'active' | 'archived' | 'processing' | 'error';
  is_public: boolean;
  tags?: Tag[];
  metadata: BookMetadata;
  created_at: string;
  updated_at: string;
}

export interface BookMetadata {
  original_file_name?: string;
  mime_type?: string;
  encoding?: string;
  has_images: boolean;
  has_toc: boolean;
  chapter_count?: number;
}

export interface Tag {
  id: string;
  user_id: string;
  name: string;
  color?: string;
  created_at: string;
  updated_at: string;
}

export interface BookUploadRequest {
  title: string;
  author: string;
  language?: string;
  genre?: string;
  publisher?: string;
  published_at?: string;
  isbn?: string;
  description?: string;
  tags?: string[];
  is_public?: boolean;
}

// Annotation Types
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

// Sage Types
export interface SageRequest {
  question: string;
  book_title?: string;
  book_author?: string;
  book_id?: string;
  passage_text?: string;
  annotation_id?: string;
  context?: string;
}

export interface SageResponse {
  answer: string;
  response_time: string;
  model: string;
  provider: string;
  tokens_used?: number;
  conversation_id?: string;
  sources?: string[];
  confidence?: number;
}

// API Functions

// Auth API
export const authApi = {
  register: (data: RegisterRequest) => 
    api.post<ApiResponse<AuthResponse>>('/auth/register', data),
  
  login: (data: LoginRequest) => 
    api.post<ApiResponse<AuthResponse>>('/auth/login', data),
  
  logout: () => 
    api.post<ApiResponse>('/auth/logout'),
  
  refreshToken: (refreshToken: string) =>
    api.post<ApiResponse<AuthResponse>>('/auth/refresh', { refresh_token: refreshToken }),
};

// User API
export const userApi = {
  getProfile: () => 
    api.get<ApiResponse<User>>('/user/profile'),
  
  updateProfile: (data: Partial<User>) => 
    api.put<ApiResponse<User>>('/user/profile', data),
  
  getProgress: () => 
    api.get<ApiResponse>('/user/progress'),
};

// Books API
export const booksApi = {
  getBooks: (params?: Record<string, any>) => 
    api.get<ApiResponse<PaginatedResponse<Book>>>('/books', { params }),
  
  getBook: (id: string) => 
    api.get<ApiResponse<Book>>(`/books/${id}`),
  
  uploadBook: (formData: FormData) => 
    api.post<ApiResponse<Book>>('/books/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    }),
  
  updateBook: (id: string, data: Partial<Book>) => 
    api.put<ApiResponse<Book>>(`/books/${id}`, data),
  
  deleteBook: (id: string) => 
    api.delete<ApiResponse>(`/books/${id}`),
  
  downloadBook: (id: string) => 
    api.get(`/books/${id}/download`, { responseType: 'blob' }),
  
  getBookContent: (id: string) => 
    api.get(`/books/${id}/content`, { responseType: 'blob' }),
  
  getStats: () => 
    api.get<ApiResponse>('/books/stats'),
  
  getTags: () => 
    api.get<ApiResponse<Tag[]>>('/books/tags'),
  
  createTag: (data: { name: string; color?: string }) => 
    api.post<ApiResponse<Tag>>('/books/tags', data),
  
  deleteTag: (id: string) => 
    api.delete<ApiResponse>(`/books/tags/${id}`),
};

// Annotations API
export const annotationsApi = {
  getAnnotations: (params?: Record<string, any>) => 
    api.get<ApiResponse<Annotation[]>>('/annotations', { params }),
  
  createAnnotation: (data: Partial<Annotation>) => 
    api.post<ApiResponse<Annotation>>('/annotations', data),
  
  updateAnnotation: (id: string, data: Partial<Annotation>) => 
    api.put<ApiResponse<Annotation>>(`/annotations/${id}`, data),
  
  deleteAnnotation: (id: string) => 
    api.delete<ApiResponse>(`/annotations/${id}`),
  
  syncAnnotations: (data: Annotation[]) => 
    api.post<ApiResponse<Annotation[]>>('/annotations/sync', { annotations: data }),
};

// Sage API
export const sageApi = {
  ask: (data: SageRequest) => 
    api.post<ApiResponse<SageResponse>>('/sage/ask', data),
  
  getCapabilities: () => 
    api.get<ApiResponse>('/sage/capabilities'),
  
  getHealth: () => 
    api.get<ApiResponse>('/sage/health'),
  
  getConversations: (params?: Record<string, any>) => 
    api.get<ApiResponse>('/sage/conversations', { params }),
  
  getConversation: (id: string) => 
    api.get<ApiResponse>(`/sage/conversations/${id}`),
  
  deleteConversation: (id: string) => 
    api.delete<ApiResponse>(`/sage/conversations/${id}`),
  
  getStats: () => 
    api.get<ApiResponse>('/sage/stats'),
  
  exportData: (format: 'json' | 'csv' | 'txt' = 'json') => 
    api.get<ApiResponse>(`/sage/export?format=${format}`),
};

export default api;