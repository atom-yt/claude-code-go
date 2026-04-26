import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useAuthStore } from '../authStore';
import { authApi } from '@/lib/api';
import type { User } from '@/types';

// Mock the authApi
vi.mock('@/lib/api', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    me: vi.fn(),
  },
}));

const mockUser: User = {
  id: 'user-1',
  email: 'test@example.com',
  username: 'testuser',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

const mockToken = 'test-token-123';

describe('authStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
    });
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('should login successfully and set user/token', async () => {
      vi.mocked(authApi.login).mockResolvedValue({
        user: mockUser,
        access_token: mockToken,
        refresh_token: 'refresh-token',
      });

      await useAuthStore.getState().login('test@example.com', 'password');

      expect(authApi.login).toHaveBeenCalledWith('test@example.com', 'password');
      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().token).toBe(mockToken);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().isLoading).toBe(false);
      expect(localStorage.setItem).toHaveBeenCalledWith('token', mockToken);
      expect(localStorage.setItem).toHaveBeenCalledWith('user', JSON.stringify(mockUser));
    });

    it('should handle login errors', async () => {
      const error = new Error('Invalid credentials');
      vi.mocked(authApi.login).mockRejectedValue(error);

      await expect(
        useAuthStore.getState().login('test@example.com', 'wrong-password')
      ).rejects.toThrow('Invalid credentials');

      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(useAuthStore.getState().isLoading).toBe(false);
    });
  });

  describe('register', () => {
    it('should register successfully and set user/token', async () => {
      vi.mocked(authApi.register).mockResolvedValue({
        user: mockUser,
        access_token: mockToken,
        refresh_token: 'refresh-token',
      });

      await useAuthStore.getState().register('test@example.com', 'password', 'Test User');

      expect(authApi.register).toHaveBeenCalledWith('test@example.com', 'password', 'Test User');
      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().token).toBe(mockToken);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().isLoading).toBe(false);
    });

    it('should handle registration errors', async () => {
      const error = new Error('Email already exists');
      vi.mocked(authApi.register).mockRejectedValue(error);

      await expect(
        useAuthStore.getState().register('test@example.com', 'password', 'Test User')
      ).rejects.toThrow('Email already exists');

      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe('logout', () => {
    it('should logout and clear user/token', async () => {
      // First, login
      vi.mocked(authApi.login).mockResolvedValue({
        user: mockUser,
        access_token: mockToken,
        refresh_token: 'refresh-token',
      });
      await useAuthStore.getState().login('test@example.com', 'password');

      // Then logout
      vi.mocked(authApi.logout).mockResolvedValue(undefined);
      await useAuthStore.getState().logout();

      expect(authApi.logout).toHaveBeenCalled();
      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
      expect(localStorage.removeItem).toHaveBeenCalledWith('user');
    });

    it('should logout even if API call fails', async () => {
      // Set initial state as logged in
      useAuthStore.setState({
        user: mockUser,
        token: mockToken,
        isAuthenticated: true,
      });

      vi.mocked(authApi.logout).mockRejectedValue(new Error('Network error'));
      await useAuthStore.getState().logout();

      // State should still be cleared
      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe('checkAuth', () => {
    it('should check auth and set user if authenticated', async () => {
      vi.mocked(authApi.me).mockResolvedValue({ user: mockUser });

      await useAuthStore.getState().checkAuth();

      expect(authApi.me).toHaveBeenCalled();
      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().isLoading).toBe(false);
    });

    it('should clear auth state if not authenticated', async () => {
      // Set initial state as logged in
      useAuthStore.setState({
        user: mockUser,
        token: mockToken,
        isAuthenticated: true,
      });

      vi.mocked(authApi.me).mockRejectedValue(new Error('Unauthorized'));
      await useAuthStore.getState().checkAuth();

      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(useAuthStore.getState().isLoading).toBe(false);
    });
  });

  describe('token persistence', () => {
    it('should save token to localStorage on login', async () => {
      vi.mocked(authApi.login).mockResolvedValue({
        user: mockUser,
        access_token: mockToken,
        refresh_token: 'refresh-token',
      });

      await useAuthStore.getState().login('test@example.com', 'password');

      expect(localStorage.setItem).toHaveBeenCalledWith('token', mockToken);
    });

    it('should remove token from localStorage on logout', async () => {
      useAuthStore.setState({
        user: mockUser,
        token: mockToken,
        isAuthenticated: true,
      });

      vi.mocked(authApi.logout).mockResolvedValue(undefined);
      await useAuthStore.getState().logout();

      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
    });
  });

  describe('401 auto-logout handling', () => {
    it('should handle 401 error by clearing auth state', async () => {
      useAuthStore.setState({
        user: mockUser,
        token: mockToken,
        isAuthenticated: true,
      });

      // Mock a 401 response from authApi.me
      const error = new Error('Unauthorized') as any;
      error.response = { status: 401 };
      vi.mocked(authApi.me).mockRejectedValue(error);

      await useAuthStore.getState().checkAuth();

      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().token).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe('setUser and setToken', () => {
    it('should set user and update localStorage', () => {
      useAuthStore.getState().setUser(mockUser);

      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(localStorage.setItem).toHaveBeenCalledWith('user', JSON.stringify(mockUser));
    });

    it('should clear user and remove from localStorage', () => {
      useAuthStore.setState({ user: mockUser });
      useAuthStore.getState().setUser(null);

      expect(useAuthStore.getState().user).toBe(null);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(localStorage.removeItem).toHaveBeenCalledWith('user');
    });

    it('should set token and update localStorage', () => {
      useAuthStore.getState().setToken(mockToken);

      expect(useAuthStore.getState().token).toBe(mockToken);
      expect(localStorage.setItem).toHaveBeenCalledWith('token', mockToken);
    });

    it('should clear token and remove from localStorage', () => {
      useAuthStore.setState({ token: mockToken });
      useAuthStore.getState().setToken(null);

      expect(useAuthStore.getState().token).toBe(null);
      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
    });
  });
});