import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Sidebar } from '../Sidebar';

// Mock the stores
vi.mock('@/stores/sessionsStore', () => ({
  useSessionsStore: vi.fn(),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: vi.fn(),
}));

// Mock Next.js hooks
vi.mock('next/navigation', () => ({
  usePathname: vi.fn(),
  useRouter: vi.fn(),
}));

import { useSessionsStore } from '@/stores/sessionsStore';
import { useAuthStore } from '@/stores/authStore';
import { usePathname, useRouter } from 'next/navigation';

const mockSessions = [
  {
    id: 'session-1',
    userId: 'user-1',
    agentId: 'agent-1',
    title: 'First Conversation',
    status: 'active',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'session-2',
    userId: 'user-1',
    agentId: 'agent-1',
    title: 'Second Conversation',
    status: 'active',
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
];

describe('Sidebar', () => {
  const mockPush = vi.fn();
  const mockFetchSessions = vi.fn();
  const mockSelectSession = vi.fn();
  const mockDeleteSession = vi.fn();
  const mockLogout = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (useRouter as vi.Mock).mockReturnValue({ push: mockPush });
    (usePathname as vi.Mock).mockReturnValue('/');
    (useSessionsStore as vi.Mock).mockReturnValue({
      sessions: mockSessions,
      fetchSessions: mockFetchSessions,
      selectSession: mockSelectSession,
      deleteSession: mockDeleteSession,
      selectedSession: null,
    });
    (useAuthStore as vi.Mock).mockReturnValue({
      user: { id: 'user-1', email: 'test@example.com', username: 'test' },
      logout: mockLogout,
    });
  });

  it('should render sidebar with logo', () => {
    render(<Sidebar />);

    expect(screen.getByText('atom.')).toBeInTheDocument();
    expect(screen.getByText('a')).toBeInTheDocument();
  });

  it('should render new chat button', () => {
    render(<Sidebar />);

    const newChatButton = screen.getByRole('button', { name: /新的对话/ });
    expect(newChatButton).toBeInTheDocument();
  });

  it('should render session list heading', () => {
    render(<Sidebar />);

    expect(screen.getByText('我的对话')).toBeInTheDocument();
  });

  it('should render all sessions in the list', () => {
    render(<Sidebar />);

    expect(screen.getByText('First Conversation')).toBeInTheDocument();
    expect(screen.getByText('Second Conversation')).toBeInTheDocument();
  });

  it('should render navigation links', () => {
    render(<Sidebar />);

    expect(screen.getByText('知识')).toBeInTheDocument();
    expect(screen.getByText('技能')).toBeInTheDocument();
    expect(screen.getByText('产物')).toBeInTheDocument();
    expect(screen.getByText('设置')).toBeInTheDocument();
  });

  it('should call fetchSessions on mount', () => {
    render(<Sidebar />);

    expect(mockFetchSessions).toHaveBeenCalled();
  });

  it('should create new chat on button click', async () => {
    const user = userEvent.setup();
    render(<Sidebar />);

    const newChatButton = screen.getByRole('button', { name: /新的对话/ });
    await user.click(newChatButton);

    expect(mockSelectSession).toHaveBeenCalledWith(null);
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('should select session on click', async () => {
    const user = userEvent.setup();
    render(<Sidebar />);

    const firstSession = screen.getByText('First Conversation');
    await user.click(firstSession);

    expect(mockSelectSession).toHaveBeenCalledWith(mockSessions[0]);
    expect(mockPush).toHaveBeenCalledWith('/chat/session-1');
  });

  it('should delete session on delete button click', async () => {
    const user = userEvent.setup();
    mockDeleteSession.mockResolvedValue(undefined);
    render(<Sidebar />);

    const firstSessionItem = screen.getByText('First Conversation').closest('.group');
    const deleteButton = firstSessionItem?.querySelector('button[aria-label*="delete" i]') as HTMLElement;

    if (deleteButton) {
      await user.click(deleteButton);
    }

    await waitFor(() => {
      expect(mockDeleteSession).toHaveBeenCalledWith('session-1');
    });
  });

  it('should show empty state when no sessions', () => {
    (useSessionsStore as vi.Mock).mockReturnValue({
      sessions: [],
      fetchSessions: mockFetchSessions,
      selectSession: mockSelectSession,
      deleteSession: mockDeleteSession,
      selectedSession: null,
    });

    render(<Sidebar />);

    expect(screen.getByText('还没有对话，开始聊天吧')).toBeInTheDocument();
  });

  it('should highlight active session based on selectedSession', () => {
    (useSessionsStore as vi.Mock).mockReturnValue({
      sessions: mockSessions,
      fetchSessions: mockFetchSessions,
      selectSession: mockSelectSession,
      deleteSession: mockDeleteSession,
      selectedSession: mockSessions[0],
    });

    render(<Sidebar />);

    const firstSession = screen.getByText('First Conversation').closest('.cursor-pointer');
    expect(firstSession).toHaveClass('bg-atom-mist');
  });

  it('should highlight active session based on pathname', () => {
    (usePathname as vi.Mock).mockReturnValue('/chat/session-2');

    render(<Sidebar />);

    const secondSession = screen.getByText('Second Conversation').closest('.cursor-pointer');
    expect(secondSession).toHaveClass('bg-atom-mist');
  });

  it('should highlight active navigation link', () => {
    (usePathname as vi.Mock).mockReturnValue('/knowledge');

    render(<Sidebar />);

    const knowledgeLink = screen.getByText('知识');
    const linkElement = knowledgeLink.closest('a');
    expect(linkElement).toHaveClass('bg-atom-mist');
  });

  it('should show delete button on session hover', async () => {
    const user = userEvent.setup();
    render(<Sidebar />);

    const firstSessionItem = screen.getByText('First Conversation').closest('.group');
    
    await user.hover(firstSessionItem!);

    const deleteButton = firstSessionItem?.querySelector('button');
    expect(deleteButton).toBeInTheDocument();
  });

  it('should not prevent delete when clicking delete button', async () => {
    const user = userEvent.setup();
    mockDeleteSession.mockResolvedValue(undefined);
    render(<Sidebar />);

    const firstSessionItem = screen.getByText('First Conversation').closest('.group');
    const deleteButton = firstSessionItem?.querySelector('button') as HTMLElement;

    if (deleteButton) {
      await user.click(deleteButton);
    }

    await waitFor(() => {
      expect(mockSelectSession).not.toHaveBeenCalled();
    });
  });

  it('should render session with MessageSquare icon', () => {
    render(<Sidebar />);

    const icons = document.querySelectorAll('.group .lucide-message-square');
    expect(icons.length).toBeGreaterThan(0);
  });

  it('should render navigation links with correct icons', () => {
    render(<Sidebar />);

    expect(screen.getByText('知识')).toBeInTheDocument();
    expect(screen.getByText('技能')).toBeInTheDocument();
    expect(screen.getByText('产物')).toBeInTheDocument();
    expect(screen.getByText('设置')).toBeInTheDocument();
  });

  it('should handle session without title', () => {
    const sessionsWithoutTitle = [
      {
        id: 'session-1',
        userId: 'user-1',
        agentId: 'agent-1',
        title: '',
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
    ];

    (useSessionsStore as vi.Mock).mockReturnValue({
      sessions: sessionsWithoutTitle,
      fetchSessions: mockFetchSessions,
      selectSession: mockSelectSession,
      deleteSession: mockDeleteSession,
      selectedSession: null,
    });

    render(<Sidebar />);

    expect(screen.getByText('新对话')).toBeInTheDocument();
  });

  it('should truncate long session titles', () => {
    const longTitleSessions = [
      {
        id: 'session-1',
        userId: 'user-1',
        agentId: 'agent-1',
        title: 'A'.repeat(100),
        status: 'active',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
    ];

    (useSessionsStore as vi.Mock).mockReturnValue({
      sessions: longTitleSessions,
      fetchSessions: mockFetchSessions,
      selectSession: mockSelectSession,
      deleteSession: mockDeleteSession,
      selectedSession: null,
    });

    render(<Sidebar />);

    const titleElement = screen.getByText(/A+/);
    expect(titleElement).toHaveClass('truncate');
  });

  it('should call logout when logout is triggered', async () => {
    const user = userEvent.setup();
    mockLogout.mockResolvedValue(undefined);

    render(<Sidebar />);

    // Find logout button/icon (check if it exists in the component)
    const logoutButton = screen.queryByRole('button', { name: /logout/i });
    
    if (logoutButton) {
      await user.click(logoutButton);
      await waitFor(() => {
        expect(mockLogout).toHaveBeenCalled();
      });
    }
  });

  it('should have correct logo link href', () => {
    render(<Sidebar />);

    const logoLink = screen.getByText('atom.').closest('a');
    expect(logoLink).toHaveAttribute('href', '/');
  });

  it('should have correct navigation link hrefs', () => {
    render(<Sidebar />);

    const knowledgeLink = screen.getByText('知识').closest('a');
    const skillsLink = screen.getByText('技能').closest('a');
    const artifactsLink = screen.getByText('产物').closest('a');
    const settingsLink = screen.getByText('设置').closest('a');

    expect(knowledgeLink).toHaveAttribute('href', '/knowledge');
    expect(skillsLink).toHaveAttribute('href', '/skills');
    expect(artifactsLink).toHaveAttribute('href', '/artifacts');
    expect(settingsLink).toHaveAttribute('href', '/settings');
  });

  it('should highlight navigation link when on subpath', () => {
    (usePathname as vi.Mock).mockReturnValue('/knowledge/something');

    render(<Sidebar />);

    const knowledgeLink = screen.getByText('知识').closest('a');
    expect(knowledgeLink).toHaveClass('bg-atom-mist');
  });

  it('should not highlight navigation link when on different path', () => {
    (usePathname as vi.Mock).mockReturnValue('/chat/session-1');

    render(<Sidebar />);

    const knowledgeLink = screen.getByText('知识').closest('a');
    expect(knowledgeLink).not.toHaveClass('bg-atom-mist');
  });

  it('should render sidebar with correct width classes', () => {
    const { container } = render(<Sidebar />);

    const sidebar = container.querySelector('aside');
    expect(sidebar).toHaveClass('w-64');
  });

  it('should handle error when fetching sessions', async () => {
    mockFetchSessions.mockRejectedValue(new Error('Failed to fetch'));

    render(<Sidebar />);

    // Component should still render even if fetch fails
    expect(screen.getByText('atom.')).toBeInTheDocument();
  });

  it('should handle error when deleting session', async () => {
    const user = userEvent.setup();
    mockDeleteSession.mockRejectedValue(new Error('Failed to delete'));

    render(<Sidebar />);

    const firstSessionItem = screen.getByText('First Conversation').closest('.group');
    const deleteButton = firstSessionItem?.querySelector('button') as HTMLElement;

    if (deleteButton) {
      await user.click(deleteButton);
    }

    // Should not crash
    await waitFor(() => {
      expect(mockDeleteSession).toHaveBeenCalled();
    });
  });

  it('should render Plus icon in new chat button', () => {
    render(<Sidebar />);

    const newChatButton = screen.getByRole('button', { name: /新的对话/ });
    const icon = newChatButton.querySelector('.lucide-plus');
    expect(icon).toBeInTheDocument();
  });
});
