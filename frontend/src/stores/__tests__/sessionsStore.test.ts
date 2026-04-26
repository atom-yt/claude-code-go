import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useSessionsStore } from '../sessionsStore';
import { sessionsApi, messagesApi } from '@/lib/api';
import type { Session, Message } from '@/types';

// Mock API modules
vi.mock('@/lib/api', () => ({
  sessionsApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
  messagesApi: {
    get: vi.fn(),
    getRecent: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
  agentsApi: {
    list: vi.fn(),
  },
}));

const mockSessions: Session[] = [
  {
    id: 'session-1',
    userId: 'user-1',
    agentId: 'agent-1',
    title: 'Test Session 1',
    status: 'active',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'session-2',
    userId: 'user-1',
    agentId: 'agent-1',
    title: 'Test Session 2',
    status: 'active',
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
];

const mockMessages: Message[] = [
  {
    id: 'msg-1',
    sessionId: 'session-1',
    role: 'user',
    content: 'Hello',
    createdAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'msg-2',
    sessionId: 'session-1',
    role: 'assistant',
    content: 'Hi there!',
    createdAt: '2024-01-01T00:00:01Z',
  },
];

describe('sessionsStore', () => {
  beforeEach(() => {
    useSessionsStore.setState({
      sessions: [],
      selectedSession: null,
      messages: [],
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchSessions', () => {
    it('should fetch sessions successfully', async () => {
      vi.mocked(sessionsApi.list).mockResolvedValue(mockSessions);
      await useSessionsStore.getState().fetchSessions();
      expect(sessionsApi.list).toHaveBeenCalled();
      expect(useSessionsStore.getState().sessions).toEqual(mockSessions);
      expect(useSessionsStore.getState().isLoading).toBe(false);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(sessionsApi.list).mockRejectedValue(error);
      await useSessionsStore.getState().fetchSessions();
      expect(useSessionsStore.getState().sessions).toEqual([]);
      expect(useSessionsStore.getState().error).toBe('Network error');
    });

    it('should set loading state during fetch', async () => {
      vi.mocked(sessionsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockSessions), 10)
      ));
      const fetchPromise = useSessionsStore.getState().fetchSessions();
      expect(useSessionsStore.getState().isLoading).toBe(true);
      await fetchPromise;
      expect(useSessionsStore.getState().isLoading).toBe(false);
    });
  });

  describe('createSession', () => {
    it('should create session successfully', async () => {
      const newSession: Session = {
        id: 'session-3',
        userId: 'user-1',
        agentId: 'agent-1',
        title: 'New Session',
        status: 'active',
        createdAt: '2024-01-03T00:00:00Z',
        updatedAt: '2024-01-03T00:00:00Z',
      };
      vi.mocked(sessionsApi.create).mockResolvedValue(newSession);
      const result = await useSessionsStore.getState().createSession({ title: 'New Session', agentId: 'agent-1' });
      expect(sessionsApi.create).toHaveBeenCalledWith({ title: 'New Session', agentId: 'agent-1' });
      expect(result).toEqual(newSession);
      expect(useSessionsStore.getState().sessions[0]).toEqual(newSession);
    });

    it('should use default agent if not provided', async () => {
      const newSession: Session = {
        id: 'session-3',
        userId: 'user-1',
        agentId: '00000000-0000-0000-0000-000000000001',
        title: 'New Session',
        status: 'active',
        createdAt: '2024-01-03T00:00:00Z',
        updatedAt: '2024-01-03T00:00:00Z',
      };
      vi.mocked(sessionsApi.create).mockResolvedValue(newSession);
      await useSessionsStore.getState().createSession({ title: 'New Session' });
      expect(sessionsApi.create).toHaveBeenCalledWith({ title: 'New Session', agentId: '00000000-0000-0000-0000-000000000001' });
    });

    it('should handle create errors', async () => {
      const error = new Error('Failed to create session');
      vi.mocked(sessionsApi.create).mockRejectedValue(error);
      await expect(
        useSessionsStore.getState().createSession({ title: 'New Session' })
      ).rejects.toThrow('Failed to create session');
    });
  });

  describe('deleteSession', () => {
    it('should delete session successfully', async () => {
      useSessionsStore.setState({ sessions: mockSessions });
      vi.mocked(sessionsApi.delete).mockResolvedValue(undefined);
      await useSessionsStore.getState().deleteSession('session-1');
      expect(sessionsApi.delete).toHaveBeenCalledWith('session-1');
      expect(useSessionsStore.getState().sessions).toHaveLength(1);
    });

    it('should clear selected session if deleted', async () => {
      useSessionsStore.setState({ sessions: mockSessions, selectedSession: mockSessions[0], messages: mockMessages });
      vi.mocked(sessionsApi.delete).mockResolvedValue(undefined);
      await useSessionsStore.getState().deleteSession('session-1');
      expect(useSessionsStore.getState().selectedSession).toBe(null);
      expect(useSessionsStore.getState().messages).toEqual([]);
    });

    it('should keep selected session if different session is deleted', async () => {
      useSessionsStore.setState({ sessions: mockSessions, selectedSession: mockSessions[1], messages: mockMessages });
      vi.mocked(sessionsApi.delete).mockResolvedValue(undefined);
      await useSessionsStore.getState().deleteSession('session-1');
      expect(useSessionsStore.getState().selectedSession).toEqual(mockSessions[1]);
      expect(useSessionsStore.getState().messages).toEqual(mockMessages);
    });
  });

  describe('selectSession', () => {
    it('should select session and fetch messages', async () => {
      vi.mocked(messagesApi.getRecent).mockResolvedValue(mockMessages);
      useSessionsStore.getState().selectSession(mockSessions[0]);
      expect(useSessionsStore.getState().selectedSession).toEqual(mockSessions[0]);
      expect(messagesApi.getRecent).toHaveBeenCalledWith('session-1');
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(useSessionsStore.getState().messages).toEqual(mockMessages);
    });

    it('should clear session and messages when selecting null', () => {
      useSessionsStore.setState({ selectedSession: mockSessions[0], messages: mockMessages });
      useSessionsStore.getState().selectSession(null);
      expect(useSessionsStore.getState().selectedSession).toBe(null);
      expect(useSessionsStore.getState().messages).toEqual([]);
    });
  });

  describe('fetchMessages', () => {
    it('should fetch messages successfully', async () => {
      vi.mocked(messagesApi.getRecent).mockResolvedValue(mockMessages);
      await useSessionsStore.getState().fetchMessages('session-1');
      expect(messagesApi.getRecent).toHaveBeenCalledWith('session-1');
      expect(useSessionsStore.getState().messages).toEqual(mockMessages);
      expect(useSessionsStore.getState().isLoading).toBe(false);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(messagesApi.getRecent).mockRejectedValue(error);
      await useSessionsStore.getState().fetchMessages('session-1');
      expect(useSessionsStore.getState().error).toBe('Network error');
    });
  });

  describe('addMessage', () => {
    it('should add message to state', () => {
      const newMessage: Message = {
        id: 'msg-3',
        sessionId: 'session-1',
        role: 'user',
        content: 'New message',
        createdAt: '2024-01-01T00:00:02Z',
      };
      useSessionsStore.setState({ messages: mockMessages });
      useSessionsStore.getState().addMessage(newMessage);
      expect(useSessionsStore.getState().messages).toHaveLength(3);
      expect(useSessionsStore.getState().messages[2]).toEqual(newMessage);
    });
  });

  describe('clearMessages', () => {
    it('should clear all messages', () => {
      useSessionsStore.setState({ messages: mockMessages });
      useSessionsStore.getState().clearMessages();
      expect(useSessionsStore.getState().messages).toEqual([]);
    });
  });

  describe('loading state', () => {
    it('should set isLoading during operations', async () => {
      vi.mocked(sessionsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockSessions), 10)
      ));
      const promise = useSessionsStore.getState().fetchSessions();
      expect(useSessionsStore.getState().isLoading).toBe(true);
      await promise;
      expect(useSessionsStore.getState().isLoading).toBe(false);
    });

    it('should reset isLoading after error', async () => {
      vi.mocked(sessionsApi.list).mockRejectedValue(new Error('Error'));
      await useSessionsStore.getState().fetchSessions();
      expect(useSessionsStore.getState().isLoading).toBe(false);
    });
  });
});
