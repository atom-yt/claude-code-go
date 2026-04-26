import { create } from 'zustand';
import { Knowledge } from '@/types';
import { knowledgeApi } from '@/lib/api';

interface KnowledgeStore {
  items: Knowledge[];
  selectedItem: Knowledge | null;
  isLoading: boolean;
  error: string | null;
  fetchKnowledge: () => Promise<void>;
  createKnowledge: (data: Partial<Knowledge>) => Promise<void>;
  updateKnowledge: (id: string, data: Partial<Knowledge>) => Promise<void>;
  deleteKnowledge: (id: string) => Promise<void>;
  setSelectedItem: (item: Knowledge | null) => void;
}

export const useKnowledgeStore = create<KnowledgeStore>((set, get) => ({
  items: [],
  selectedItem: null,
  isLoading: false,
  error: null,

  fetchKnowledge: async () => {
    set({ isLoading: true, error: null });
    try {
      const items = await knowledgeApi.list();
      set({ items, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  createKnowledge: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const item = await knowledgeApi.create(data);
      set((state) => ({
        items: [item, ...state.items],
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  updateKnowledge: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      const updated = await knowledgeApi.update(id, data);
      set((state) => ({
        items: state.items.map((item) => (item.id === id ? updated : item)),
        selectedItem: state.selectedItem?.id === id ? updated : state.selectedItem,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  deleteKnowledge: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await knowledgeApi.delete(id);
      set((state) => ({
        items: state.items.filter((item) => item.id !== id),
        selectedItem: state.selectedItem?.id === id ? null : state.selectedItem,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  setSelectedItem: (item) => {
    set({ selectedItem: item });
  },
}));
