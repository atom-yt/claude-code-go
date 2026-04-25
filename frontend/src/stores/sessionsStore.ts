import { create } from 'zustand';
import { Session, Message, SessionsState } from '@/types';
import { sessionsApi } from '@/lib/api';

interface SessionsStore extends SessionsState {
  fetchSessions: () => Promise<void>;
  fetchMessages: (sessionId: string) => Promise<void>;
  selectSession: (session: Session | null) => void;
  createSession: (data: { title?: string; model?: string; provider?: string }) => Promise<Session>;
  deleteSession: (id: string) => Promise<void>;
  sendMessage: (content: string, onMessage?: (message: Message) => void) => Promise<void>;
  addMessage: (message: Message) => void;
}

export const useSessionsStore = create<SessionsStore>((set, get) => ({
  sessions: [],
  selectedSession: null,
  messages: [],
  isLoading: false,
  error: null,

  fetchSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const sessions = await sessionsApi.list();
      set({ sessions, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  fetchMessages: async (sessionId: string) => {
    set({ isLoading: true, error: null });
    try {
      const data = await sessionsApi.get(sessionId);
      set({
        messages: data.messages || [],
        isLoading: false,
      });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  selectSession: (session) => {
    set({ selectedSession: session, messages: [] });
    if (session) {
      get().fetchMessages(session.id);
    }
  },

  createSession: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const session = await sessionsApi.create(data);
      set((state) => ({
        sessions: [session, ...state.sessions],
        isLoading: false,
      }));
      return session;
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },

  deleteSession: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await sessionsApi.delete(id);
      set((state) => ({
        sessions: state.sessions.filter((s) => s.id !== id),
        selectedSession: state.selectedSession?.id === id ? null : state.selectedSession,
        messages: state.selectedSession?.id === id ? [] : state.messages,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  sendMessage: async (content, onMessage) => {
    const { selectedSession } = get();
    if (!selectedSession) {
      throw new Error('No session selected');
    }

    set({ isLoading: true, error: null });

    // Add user message
    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      sessionId: selectedSession.id,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };

    set((state) => ({
      messages: [...state.messages, userMessage],
    }));

    if (onMessage) onMessage(userMessage);

    try {
      // For streaming, we'll use WebSocket
      // For now, use REST API
      const response = await sessionsApi.create(selectedSession.id + '/messages', { content, stream: true });
      const assistantMessage: Message = {
        id: response.id || `msg-${Date.now() + 1}`,
        sessionId: selectedSession.id,
        role: response.role || 'assistant',
        content: response.content || '',
        toolCalls: response.toolCalls,
        createdAt: response.createdAt || new Date().toISOString(),
      };

      set((state) => ({
        messages: [...state.messages, assistantMessage],
        isLoading: false,
      }));

      if (onMessage) onMessage(assistantMessage);
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  addMessage: (message) => {
    set((state) => ({
      messages: [...state.messages, message],
    }));
  },
}));