import { create } from 'zustand';
import { AuthState, User } from '@/types';
import { authApi } from '@/lib/api';

interface AuthStore extends AuthState {
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  _initialize: () => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthStore>((set, get) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: false,

  // Initialize from localStorage
  _initialize: () => {
    if (typeof window !== 'undefined') {
      const savedToken = localStorage.getItem('token');
      const savedUser = localStorage.getItem('user');
      if (savedToken && savedUser) {
        set({
          token: savedToken,
          user: JSON.parse(savedUser),
          isAuthenticated: true,
        });
      }
    }
  },

  setUser: (user) => {
    set({ user, isAuthenticated: !!user });
    if (!user) {
      localStorage.removeItem('user');
    } else {
      localStorage.setItem('user', JSON.stringify(user));
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
      const token = response.access_token;
      set({
        user: response.user,
        token: token,
        isAuthenticated: true,
        isLoading: false,
      });
      if (token) {
        localStorage.setItem('token', token);
      }
      if (response.user) {
        localStorage.setItem('user', JSON.stringify(response.user));
      }
    } catch (error: any) {
      set({ isLoading: false });
      throw error;
    }
  },

  register: async (email, password, name) => {
    set({ isLoading: true });
    try {
      const response = await authApi.register(email, password, name);
      const token = response.access_token;
      set({
        user: response.user,
        token: token,
        isAuthenticated: true,
        isLoading: false,
      });
      if (token) {
        localStorage.setItem('token', token);
      }
      if (response.user) {
        localStorage.setItem('user', JSON.stringify(response.user));
      }
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
      localStorage.removeItem('user');
    }
  },

  checkAuth: async () => {
    set({ isLoading: true });
    try {
      const response = await authApi.me();
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
      if (response.user) {
        localStorage.setItem('user', JSON.stringify(response.user));
      }
    } catch (error) {
      set({ user: null, token: null, isAuthenticated: false, isLoading: false });
    }
  },
}));

// Initialize store on client side
if (typeof window !== 'undefined') {
  useAuthStore.getState()._initialize();
}