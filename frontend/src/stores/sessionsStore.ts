import { create } from 'zustand';
import { Session, Message, SessionsState } from '@/types';
import { sessionsApi, messagesApi, agentsApi } from '@/lib/api';

// Default agent ID (system default)
const DEFAULT_AGENT_ID = '00000000-0000-0000-0000-000000000001';

interface SessionsStore extends SessionsState {
  fetchSessions: () => Promise<void>;
  fetchMessages: (sessionId: string) => Promise<void>;
  selectSession: (session: Session | null) => void;
  createSession: (data: { title?: string; agentId?: string }) => Promise<Session>;
  deleteSession: (id: string) => Promise<void>;
  addMessage: (message: Message) => void;
  clearMessages: () => void;
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
      const messages = await messagesApi.getRecent(sessionId);
      set({ messages: messages || [], isLoading: false });
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
      // Use provided agentId or default agent
      const agentId = data.agentId || DEFAULT_AGENT_ID;
      const session = await sessionsApi.create({ ...data, agentId });
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

  addMessage: (message) => {
    set((state) => ({
      messages: [...state.messages, message],
    }));
  },

  clearMessages: () => {
    set({ messages: [] });
  },
}));
