import { create } from 'zustand';
import { Skill } from '@/types';
import { skillsApi } from '@/lib/api';

interface SkillsStore {
  skills: Skill[];
  category: string;
  isLoading: boolean;
  error: string | null;
  fetchSkills: (category?: string) => Promise<void>;
  createSkill: (data: Partial<Skill>) => Promise<Skill>;
  toggleSkill: (id: string, enabled: boolean) => void;
  deleteSkill: (id: string) => Promise<void>;
  setCategory: (cat: string) => void;
}

export const useSkillsStore = create<SkillsStore>((set, get) => ({
  skills: [],
  category: '',
  isLoading: false,
  error: null,

  fetchSkills: async (category?: string) => {
    set({ isLoading: true, error: null });
    try {
      const skills = await skillsApi.list(category);
      set({ skills, isLoading: false });
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  createSkill: async (data: Partial<Skill>) => {
    set({ isLoading: true, error: null });
    try {
      const skill = await skillsApi.create(data);
      set((state) => ({
        skills: [skill, ...state.skills],
        isLoading: false,
      }));
      return skill;
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
      throw error;
    }
  },

  toggleSkill: (id: string, enabled: boolean) => {
    // Optimistic update
    set((state) => ({
      skills: state.skills.map((s) =>
        s.id === id ? { ...s, enabled } : s
      ),
    }));

    // Fire API call in background; revert on failure
    skillsApi.toggle(id, enabled).catch(() => {
      set((state) => ({
        skills: state.skills.map((s) =>
          s.id === id ? { ...s, enabled: !enabled } : s
        ),
        error: '切换技能状态失败',
      }));
    });
  },

  deleteSkill: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await skillsApi.delete(id);
      set((state) => ({
        skills: state.skills.filter((s) => s.id !== id),
        isLoading: false,
      }));
    } catch (error: any) {
      set({ error: error.message, isLoading: false });
    }
  },

  setCategory: (cat: string) => {
    set({ category: cat });
    get().fetchSkills(cat || undefined);
  },
}));
