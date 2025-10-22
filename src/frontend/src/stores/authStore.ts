import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import Cookies from 'js-cookie';
import { User, authApi, LoginRequest, RegisterRequest } from '@/lib/api';
import { toast } from 'react-hot-toast';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (credentials: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => void;
  updateUser: (user: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,

      login: async (credentials: LoginRequest) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(credentials);
          const { user, access_token, refresh_token, expires_in } = response.data.data;

          // Store tokens in cookies
          Cookies.set('access_token', access_token, { 
            expires: new Date(Date.now() + expires_in * 1000),
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict'
          });
          Cookies.set('refresh_token', refresh_token, { 
            expires: 7, // 7 days
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict'
          });

          set({ 
            user, 
            isAuthenticated: true, 
            isLoading: false 
          });

          toast.success(`Welcome back, ${user.full_name || user.username}!`);
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      register: async (credentials: RegisterRequest) => {
        set({ isLoading: true });
        try {
          const response = await authApi.register(credentials);
          const { user, access_token, refresh_token, expires_in } = response.data.data;

          // Store tokens in cookies
          Cookies.set('access_token', access_token, { 
            expires: new Date(Date.now() + expires_in * 1000),
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict'
          });
          Cookies.set('refresh_token', refresh_token, { 
            expires: 7, // 7 days
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict'
          });

          set({ 
            user, 
            isAuthenticated: true, 
            isLoading: false 
          });

          toast.success(`Welcome to Classius, ${user.full_name || user.username}!`);
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: async () => {
        set({ isLoading: true });
        try {
          // Call logout endpoint (optional, for server-side cleanup)
          await authApi.logout();
        } catch (error) {
          // Continue with logout even if server call fails
          console.error('Logout error:', error);
        } finally {
          // Clear cookies
          Cookies.remove('access_token');
          Cookies.remove('refresh_token');

          // Clear state
          set({ 
            user: null, 
            isAuthenticated: false, 
            isLoading: false 
          });

          toast.success('Logged out successfully');
        }
      },

      checkAuth: () => {
        const token = Cookies.get('access_token');
        const { user } = get();
        
        if (token && user) {
          set({ isAuthenticated: true });
        } else {
          set({ 
            user: null, 
            isAuthenticated: false 
          });
          Cookies.remove('access_token');
          Cookies.remove('refresh_token');
        }
      },

      updateUser: (userData: Partial<User>) => {
        const { user } = get();
        if (user) {
          set({ user: { ...user, ...userData } });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// Initialize auth check on app start
if (typeof window !== 'undefined') {
  useAuthStore.getState().checkAuth();
}