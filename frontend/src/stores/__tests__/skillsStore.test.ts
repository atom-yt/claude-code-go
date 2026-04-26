import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useSkillsStore } from '../skillsStore';
import { skillsApi } from '@/lib/api';
import type { Skill } from '@/types';

// Mock the skillsApi
vi.mock('@/lib/api', () => ({
  skillsApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    toggle: vi.fn(),
  },
}));

const mockSkills: Skill[] = [
  {
    id: 'skill-1',
    userId: 'user-1',
    name: 'Web Search',
    description: 'Search the web',
    category: 'personal',
    icon: 'search',
    enabled: true,
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'skill-2',
    userId: 'user-1',
    name: 'Calculator',
    description: 'Perform calculations',
    category: 'personal',
    icon: 'calculator',
    enabled: false,
    config: {},
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
  {
    id: 'skill-3',
    teamId: 'team-1',
    name: 'Team Skill',
    description: 'Team shared skill',
    category: 'team',
    icon: 'users',
    enabled: true,
    config: {},
    createdAt: '2024-01-03T00:00:00Z',
    updatedAt: '2024-01-03T00:00:00Z',
  },
];

describe('skillsStore', () => {
  beforeEach(() => {
    useSkillsStore.setState({
      skills: [],
      category: '',
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchSkills', () => {
    it('should fetch all skills successfully', async () => {
      vi.mocked(skillsApi.list).mockResolvedValue(mockSkills);
      await useSkillsStore.getState().fetchSkills();
      expect(skillsApi.list).toHaveBeenCalledWith(undefined);
      expect(useSkillsStore.getState().skills).toEqual(mockSkills);
      expect(useSkillsStore.getState().isLoading).toBe(false);
    });

    it('should fetch skills by category', async () => {
      const personalSkills = mockSkills.filter(s => s.category === 'personal');
      vi.mocked(skillsApi.list).mockResolvedValue(personalSkills);
      await useSkillsStore.getState().fetchSkills('personal');
      expect(skillsApi.list).toHaveBeenCalledWith('personal');
      expect(useSkillsStore.getState().skills).toEqual(personalSkills);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(skillsApi.list).mockRejectedValue(error);
      await useSkillsStore.getState().fetchSkills();
      expect(useSkillsStore.getState().skills).toEqual([]);
      expect(useSkillsStore.getState().error).toBe('Network error');
    });

    it('should set loading state during fetch', async () => {
      vi.mocked(skillsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockSkills), 10)
      ));
      const fetchPromise = useSkillsStore.getState().fetchSkills();
      expect(useSkillsStore.getState().isLoading).toBe(true);
      await fetchPromise;
      expect(useSkillsStore.getState().isLoading).toBe(false);
    });
  });

  describe('createSkill', () => {
    it('should create skill successfully', async () => {
      const newSkill: Skill = {
        id: 'skill-4',
        userId: 'user-1',
        name: 'New Skill',
        description: 'A new skill',
        category: 'personal',
        icon: 'star',
        enabled: true,
        config: {},
        createdAt: '2024-01-04T00:00:00Z',
        updatedAt: '2024-01-04T00:00:00Z',
      };
      vi.mocked(skillsApi.create).mockResolvedValue(newSkill);
      const result = await useSkillsStore.getState().createSkill({
        name: 'New Skill',
        description: 'A new skill',
        category: 'personal',
        icon: 'star',
        enabled: true,
      });
      expect(skillsApi.create).toHaveBeenCalled();
      expect(result).toEqual(newSkill);
      expect(useSkillsStore.getState().skills[0]).toEqual(newSkill);
      expect(useSkillsStore.getState().isLoading).toBe(false);
    });

    it('should handle create errors', async () => {
      const error = new Error('Failed to create skill');
      vi.mocked(skillsApi.create).mockRejectedValue(error);
      await expect(
        useSkillsStore.getState().createSkill({
          name: 'New Skill',
          description: 'A new skill',
          category: 'personal',
          icon: 'star',
          enabled: true,
        })
      ).rejects.toThrow('Failed to create skill');
      expect(useSkillsStore.getState().error).toBe('Failed to create skill');
    });
  });

  describe('toggleSkill with optimistic update and rollback', () => {
    it('should toggle skill enabled with optimistic update', async () => {
      useSkillsStore.setState({ skills: mockSkills });
      vi.mocked(skillsApi.toggle).mockResolvedValue(undefined);
      useSkillsStore.getState().toggleSkill('skill-1', false);
      expect(skillsApi.toggle).toHaveBeenCalledWith('skill-1', false);
      expect(useSkillsStore.getState().skills[0].enabled).toBe(false);
    });

    it('should revert optimistic update on failure', async () => {
      useSkillsStore.setState({ skills: mockSkills });
      vi.mocked(skillsApi.toggle).mockRejectedValue(new Error('Failed to toggle'));
      useSkillsStore.getState().toggleSkill('skill-1', false);
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(useSkillsStore.getState().skills[0].enabled).toBe(true);
      expect(useSkillsStore.getState().error).toBe('切换技能状态失败');
    });

    it('should enable a disabled skill', async () => {
      useSkillsStore.setState({ skills: mockSkills });
      vi.mocked(skillsApi.toggle).mockResolvedValue(undefined);
      useSkillsStore.getState().toggleSkill('skill-2', true);
      expect(useSkillsStore.getState().skills[1].enabled).toBe(true);
    });
  });

  describe('deleteSkill', () => {
    it('should delete skill successfully', async () => {
      useSkillsStore.setState({ skills: mockSkills });
      vi.mocked(skillsApi.delete).mockResolvedValue(undefined);
      await useSkillsStore.getState().deleteSkill('skill-1');
      expect(skillsApi.delete).toHaveBeenCalledWith('skill-1');
      expect(useSkillsStore.getState().skills).toHaveLength(2);
      expect(useSkillsStore.getState().skills[0].id).toBe('skill-2');
    });

    it('should handle delete errors', async () => {
      useSkillsStore.setState({ skills: mockSkills });
      const error = new Error('Failed to delete skill');
      vi.mocked(skillsApi.delete).mockRejectedValue(error);
      await useSkillsStore.getState().deleteSkill('skill-1');
      expect(useSkillsStore.getState().error).toBe('Failed to delete skill');
      expect(useSkillsStore.getState().skills).toHaveLength(3);
    });
  });

  describe('setCategory', () => {
    it('should set category and fetch skills', async () => {
      const personalSkills = mockSkills.filter(s => s.category === 'personal');
      vi.mocked(skillsApi.list).mockResolvedValue(personalSkills);
      useSkillsStore.getState().setCategory('personal');
      expect(useSkillsStore.getState().category).toBe('personal');
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(skillsApi.list).toHaveBeenCalledWith('personal');
      expect(useSkillsStore.getState().skills).toEqual(personalSkills);
    });

    it('should clear category and fetch all skills', async () => {
      vi.mocked(skillsApi.list).mockResolvedValue(mockSkills);
      useSkillsStore.getState().setCategory('');
      expect(useSkillsStore.getState().category).toBe('');
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(skillsApi.list).toHaveBeenCalledWith(undefined);
    });
  });

  describe('category filtering', () => {
    it('should filter skills by personal category', async () => {
      const personalSkills = mockSkills.filter(s => s.category === 'personal');
      vi.mocked(skillsApi.list).mockResolvedValue(personalSkills);
      await useSkillsStore.getState().fetchSkills('personal');
      expect(useSkillsStore.getState().skills.every(s => s.category === 'personal')).toBe(true);
    });

    it('should filter skills by team category', async () => {
      const teamSkills = mockSkills.filter(s => s.category === 'team');
      vi.mocked(skillsApi.list).mockResolvedValue(teamSkills);
      await useSkillsStore.getState().fetchSkills('team');
      expect(useSkillsStore.getState().skills.every(s => s.category === 'team')).toBe(true);
    });
  });

  describe('loading state', () => {
    it('should set isLoading during operations', async () => {
      vi.mocked(skillsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockSkills), 10)
      ));
      const promise = useSkillsStore.getState().fetchSkills();
      expect(useSkillsStore.getState().isLoading).toBe(true);
      await promise;
      expect(useSkillsStore.getState().isLoading).toBe(false);
    });

    it('should reset isLoading after error', async () => {
      vi.mocked(skillsApi.list).mockRejectedValue(new Error('Error'));
      await useSkillsStore.getState().fetchSkills();
      expect(useSkillsStore.getState().isLoading).toBe(false);
    });
  });

  describe('initial state', () => {
    it('should initialize with default values', () => {
      const state = useSkillsStore.getState();
      expect(state.skills).toEqual([]);
      expect(state.category).toBe('');
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe(null);
    });
  });
});
