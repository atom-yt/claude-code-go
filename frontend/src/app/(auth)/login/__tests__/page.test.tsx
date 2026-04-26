import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import LoginPage from '../page';

// Mock the auth store
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn(() => ({
    login: vi.fn(),
    isLoading: false,
  })),
}));

// Mock Next.js hooks
vi.mock('next/navigation', () => ({
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));

import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';

describe('LoginPage', () => {
  const mockPush = vi.fn();
  const mockLogin = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (useRouter as vi.Mock).mockReturnValue({ push: mockPush });
    (useAuthStore as vi.Mock).mockReturnValue({
      login: mockLogin,
      isLoading: false,
    });
  });

  it('should render login form with email and password fields', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');

    expect(emailInput).toBeInTheDocument();
    expect(passwordInput).toBeInTheDocument();
  });

  it('should render login button', () => {
    render(<LoginPage />);

    const loginButton = screen.getByRole('button', { name: /登录/ });
    expect(loginButton).toBeInTheDocument();
  });

  it('should render register link', () => {
    render(<LoginPage />);

    const registerLink = screen.getByText('注册');
    expect(registerLink).toBeInTheDocument();
    expect(registerLink).toHaveAttribute('href', '/register');
  });

  it('should show error message on login failure', async () => {
    const errorMessage = 'Invalid credentials';
    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn().mockRejectedValue(new Error(errorMessage)),
      isLoading: false,
    });

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const loginButton = screen.getByRole('button', { name: /登录/ });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('should call login with correct credentials on submit', async () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const loginButton = screen.getByRole('button', { name: /登录/ });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123');
    });
  });

  it('should navigate to home on successful login', async () => {
    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn().mockResolvedValue(undefined),
      isLoading: false,
    });

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const loginButton = screen.getByRole('button', { name: /登录/ });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/');
    });
  });

  it('should disable submit button when loading', () => {
    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn(),
      isLoading: true,
    });

    render(<LoginPage />);

    const loginButton = screen.getByRole('button', { name: /登录中.../ });
    expect(loginButton).toBeDisabled();
  });

  it('should show loading text when logging in', () => {
    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn(),
      isLoading: true,
    });

    render(<LoginPage />);

    const loginButton = screen.getByRole('button', { name: /登录中.../ });
    expect(loginButton).toHaveTextContent('登录中...');
  });

  it('should have email input with correct type', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    expect(emailInput).toHaveAttribute('type', 'email');
  });

  it('should have password input with correct type', () => {
    render(<LoginPage />);

    const passwordInput = screen.getByPlaceholderText('••••••••');
    expect(passwordInput).toHaveAttribute('type', 'password');
  });

  it('should require email and password fields', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');

    expect(emailInput).toBeRequired();
    expect(passwordInput).toBeRequired();
  });

  it('should clear error message after new submission', async () => {
    const errorMessage = 'Invalid credentials';
    let callCount = 0;

    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn().mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.reject(new Error(errorMessage));
        }
        return Promise.resolve();
      }),
      isLoading: false,
    });

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const loginButton = screen.getByRole('button', { name: /登录/ });

    // First attempt - should show error
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    // Second attempt - should clear error
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.queryByText(errorMessage)).not.toBeInTheDocument();
    });
  });

  it('should render logo and branding', () => {
    render(<LoginPage />);

    expect(screen.getByText('登录 atom')).toBeInTheDocument();
    expect(screen.getByText('小原子，大能量')).toBeInTheDocument();
  });

  it('should update email state on input change', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com') as HTMLInputElement;

    fireEvent.change(emailInput, { target: { value: 'new@example.com' } });

    expect(emailInput.value).toBe('new@example.com');
  });

  it('should update password state on input change', () => {
    render(<LoginPage />);

    const passwordInput = screen.getByPlaceholderText('••••••••') as HTMLInputElement;

    fireEvent.change(passwordInput, { target: { value: 'newpassword' } });

    expect(passwordInput.value).toBe('newpassword');
  });

  it('should submit form on Enter key press', async () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });

    // Simulate Enter key press on the form
    const form = emailInput.closest('form');
    if (form) {
      fireEvent.submit(form);
    }

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123');
    });
  });

  it('should display error with destructive styling', async () => {
    const errorMessage = 'Login failed';
    (useAuthStore as vi.Mock).mockReturnValue({
      login: vi.fn().mockRejectedValue(new Error(errorMessage)),
      isLoading: false,
    });

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const loginButton = screen.getByRole('button', { name: /登录/ });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      const errorElement = screen.getByText(errorMessage);
      expect(errorElement).toBeInTheDocument();
      expect(errorElement.className).toContain('text-destructive');
    });
  });
});