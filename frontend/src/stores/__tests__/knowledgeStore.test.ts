import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKnowledgeStore } from '../knowledgeStore';
import { knowledgeApi } from '@/lib/api';
import type { Knowledge } from '@/types';

// Mock the knowledgeApi
vi.mock('@/lib/api', () => ({
  knowledgeApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

const mockKnowledgeItems: Knowledge[] = [
  {
    id: 'knowledge-1',
    userId: 'user-1',
    name: 'Project Documentation',
    description: 'Documentation for the project',
    type: 'document',
    source: 'user',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'knowledge-2',
    userId: 'user-1',
    name: 'Code Base',
    description: 'Source code knowledge',
    type: 'code',
    source: 'ark',
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
];

describe('knowledgeStore', () => {
  beforeEach(() => {
    useKnowledgeStore.setState({
      items: [],
      selectedItem: null,
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchKnowledge', () => {
    it('should fetch knowledge items successfully', async () => {
      vi.mocked(knowledgeApi.list).mockResolvedValue(mockKnowledgeItems);
      await useKnowledgeStore.getState().fetchKnowledge();
      expect(knowledgeApi.list).toHaveBeenCalled();
      expect(useKnowledgeStore.getState().items).toEqual(mockKnowledgeItems);
      expect(useKnowledgeStore.getState().isLoading).toBe(false);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(knowledgeApi.list).mockRejectedValue(error);
      await useKnowledgeStore.getState().fetchKnowledge();
      expect(useKnowledgeStore.getState().error).toBe('Network error');
    });
  });

  describe('createKnowledge', () => {
    it('should create knowledge item successfully', async () => {
      const newItem: Knowledge = {
        id: 'knowledge-3',
        userId: 'user-1',
        name: 'New Knowledge',
        description: 'A new knowledge item',
        type: 'document',
        source: 'user',
        createdAt: '2024-01-03T00:00:00Z',
        updatedAt: '2024-01-03T00:00:00Z',
      };
      vi.mocked(knowledgeApi.create).mockResolvedValue(newItem);
      await useKnowledgeStore.getState().createKnowledge({
        name: 'New Knowledge',
        description: 'A new knowledge item',
        type: 'document',
        source: 'user',
      });
      expect(useKnowledgeStore.getState().items[0]).toEqual(newItem);
    });

    it('should handle create errors', async () => {
      const error = new Error('Failed to create knowledge item');
      vi.mocked(knowledgeApi.create).mockRejectedValue(error);
      await useKnowledgeStore.getState().createKnowledge({ name: 'New', description: '', type: 'doc', source: 'user' });
      expect(useKnowledgeStore.getState().error).toBe('Failed to create knowledge item');
    });
  });

  describe('updateKnowledge', () => {
    it('should update knowledge item successfully', async () => {
      const updatedItem: Knowledge = { ...mockKnowledgeItems[0], name: 'Updated Documentation' };
      useKnowledgeStore.setState({ items: mockKnowledgeItems });
      vi.mocked(knowledgeApi.update).mockResolvedValue(updatedItem);
      await useKnowledgeStore.getState().updateKnowledge('knowledge-1', { name: 'Updated Documentation' });
      expect(useKnowledgeStore.getState().items[0].name).toBe('Updated Documentation');
    });

    it('should update selected item if it matches', async () => {
      const updatedItem: Knowledge = { ...mockKnowledgeItems[0], name: 'Updated' };
      useKnowledgeStore.setState({ items: mockKnowledgeItems, selectedItem: mockKnowledgeItems[0] });
      vi.mocked(knowledgeApi.update).mockResolvedValue(updatedItem);
      await useKnowledgeStore.getState().updateKnowledge('knowledge-1', { name: 'Updated' });
      expect(useKnowledgeStore.getState().selectedItem).toEqual(updatedItem);
    });
  });

  describe('deleteKnowledge', () => {
    it('should delete knowledge item successfully', async () => {
      useKnowledgeStore.setState({ items: mockKnowledgeItems });
      vi.mocked(knowledgeApi.delete).mockResolvedValue(undefined);
      await useKnowledgeStore.getState().deleteKnowledge('knowledge-1');
      expect(useKnowledgeStore.getState().items).toHaveLength(1);
    });

    it('should clear selected item if deleted', async () => {
      useKnowledgeStore.setState({ items: mockKnowledgeItems, selectedItem: mockKnowledgeItems[0] });
      vi.mocked(knowledgeApi.delete).mockResolvedValue(undefined);
      await useKnowledgeStore.getState().deleteKnowledge('knowledge-1');
      expect(useKnowledgeStore.getState().selectedItem).toBe(null);
    });
  });

  describe('setSelectedItem', () => {
    it('should set selected item', () => {
      useKnowledgeStore.getState().setSelectedItem(mockKnowledgeItems[0]);
      expect(useKnowledgeStore.getState().selectedItem).toEqual(mockKnowledgeItems[0]);
    });

    it('should clear selected item', () => {
      useKnowledgeStore.setState({ selectedItem: mockKnowledgeItems[0] });
      useKnowledgeStore.getState().setSelectedItem(null);
      expect(useKnowledgeStore.getState().selectedItem).toBe(null);
    });
  });
});
