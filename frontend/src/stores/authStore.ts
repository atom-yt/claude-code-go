import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { AuthState, User } from '@/types';
import { authApi } from '@/lib/api';

interface AuthStore extends AuthState {
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,

      setUser: (user) => {
        set({ user, isAuthenticated: !!user });
        if (!user) {
          localStorage.removeItem('token');
        }
      },

      setToken: (token) => {
        set({ token });
        if (token) {
          localStorage.setItem('token', token);
        } else {
          localStorage.removeItem('token');
        }
      },

      login: async (email, password) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login(email, password);
          const tokenStr = response.token?.accessToken || (response.token as any);
          set({
            user: response.user,
            token: typeof tokenStr === 'string' ? tokenStr : JSON.stringify(tokenStr),
            isAuthenticated: true,
            isLoading: false,
          });
          localStorage.setItem('token', typeof tokenStr === 'string' ? tokenStr : JSON.stringify(tokenStr));
        } catch (error: any) {
          set({ isLoading: false });
          throw error;
        }
      },

      register: async (email, password, name) => {
        set({ isLoading: true });
        try {
          const response = await authApi.register(email, password, name);
          const tokenStr = response.token?.accessToken || (response.token as any);
          set({
            user: response.user,
            token: typeof tokenStr === 'string' ? tokenStr : JSON.stringify(tokenStr),
            isAuthenticated: true,
            isLoading: false,
          });
          localStorage.setItem('token', typeof tokenStr === 'string' ? tokenStr : JSON.stringify(tokenStr));
        } catch (error: any) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: async () => {
        try {
          await authApi.logout();
        } catch (error) {
          // Ignore logout errors
        } finally {
          set({
            user: null,
            token: null,
            isAuthenticated: false,
          });
          localStorage.removeItem('token');
        }
      },

      checkAuth: async () => {
        const token = localStorage.getItem('token');
        if (!token) {
          set({ isAuthenticated: false, user: null });
          return;
        }

        set({ isLoading: true });
        try {
          const response = await authApi.me();
          const tokenStr = response.token?.accessToken || (response.token as any);
          set({
            user: response.user,
            token: typeof tokenStr === 'string' ? tokenStr : JSON.stringify(tokenStr),
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
          });
          localStorage.removeItem('token');
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);