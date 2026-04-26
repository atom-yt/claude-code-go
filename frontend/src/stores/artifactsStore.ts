import { create } from 'zustand';
import { Artifact } from '@/types';
import { artifactsApi } from '@/lib/api';

interface ArtifactsStore {
  artifacts: Artifact[];
  total: number;
  search: string;
  viewMode: 'list' | 'grid';
  isLoading: boolean;
  error: string | null;
  fetchArtifacts: (search?: string) => Promise<void>;
  deleteArtifact: (id: string) => Promise<void>;
  setSearch: (s: string) => void;
  setViewMode: (mode: 'list' | 'grid') => void;
}

export const useArtifactsStore = create<ArtifactsStore>((set, get) => ({
  artifacts: [],
  total: 0,
  search: '',
  viewMode: 'grid',
  isLoading: false,
  error: null,

  fetchArtifacts: async (search?: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await artifactsApi.list(search);
      set({ artifacts: response.items, total: response.total, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  deleteArtifact: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await artifactsApi.delete(id);
      set((state) => ({
        artifacts: state.artifacts.filter((a) => a.id !== id),
        total: state.total - 1,
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  setSearch: (s: string) => {
    set({ search: s });
  },

  setViewMode: (mode: 'list' | 'grid') => {
    set({ viewMode: mode });
  },
}));
