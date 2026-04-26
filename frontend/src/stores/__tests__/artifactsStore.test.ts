import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useArtifactsStore } from '../artifactsStore';
import { artifactsApi } from '@/lib/api';
import type { Artifact } from '@/types';

// Mock the artifactsApi
vi.mock('@/lib/api', () => ({
  artifactsApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
    stats: vi.fn(),
  },
}));

const mockArtifacts: Artifact[] = [
  {
    id: 'artifact-1',
    userId: 'user-1',
    sessionId: 'session-1',
    title: 'Test Document',
    content: 'This is a test document',
    fileType: 'txt',
    tags: ['test', 'document'],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'artifact-2',
    userId: 'user-1',
    title: 'Test Code',
    content: 'console.log("hello")',
    fileType: 'js',
    tags: ['code', 'javascript'],
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
];

describe('artifactsStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useArtifactsStore.setState({
      artifacts: [],
      total: 0,
      search: '',
      viewMode: 'grid',
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchArtifacts', () => {
    it('should fetch artifacts successfully', async () => {
      vi.mocked(artifactsApi.list).mockResolvedValue({
        items: mockArtifacts,
        total: 2,
      });

      await useArtifactsStore.getState().fetchArtifacts();

      expect(artifactsApi.list).toHaveBeenCalledWith(undefined);
      expect(useArtifactsStore.getState().artifacts).toEqual(mockArtifacts);
      expect(useArtifactsStore.getState().total).toBe(2);
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });

    it('should fetch artifacts with search filter', async () => {
      const searchResults = [mockArtifacts[0]];
      vi.mocked(artifactsApi.list).mockResolvedValue({
        items: searchResults,
        total: 1,
      });

      await useArtifactsStore.getState().fetchArtifacts('test');

      expect(artifactsApi.list).toHaveBeenCalledWith('test');
      expect(useArtifactsStore.getState().artifacts).toEqual(searchResults);
      expect(useArtifactsStore.getState().total).toBe(1);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(artifactsApi.list).mockRejectedValue(error);

      await useArtifactsStore.getState().fetchArtifacts();

      expect(useArtifactsStore.getState().artifacts).toEqual([]);
      expect(useArtifactsStore.getState().total).toBe(0);
      expect(useArtifactsStore.getState().error).toBe('Network error');
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });

    it('should set loading state during fetch', async () => {
      vi.mocked(artifactsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve({ items: mockArtifacts, total: 2 }), 10)
      ));

      const fetchPromise = useArtifactsStore.getState().fetchArtifacts();
      expect(useArtifactsStore.getState().isLoading).toBe(true);
      await fetchPromise;
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });
  });

  describe('deleteArtifact', () => {
    it('should delete artifact successfully', async () => {
      useArtifactsStore.setState({ artifacts: mockArtifacts, total: 2 });
      vi.mocked(artifactsApi.delete).mockResolvedValue(undefined);

      await useArtifactsStore.getState().deleteArtifact('artifact-1');

      expect(artifactsApi.delete).toHaveBeenCalledWith('artifact-1');
      expect(useArtifactsStore.getState().artifacts).toHaveLength(1);
      expect(useArtifactsStore.getState().artifacts[0].id).toBe('artifact-2');
      expect(useArtifactsStore.getState().total).toBe(1);
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });

    it('should handle delete errors', async () => {
      useArtifactsStore.setState({ artifacts: mockArtifacts, total: 2 });
      const error = new Error('Failed to delete artifact');
      vi.mocked(artifactsApi.delete).mockRejectedValue(error);

      await useArtifactsStore.getState().deleteArtifact('artifact-1');

      expect(useArtifactsStore.getState().error).toBe('Failed to delete artifact');
      expect(useArtifactsStore.getState().artifacts).toHaveLength(2); // Not deleted
      expect(useArtifactsStore.getState().total).toBe(2);
    });
  });

  describe('search filtering', () => {
    it('should set search query', () => {
      useArtifactsStore.getState().setSearch('test query');

      expect(useArtifactsStore.getState().search).toBe('test query');
    });

    it('should clear search query', () => {
      useArtifactsStore.setState({ search: 'old query' });
      useArtifactsStore.getState().setSearch('');

      expect(useArtifactsStore.getState().search).toBe('');
    });

    it('should fetch with search filter when search is set', async () => {
      vi.mocked(artifactsApi.list).mockResolvedValue({
        items: mockArtifacts,
        total: 2,
      });

      useArtifactsStore.getState().setSearch('test');
      await useArtifactsStore.getState().fetchArtifacts('test');

      expect(artifactsApi.list).toHaveBeenCalledWith('test');
    });

    it('should handle empty search query', async () => {
      vi.mocked(artifactsApi.list).mockResolvedValue({
        items: mockArtifacts,
        total: 2,
      });

      useArtifactsStore.getState().setSearch('');
      await useArtifactsStore.getState().fetchArtifacts('');

      expect(artifactsApi.list).toHaveBeenCalledWith('');
    });
  });

  describe('viewMode', () => {
    it('should set viewMode to list', () => {
      useArtifactsStore.getState().setViewMode('list');

      expect(useArtifactsStore.getState().viewMode).toBe('list');
    });

    it('should set viewMode to grid', () => {
      useArtifactsStore.setState({ viewMode: 'list' });
      useArtifactsStore.getState().setViewMode('grid');

      expect(useArtifactsStore.getState().viewMode).toBe('grid');
    });

    it('should start with grid as default viewMode', () => {
      const store = useArtifactsStore.getState();
      expect(store.viewMode).toBe('grid');
    });
  });

  describe('loading state', () => {
    it('should set isLoading during operations', async () => {
      vi.mocked(artifactsApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve({ items: mockArtifacts, total: 2 }), 10)
      ));

      const promise = useArtifactsStore.getState().fetchArtifacts();
      expect(useArtifactsStore.getState().isLoading).toBe(true);
      await promise;
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });

    it('should set isLoading during delete', async () => {
      useArtifactsStore.setState({ artifacts: mockArtifacts, total: 2 });
      vi.mocked(artifactsApi.delete).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(undefined), 10)
      ));

      const promise = useArtifactsStore.getState().deleteArtifact('artifact-1');
      expect(useArtifactsStore.getState().isLoading).toBe(true);
      await promise;
      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });

    it('should reset isLoading after error', async () => {
      vi.mocked(artifactsApi.list).mockRejectedValue(new Error('Error'));

      await useArtifactsStore.getState().fetchArtifacts();

      expect(useArtifactsStore.getState().isLoading).toBe(false);
    });
  });

  describe('total count', () => {
    it('should update total count on fetch', async () => {
      vi.mocked(artifactsApi.list).mockResolvedValue({
        items: mockArtifacts,
        total: 5,
      });

      await useArtifactsStore.getState().fetchArtifacts();

      expect(useArtifactsStore.getState().total).toBe(5);
    });

    it('should decrement total count on delete', async () => {
      useArtifactsStore.setState({ artifacts: mockArtifacts, total: 3 });
      vi.mocked(artifactsApi.delete).mockResolvedValue(undefined);

      await useArtifactsStore.getState().deleteArtifact('artifact-1');

      expect(useArtifactsStore.getState().total).toBe(2);
    });
  });

  describe('initial state', () => {
    it('should initialize with default values', () => {
      const state = useArtifactsStore.getState();

      expect(state.artifacts).toEqual([]);
      expect(state.total).toBe(0);
      expect(state.search).toBe('');
      expect(state.viewMode).toBe('grid');
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe(null);
    });
  });
});