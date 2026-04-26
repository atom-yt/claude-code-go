import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useAgentsStore } from '../agentsStore';
import { agentsApi } from '@/lib/api';
import type { Agent } from '@/types';

// Mock the agentsApi
vi.mock('@/lib/api', () => ({
  agentsApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

const mockAgents: Agent[] = [
  {
    id: 'agent-1',
    userId: 'user-1',
    name: 'Test Agent 1',
    description: 'First test agent',
    systemPrompt: 'You are a helpful assistant',
    tools: ['search', 'calculator'],
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'agent-2',
    userId: 'user-1',
    name: 'Test Agent 2',
    description: 'Second test agent',
    systemPrompt: 'You are an expert',
    tools: ['search'],
    config: {},
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
];

describe('agentsStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAgentsStore.setState({
      agents: [],
      selectedAgent: null,
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchAgents', () => {
    it('should fetch agents successfully', async () => {
      vi.mocked(agentsApi.list).mockResolvedValue(mockAgents);

      await useAgentsStore.getState().fetchAgents();

      expect(agentsApi.list).toHaveBeenCalled();
      expect(useAgentsStore.getState().agents).toEqual(mockAgents);
      expect(useAgentsStore.getState().isLoading).toBe(false);
      expect(useAgentsStore.getState().error).toBe(null);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(agentsApi.list).mockRejectedValue(error);

      await useAgentsStore.getState().fetchAgents();

      expect(useAgentsStore.getState().agents).toEqual([]);
      expect(useAgentsStore.getState().error).toBe('Network error');
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });

    it('should set loading state during fetch', async () => {
      vi.mocked(agentsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockAgents), 10)
      ));

      const fetchPromise = useAgentsStore.getState().fetchAgents();
      expect(useAgentsStore.getState().isLoading).toBe(true);
      await fetchPromise;
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });
  });

  describe('createAgent', () => {
    it('should create agent successfully', async () => {
      const newAgent: Agent = {
        id: 'agent-3',
        userId: 'user-1',
        name: 'New Agent',
        description: 'A new agent',
        systemPrompt: 'You are new',
        tools: [],
        config: {},
        createdAt: '2024-01-03T00:00:00Z',
        updatedAt: '2024-01-03T00:00:00Z',
      };

      vi.mocked(agentsApi.create).mockResolvedValue(newAgent);

      const result = await useAgentsStore.getState().createAgent({
        name: 'New Agent',
        description: 'A new agent',
        systemPrompt: 'You are new',
        tools: [],
      });

      expect(agentsApi.create).toHaveBeenCalledWith({
        name: 'New Agent',
        description: 'A new agent',
        systemPrompt: 'You are new',
        tools: [],
      });
      expect(result).toEqual(newAgent);
      expect(useAgentsStore.getState().agents[0]).toEqual(newAgent);
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });

    it('should handle create errors', async () => {
      const error = new Error('Failed to create agent');
      vi.mocked(agentsApi.create).mockRejectedValue(error);

      await expect(
        useAgentsStore.getState().createAgent({
          name: 'New Agent',
          description: 'A new agent',
          systemPrompt: 'You are new',
          tools: [],
        })
      ).rejects.toThrow('Failed to create agent');

      expect(useAgentsStore.getState().error).toBe('Failed to create agent');
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });
  });

  describe('updateAgent', () => {
    it('should update agent successfully', async () => {
      const updatedAgent: Agent = {
        ...mockAgents[0],
        name: 'Updated Agent 1',
        description: 'Updated description',
      };

      useAgentsStore.setState({ agents: mockAgents });
      vi.mocked(agentsApi.update).mockResolvedValue(updatedAgent);

      await useAgentsStore.getState().updateAgent('agent-1', {
        name: 'Updated Agent 1',
        description: 'Updated description',
      });

      expect(agentsApi.update).toHaveBeenCalledWith('agent-1', {
        name: 'Updated Agent 1',
        description: 'Updated description',
      });
      expect(useAgentsStore.getState().agents[0]).toEqual(updatedAgent);
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });

    it('should handle update errors', async () => {
      useAgentsStore.setState({ agents: mockAgents });
      const error = new Error('Failed to update agent');
      vi.mocked(agentsApi.update).mockRejectedValue(error);

      await expect(
        useAgentsStore.getState().updateAgent('agent-1', { name: 'Updated' })
      ).rejects.toThrow('Failed to update agent');

      expect(useAgentsStore.getState().error).toBe('Failed to update agent');
      expect(useAgentsStore.getState().agents[0].name).toBe('Test Agent 1'); // Not updated
    });
  });

  describe('deleteAgent', () => {
    it('should delete agent successfully', async () => {
      useAgentsStore.setState({ agents: mockAgents });
      vi.mocked(agentsApi.delete).mockResolvedValue(undefined);

      await useAgentsStore.getState().deleteAgent('agent-1');

      expect(agentsApi.delete).toHaveBeenCalledWith('agent-1');
      expect(useAgentsStore.getState().agents).toHaveLength(1);
      expect(useAgentsStore.getState().agents[0].id).toBe('agent-2');
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });

    it('should clear selected agent if deleted', async () => {
      useAgentsStore.setState({
        agents: mockAgents,
        selectedAgent: mockAgents[0],
      });
      vi.mocked(agentsApi.delete).mockResolvedValue(undefined);

      await useAgentsStore.getState().deleteAgent('agent-1');

      expect(useAgentsStore.getState().selectedAgent).toBe(null);
    });

    it('should keep selected agent if different agent is deleted', async () => {
      useAgentsStore.setState({
        agents: mockAgents,
        selectedAgent: mockAgents[1],
      });
      vi.mocked(agentsApi.delete).mockResolvedValue(undefined);

      await useAgentsStore.getState().deleteAgent('agent-1');

      expect(useAgentsStore.getState().selectedAgent).toEqual(mockAgents[1]);
    });

    it('should handle delete errors', async () => {
      useAgentsStore.setState({ agents: mockAgents });
      const error = new Error('Failed to delete agent');
      vi.mocked(agentsApi.delete).mockRejectedValue(error);

      await expect(
        useAgentsStore.getState().deleteAgent('agent-1')
      ).rejects.toThrow('Failed to delete agent');

      expect(useAgentsStore.getState().error).toBe('Failed to delete agent');
      expect(useAgentsStore.getState().agents).toHaveLength(2); // Not deleted
    });
  });

  describe('setSelectedAgent', () => {
    it('should set selected agent', () => {
      useAgentsStore.getState().setSelectedAgent(mockAgents[0]);

      expect(useAgentsStore.getState().selectedAgent).toEqual(mockAgents[0]);
    });

    it('should clear selected agent', () => {
      useAgentsStore.setState({ selectedAgent: mockAgents[0] });
      useAgentsStore.getState().setSelectedAgent(null);

      expect(useAgentsStore.getState().selectedAgent).toBe(null);
    });
  });

  describe('setAgents', () => {
    it('should set agents', () => {
      useAgentsStore.getState().setAgents(mockAgents);

      expect(useAgentsStore.getState().agents).toEqual(mockAgents);
    });

    it('should replace existing agents', () => {
      useAgentsStore.setState({ agents: [mockAgents[0]] });
      useAgentsStore.getState().setAgents(mockAgents);

      expect(useAgentsStore.getState().agents).toHaveLength(2);
    });
  });

  describe('setError', () => {
    it('should set error', () => {
      useAgentsStore.getState().setError('Test error');

      expect(useAgentsStore.getState().error).toBe('Test error');
    });

    it('should clear error', () => {
      useAgentsStore.setState({ error: 'Old error' });
      useAgentsStore.getState().setError(null);

      expect(useAgentsStore.getState().error).toBe(null);
    });
  });

  describe('loading state', () => {
    it('should set isLoading during operations', async () => {
      vi.mocked(agentsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockAgents), 10)
      ));

      const promise = useAgentsStore.getState().fetchAgents();
      expect(useAgentsStore.getState().isLoading).toBe(true);
      await promise;
      expect(useAgentsStore.getState().isLoading).toBe(false);
    });

    it('should reset isLoading after error', async () => {
      vi.mocked(agentsApi.list).mockRejectedValue(new Error('Error'));

      await useAgentsStore.getState().fetchAgents();

      expect(useAgentsStore.getState().isLoading).toBe(false);
    });
  });
});