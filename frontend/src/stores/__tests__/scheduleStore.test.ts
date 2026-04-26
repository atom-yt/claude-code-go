import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useScheduleStore } from '../scheduleStore';
import { schedulesApi } from '@/lib/api';
import type { ScheduledTask } from '@/types';

// Mock the schedulesApi
vi.mock('@/lib/api', () => ({
  schedulesApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    toggle: vi.fn(),
  },
}));

const mockTasks: ScheduledTask[] = [
  {
    id: 'task-1',
    userId: 'user-1',
    title: 'Daily Report',
    prompt: 'Generate daily report',
    scheduleType: 'daily',
    scheduleTime: '09:00',
    model: 'gpt-4',
    enabled: true,
    notifyOnDone: true,
    executionCount: 5,
    lastRunAt: '2024-01-25T09:00:00Z',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-25T09:00:00Z',
  },
  {
    id: 'task-2',
    userId: 'user-1',
    title: 'Weekly Summary',
    prompt: 'Generate weekly summary',
    scheduleType: 'weekly',
    scheduleTime: '18:00',
    model: 'gpt-4',
    enabled: false,
    notifyOnDone: false,
    executionCount: 2,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('scheduleStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useScheduleStore.setState({
      tasks: [],
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('fetchTasks', () => {
    it('should fetch tasks successfully', async () => {
      vi.mocked(schedulesApi.list).mockResolvedValue(mockTasks);

      await useScheduleStore.getState().fetchTasks();

      expect(schedulesApi.list).toHaveBeenCalled();
      expect(useScheduleStore.getState().tasks).toEqual(mockTasks);
      expect(useScheduleStore.getState().isLoading).toBe(false);
      expect(useScheduleStore.getState().error).toBe(null);
    });

    it('should handle fetch errors', async () => {
      const error = new Error('Network error');
      vi.mocked(schedulesApi.list).mockRejectedValue(error);

      await useScheduleStore.getState().fetchTasks();

      expect(useScheduleStore.getState().tasks).toEqual([]);
      expect(useScheduleStore.getState().error).toBe('Network error');
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should set loading state during fetch', async () => {
      vi.mocked(schedulesApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockTasks), 10)
      ));

      const fetchPromise = useScheduleStore.getState().fetchTasks();
      expect(useScheduleStore.getState().isLoading).toBe(true);
      await fetchPromise;
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });
  });

  describe('createTask', () => {
    it('should create task successfully', async () => {
      const newTask: ScheduledTask = {
        id: 'task-3',
        userId: 'user-1',
        title: 'New Task',
        prompt: 'Execute new task',
        scheduleType: 'daily',
        scheduleTime: '10:00',
        model: 'gpt-4',
        enabled: true,
        notifyOnDone: true,
        executionCount: 0,
        createdAt: '2024-01-03T00:00:00Z',
        updatedAt: '2024-01-03T00:00:00Z',
      };

      vi.mocked(schedulesApi.create).mockResolvedValue(newTask);

      await useScheduleStore.getState().createTask({
        title: 'New Task',
        prompt: 'Execute new task',
        scheduleType: 'daily',
        scheduleTime: '10:00',
        model: 'gpt-4',
        enabled: true,
        notifyOnDone: true,
        executionCount: 0,
      });

      expect(schedulesApi.create).toHaveBeenCalledWith({
        title: 'New Task',
        prompt: 'Execute new task',
        scheduleType: 'daily',
        scheduleTime: '10:00',
        model: 'gpt-4',
        enabled: true,
        notifyOnDone: true,
        executionCount: 0,
      });
      expect(useScheduleStore.getState().tasks[0]).toEqual(newTask);
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should handle create errors', async () => {
      const error = new Error('Failed to create task');
      vi.mocked(schedulesApi.create).mockRejectedValue(error);

      await useScheduleStore.getState().createTask({
        title: 'New Task',
        prompt: 'Execute new task',
        scheduleType: 'daily',
        scheduleTime: '10:00',
        model: 'gpt-4',
        enabled: true,
        notifyOnDone: true,
        executionCount: 0,
      });

      expect(useScheduleStore.getState().error).toBe('Failed to create task');
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });
  });

  describe('updateTask', () => {
    it('should update task successfully', async () => {
      const updatedTask: ScheduledTask = {
        ...mockTasks[0],
        title: 'Updated Daily Report',
        prompt: 'Updated prompt',
      };

      useScheduleStore.setState({ tasks: mockTasks });
      vi.mocked(schedulesApi.update).mockResolvedValue(updatedTask);

      await useScheduleStore.getState().updateTask('task-1', {
        title: 'Updated Daily Report',
        prompt: 'Updated prompt',
      });

      expect(schedulesApi.update).toHaveBeenCalledWith('task-1', {
        title: 'Updated Daily Report',
        prompt: 'Updated prompt',
      });
      expect(useScheduleStore.getState().tasks[0]).toEqual(updatedTask);
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should handle update errors', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      const error = new Error('Failed to update task');
      vi.mocked(schedulesApi.update).mockRejectedValue(error);

      await useScheduleStore.getState().updateTask('task-1', {
        title: 'Updated',
      });

      expect(useScheduleStore.getState().error).toBe('Failed to update task');
      expect(useScheduleStore.getState().tasks[0].title).toBe('Daily Report'); // Not updated
    });
  });

  describe('deleteTask', () => {
    it('should delete task successfully', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      vi.mocked(schedulesApi.delete).mockResolvedValue(undefined);

      await useScheduleStore.getState().deleteTask('task-1');

      expect(schedulesApi.delete).toHaveBeenCalledWith('task-1');
      expect(useScheduleStore.getState().tasks).toHaveLength(1);
      expect(useScheduleStore.getState().tasks[0].id).toBe('task-2');
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should handle delete errors', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      const error = new Error('Failed to delete task');
      vi.mocked(schedulesApi.delete).mockRejectedValue(error);

      await useScheduleStore.getState().deleteTask('task-1');

      expect(useScheduleStore.getState().error).toBe('Failed to delete task');
      expect(useScheduleStore.getState().tasks).toHaveLength(2); // Not deleted
    });
  });

  describe('toggleTask with optimistic update', () => {
    it('should toggle task enabled state with optimistic update', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      vi.mocked(schedulesApi.toggle).mockResolvedValue(undefined);

      // Toggle from true to false
      await useScheduleStore.getState().toggleTask('task-1', false);

      expect(schedulesApi.toggle).toHaveBeenCalledWith('task-1', false);
      expect(useScheduleStore.getState().tasks[0].enabled).toBe(false);
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should revert optimistic update on failure', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      vi.mocked(schedulesApi.toggle).mockRejectedValue(new Error('Failed to toggle'));

      await useScheduleStore.getState().toggleTask('task-1', false);

      expect(useScheduleStore.getState().tasks[0].enabled).toBe(true); // Reverted
      expect(useScheduleStore.getState().error).toBe('Failed to toggle');
    });

    it('should update to enabled from disabled', async () => {
      useScheduleStore.setState({ tasks: mockTasks });
      vi.mocked(schedulesApi.toggle).mockResolvedValue(undefined);

      // Toggle from false to true
      await useScheduleStore.getState().toggleTask('task-2', true);

      expect(schedulesApi.toggle).toHaveBeenCalledWith('task-2', true);
      expect(useScheduleStore.getState().tasks[1].enabled).toBe(true);
    });
  });

  describe('loading state', () => {
    it('should set isLoading during operations', async () => {
      vi.mocked(schedulesApi.list).mockImplementation(() => new Promise(resolve =>
        setTimeout(() => resolve(mockTasks), 10)
      ));

      const promise = useScheduleStore.getState().fetchTasks();
      expect(useScheduleStore.getState().isLoading).toBe(true);
      await promise;
      expect(useScheduleStore.getState().isLoading).toBe(false);
    });

    it('should reset isLoading after error', async () => {
      vi.mocked(schedulesApi.list).mockRejectedValue(new Error('Error'));

      await useScheduleStore.getState().fetchTasks();

      expect(useScheduleStore.getState().isLoading).toBe(false);
    });
  });

  describe('initial state', () => {
    it('should initialize with default values', () => {
      const state = useScheduleStore.getState();

      expect(state.tasks).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe(null);
    });
  });
});