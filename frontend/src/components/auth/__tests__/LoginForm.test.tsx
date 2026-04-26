import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import LoginPage from '@/app/(auth)/login/page';

// Mock the auth store
vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn(),
}));

// Mock Next.js hooks
vi.mock('next/navigation', () => ({
  useRouter: vi.fn(),
}));

import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';

describe('LoginForm', () => {
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

  it('should render email and password input fields', () => {
    render(<LoginPage />);

    expect(screen.getByPlaceholderText('your@email.com')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('••••••••')).toBeInTheDocument();
  });

  it('should have email input with type="email"', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    expect(emailInput).toHaveAttribute('type', 'email');
  });

  it('should have password input with type="password"', () => {
    render(<LoginPage />);

    const passwordInput = screen.getByPlaceholderText('••••••••');
    expect(passwordInput).toHaveAttribute('type', 'password');
  });

  it('should mark email and password as required fields', () => {
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');

    expect(emailInput).toBeRequired();
    expect(passwordInput).toBeRequired();
  });

  it('should update email value on input change', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    await user.type(emailInput, 'test@example.com');

    expect(emailInput).toHaveValue('test@example.com');
  });

  it('should update password value on input change', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const passwordInput = screen.getByPlaceholderText('••••••••');
    await user.type(passwordInput, 'password123');

    expect(passwordInput).toHaveValue('password123');
  });

  it('should submit form with email and password on button click', async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValue(undefined);

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'password123');
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123');
    });
  });

  it('should not call login when email is empty', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(passwordInput, 'password123');
    await user.click(submitButton);

    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('should not call login when password is empty', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.click(submitButton);

    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('should not call login when both fields are empty', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const submitButton = screen.getByRole('button', { name: /登录/ });
    await user.click(submitButton);

    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('should display error message when login fails', async () => {
    const user = userEvent.setup();
    const errorMessage = 'Invalid credentials';
    mockLogin.mockRejectedValue(new Error(errorMessage));

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'wrongpassword');
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('should display error message with destructive styling', async () => {
    const user = userEvent.setup();
    const errorMessage = 'Login failed';
    mockLogin.mockRejectedValue(new Error(errorMessage));

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'password123');
    await user.click(submitButton);

    await waitFor(() => {
      const errorElement = screen.getByText(errorMessage);
      expect(errorElement).toBeInTheDocument();
      expect(errorElement.className).toContain('text-destructive');
    });
  });

  it('should clear error message before new submission attempt', async () => {
    const user = userEvent.setup();
    let callCount = 0;

    mockLogin.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return Promise.reject(new Error('First error'));
      }
      return Promise.resolve();
    });

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'wrongpassword');
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('First error')).toBeInTheDocument();
    });

    await user.clear(passwordInput);
    await user.type(passwordInput, 'correctpassword');
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.queryByText('First error')).not.toBeInTheDocument();
    });
  });

  it('should display default error message when error has no message property', async () => {
    const user = userEvent.setup();
    mockLogin.mockRejectedValue({});

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'password123');
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('登录失败')).toBeInTheDocument();
    });
  });

  it('should submit form on Enter key press', async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValue(undefined);

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com');
    const passwordInput = screen.getByPlaceholderText('••••••••');

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'password123');

    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123');
    });
  });

  it('should render email and password labels', () => {
    render(<LoginPage />);

    expect(screen.getByText('邮箱')).toBeInTheDocument();
    expect(screen.getByText('密码')).toBeInTheDocument();
  });

  it('should maintain input values after failed login', async () => {
    const user = userEvent.setup();
    mockLogin.mockRejectedValue(new Error('Invalid'));

    render(<LoginPage />);

    const emailInput = screen.getByPlaceholderText('your@email.com') as HTMLInputElement;
    const passwordInput = screen.getByPlaceholderText('••••••••') as HTMLInputElement;
    const submitButton = screen.getByRole('button', { name: /登录/ });

    await user.type(emailInput, 'test@example.com');
    await user.type(passwordInput, 'password123');
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Invalid')).toBeInTheDocument();
    });

    expect(emailInput.value).toBe('test@example.com');
    expect(passwordInput.value).toBe('password123');
  });
});
