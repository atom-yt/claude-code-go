import { create } from 'zustand';
import { ScheduledTask } from '@/types';
import { schedulesApi } from '@/lib/api';

interface ScheduleStore {
  tasks: ScheduledTask[];
  isLoading: boolean;
  error: string | null;
  fetchTasks: () => Promise<void>;
  createTask: (data: Partial<ScheduledTask>) => Promise<void>;
  updateTask: (id: string, data: Partial<ScheduledTask>) => Promise<void>;
  deleteTask: (id: string) => Promise<void>;
  toggleTask: (id: string, enabled: boolean) => Promise<void>;
}

export const useScheduleStore = create<ScheduleStore>((set) => ({
  tasks: [],
  isLoading: false,
  error: null,

  fetchTasks: async () => {
    set({ isLoading: true, error: null });
    try {
      const tasks = await schedulesApi.list();
      set({ tasks, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  createTask: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const task = await schedulesApi.create(data);
      set((state) => ({
        tasks: [task, ...state.tasks],
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  updateTask: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      const updated = await schedulesApi.update(id, data);
      set((state) => ({
        tasks: state.tasks.map((t) => (t.id === id ? updated : t)),
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  deleteTask: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await schedulesApi.delete(id);
      set((state) => ({
        tasks: state.tasks.filter((t) => t.id !== id),
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  toggleTask: async (id, enabled) => {
    // Optimistic update
    set((state) => ({
      tasks: state.tasks.map((t) => (t.id === id ? { ...t, enabled } : t)),
    }));
    try {
      await schedulesApi.toggle(id, enabled);
    } catch (error: any) {
      // Revert on failure
      set((state) => ({
        tasks: state.tasks.map((t) => (t.id === id ? { ...t, enabled: !enabled } : t)),
        error: error.message,
      }));
    }
  },
}));
