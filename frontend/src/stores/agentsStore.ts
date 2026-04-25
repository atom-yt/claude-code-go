import { create } from 'zustand';
import { AgentsState, Agent } from '@/types';
import { agentsApi } from '@/lib/api';

interface AgentsStore extends AgentsState {
  setAgents: (agents: Agent[]) => void;
  setSelectedAgent: (agent: Agent | null) => void;
  setError: (error: string | null) => void;
  fetchAgents: () => Promise<void>;
  createAgent: (data: any) => Promise<Agent>;
  updateAgent: (id: string, data: any) => Promise<void>;
  deleteAgent: (id: string) => Promise<void>;
}

export const useAgentsStore = create<AgentsStore>((set, get) => ({
  agents: [],
  selectedAgent: null,
  isLoading: false,
  error: null,

  setAgents: (agents) => set({ agents }),
  setSelectedAgent: (agent) => set({ selectedAgent: agent }),
  setError: (error) => set({ error }),

  fetchAgents: async () => {
    set({ isLoading: true, error: null });
    try {
      const agents = await agentsApi.list();
      set({ agents, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  createAgent: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const agent = await agentsApi.create(data);
      set((state) => ({
        agents: [...state.agents, agent],
        isLoading: false,
      }));
      return agent;
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },

  updateAgent: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      const updatedAgent = await agentsApi.update(id, data);
      set((state) => ({
        agents: state.agents.map((agent) =>
          agent.id === id ? updatedAgent : agent
        ),
        selectedAgent:
          state.selectedAgent?.id === id ? updatedAgent : state.selectedAgent,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },

  deleteAgent: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await agentsApi.delete(id);
      set((state) => ({
        agents: state.agents.filter((agent) => agent.id !== id),
        selectedAgent:
          state.selectedAgent?.id === id ? null : state.selectedAgent,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },
}));