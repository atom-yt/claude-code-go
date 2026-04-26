import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock axios with factory function
const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockDelete = vi.fn();
const mockRequest = vi.fn();
const mockRequestUse = vi.fn();
const mockResponseUse = vi.fn();

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      interceptors: {
        request: { use: mockRequestUse },
        response: { use: mockResponseUse },
      },
      get: mockGet,
      post: mockPost,
      put: mockPut,
      delete: mockDelete,
      request: mockRequest,
    })),
  },
}));

// Import after mock is set up
import { ApiClient, authApi, agentsApi, sessionsApi, messagesApi, chatApi, skillsApi, artifactsApi, knowledgeApi, schedulesApi, ChatStreamEvent } from '../api';

describe('ApiClient', () => {
  let apiClient: ApiClient;

  beforeEach(() => {
    vi.clearAllMocks();
    mockGet.mockClear();
    mockPost.mockClear();
    mockPut.mockClear();
    mockDelete.mockClear();
    mockRequestUse.mockClear();
    mockResponseUse.mockClear();
    apiClient = new ApiClient();
  });

  describe('Request Interceptor', () => {
    it('should add Bearer token to request headers when token exists in localStorage', async () => {
      const requestInterceptor = mockRequestUse.mock.calls[0][0];
      vi.mocked(localStorage.getItem).mockReturnValue('test-jwt-token');

      const config = { headers: {} };
      const result = requestInterceptor(config);

      expect(result.headers.Authorization).toBe('Bearer test-jwt-token');
    });

    it('should not add Bearer token when token does not exist in localStorage', async () => {
      const requestInterceptor = mockRequestUse.mock.calls[0][0];
      vi.mocked(localStorage.getItem).mockReturnValue(null);

      const config = { headers: {} };
      const result = requestInterceptor(config);

      expect(result.headers.Authorization).toBeUndefined();
    });
  });

  describe('Response Interceptor', () => {
    it('should clear token and redirect to /login on 401 error', async () => {
      const responseErrorInterceptor = mockResponseUse.mock.calls[1][1];
      const error = { response: { status: 401, data: { message: 'Unauthorized' } } };

      await expect(responseErrorInterceptor(error)).rejects.toHaveProperty('message');

      expect(localStorage.removeItem).toHaveBeenCalledWith('token');
      expect(window.location.href).toBe('/login');
    });

    it('should not redirect on non-401 errors', async () => {
      const responseErrorInterceptor = mockResponseUse.mock.calls[1][1];
      const error = { response: { status: 500, data: { message: 'Internal server error' } } };

      await expect(responseErrorInterceptor(error)).rejects.toHaveProperty('message');

      expect(localStorage.removeItem).not.toHaveBeenCalled();
      expect(window.location.href).not.toBe('/login');
    });
  });

  describe('HTTP Methods', () => {
    it('should make GET request and return data', async () => {
      const mockData = { id: '1', name: 'test' };
      mockGet.mockResolvedValue({ data: mockData });

      const result = await apiClient.get('/test');

      expect(result).toEqual(mockData);
      expect(mockGet).toHaveBeenCalledWith('/test', undefined);
    });

    it('should make GET request with config', async () => {
      const mockData = { id: '1', name: 'test' };
      const config = { params: { page: 1 } };
      mockGet.mockResolvedValue({ data: mockData });

      const result = await apiClient.get('/test', config);

      expect(result).toEqual(mockData);
      expect(mockGet).toHaveBeenCalledWith('/test', config);
    });

    it('should make POST request and return data', async () => {
      const mockData = { id: '1', name: 'test' };
      const payload = { name: 'test' };
      mockPost.mockResolvedValue({ data: mockData });

      const result = await apiClient.post('/test', payload);

      expect(result).toEqual(mockData);
      expect(mockPost).toHaveBeenCalledWith('/test', payload, undefined);
    });

    it('should make PUT request and return data', async () => {
      const mockData = { id: '1', name: 'updated' };
      const payload = { name: 'updated' };
      mockPut.mockResolvedValue({ data: mockData });

      const result = await apiClient.put('/test/1', payload);

      expect(result).toEqual(mockData);
      expect(mockPut).toHaveBeenCalledWith('/test/1', payload, undefined);
    });

    it('should make DELETE request and return data', async () => {
      const mockData = { success: true };
      mockDelete.mockResolvedValue({ data: mockData });

      const result = await apiClient.delete('/test/1');

      expect(result).toEqual(mockData);
      expect(mockDelete).toHaveBeenCalledWith('/test/1', undefined);
    });

    it('should handle GET request errors', async () => {
      const error = { response: { status: 404, data: { message: 'Not found' } } };
      mockGet.mockRejectedValue(error);

      await expect(apiClient.get('/test')).rejects.toHaveProperty('message', 'Not found');
    });

    it('should handle POST request errors', async () => {
      const error = { response: { status: 400, data: { message: 'Bad request' } } };
      mockPost.mockRejectedValue(error);

      await expect(apiClient.post('/test', {})).rejects.toHaveProperty('message', 'Bad request');
    });

    it('should handle PUT request errors', async () => {
      const error = { response: { status: 403, data: { message: 'Forbidden' } } };
      mockPut.mockRejectedValue(error);

      await expect(apiClient.put('/test/1', {})).rejects.toHaveProperty('message', 'Forbidden');
    });

    it('should handle DELETE request errors', async () => {
      const error = { response: { status: 500, data: { message: 'Server error' } } };
      mockDelete.mockRejectedValue(error);

      await expect(apiClient.delete('/test/1')).rejects.toHaveProperty('message', 'Server error');
    });
  });

  describe('Timeout Handling', () => {
    it('should handle timeout errors', async () => {
      const error = { code: 'ECONNABORTED', message: 'timeout of 30000ms exceeded' };
      mockGet.mockRejectedValue(error);

      await expect(apiClient.get('/test')).rejects.toHaveProperty('message');
    });
  });

  describe('Network Error Handling', () => {
    it('should handle network errors (no response)', async () => {
      const error = { request: {}, message: 'Network Error' };
      mockGet.mockRejectedValue(error);

      const result = await apiClient.get('/test').catch(e => e);

      expect(result).toEqual({
        message: 'No response from server',
        status: 0,
        code: 'NETWORK_ERROR',
      });
    });
  });
});

describe('authApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { post: vi.fn(), get: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should call login endpoint with correct payload', async () => {
    const mockResponse = { user: { id: '1', email: 'test@example.com' }, access_token: 'token' };
    apiClient.post.mockResolvedValue(mockResponse);

    const result = await authApi.login('test@example.com', 'password');

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/auth/login', {
      email: 'test@example.com',
      password: 'password',
    });
    expect(result).toEqual(mockResponse);
  });

  it('should call register endpoint with correct payload', async () => {
    const mockResponse = { user: { id: '1', email: 'test@example.com' }, access_token: 'token' };
    apiClient.post.mockResolvedValue(mockResponse);

    const result = await authApi.register('test@example.com', 'password', 'Test User');

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/auth/register', {
      email: 'test@example.com',
      password: 'password',
      display_name: 'Test User',
    });
    expect(result).toEqual(mockResponse);
  });

  it('should call logout endpoint', async () => {
    apiClient.post.mockResolvedValue(undefined);

    await authApi.logout();

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/auth/logout');
  });

  it('should call me endpoint', async () => {
    const mockResponse = { user: { id: '1', email: 'test@example.com' } };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await authApi.me();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/auth/me');
    expect(result).toEqual(mockResponse);
  });
});

describe('agentsApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list agents and extract items', async () => {
    const mockResponse = { items: [{ id: '1', name: 'Agent 1' }], total: 1 };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await agentsApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/agents');
    expect(result).toEqual([{ id: '1', name: 'Agent 1' }]);
  });

  it('should get a single agent by id', async () => {
    const mockAgent = { id: '1', name: 'Agent 1' };
    apiClient.get.mockResolvedValue(mockAgent);

    const result = await agentsApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/agents/1');
    expect(result).toEqual(mockAgent);
  });

  it('should create a new agent', async () => {
    const mockAgent = { id: '1', name: 'New Agent' };
    const data = { name: 'New Agent', description: 'Test' };
    apiClient.post.mockResolvedValue(mockAgent);

    const result = await agentsApi.create(data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/agents', data);
    expect(result).toEqual(mockAgent);
  });

  it('should update an existing agent', async () => {
    const mockAgent = { id: '1', name: 'Updated Agent' };
    const data = { name: 'Updated Agent' };
    apiClient.put.mockResolvedValue(mockAgent);

    const result = await agentsApi.update('1', data);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/agents/1', data);
    expect(result).toEqual(mockAgent);
  });

  it('should delete an agent', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await agentsApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/agents/1');
  });
});

describe('sessionsApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list sessions without agent filter', async () => {
    const mockResponse = { items: [{ id: '1', title: 'Session 1' }], total: 1 };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await sessionsApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions', { params: {} });
    expect(result).toEqual([{ id: '1', title: 'Session 1' }]);
  });

  it('should list sessions with agent filter', async () => {
    const mockResponse = { items: [{ id: '1', title: 'Session 1' }], total: 1 };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await sessionsApi.list('agent-1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions', { params: { agent_id: 'agent-1' } });
    expect(result).toEqual([{ id: '1', title: 'Session 1' }]);
  });

  it('should get a single session by id', async () => {
    const mockSession = { id: '1', title: 'Session 1' };
    apiClient.get.mockResolvedValue(mockSession);

    const result = await sessionsApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions/1');
    expect(result).toEqual(mockSession);
  });

  it('should create a session with title and agentId', async () => {
    const mockSession = { id: '1', title: 'New Session', agent_id: 'agent-1' };
    apiClient.post.mockResolvedValue(mockSession);

    const result = await sessionsApi.create({ title: 'New Session', agentId: 'agent-1' });

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/sessions', {
      title: 'New Session',
      agent_id: 'agent-1',
    });
    expect(result).toEqual(mockSession);
  });

  it('should create a session with only title', async () => {
    const mockSession = { id: '1', title: 'New Session' };
    apiClient.post.mockResolvedValue(mockSession);

    const result = await sessionsApi.create({ title: 'New Session' });

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/sessions', {
      title: 'New Session',
    });
    expect(result).toEqual(mockSession);
  });

  it('should delete a session', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await sessionsApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/sessions/1');
  });
});

describe('messagesApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should get messages for a session with pagination', async () => {
    const mockResponse = {
      items: [{ id: '1', content: 'Message 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await messagesApi.get('session-1', 2, 25);

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions/session-1/messages', {
      params: { page: 2, page_size: 25 },
    });
    expect(result).toEqual([{ id: '1', content: 'Message 1' }]);
  });

  it('should get messages with default pagination', async () => {
    const mockResponse = {
      items: [{ id: '1', content: 'Message 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await messagesApi.get('session-1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions/session-1/messages', {
      params: { page: 1, page_size: 50 },
    });
    expect(result).toEqual([{ id: '1', content: 'Message 1' }]);
  });

  it('should get recent messages for a session', async () => {
    const mockMessages = [{ id: '1', content: 'Message 1' }];
    apiClient.get.mockResolvedValue(mockMessages);

    const result = await messagesApi.getRecent('session-1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sessions/session-1/messages/recent');
    expect(result).toEqual(mockMessages);
  });

  it('should create a message', async () => {
    const mockMessage = { id: '1', content: 'New message' };
    const data = { role: 'user', content: 'New message' };
    apiClient.post.mockResolvedValue(mockMessage);

    const result = await messagesApi.create('session-1', data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/sessions/session-1/messages', data);
    expect(result).toEqual(mockMessage);
  });

  it('should delete a message', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await messagesApi.delete('message-1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/messages/message-1');
  });
});

describe('chatApi', () => {
  let originalFetch: typeof fetch;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    originalFetch = global.fetch;
    mockFetch = vi.fn();
    global.fetch = mockFetch;
    vi.mocked(localStorage.getItem).mockReturnValue('test-token');
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('should send a non-streaming message', async () => {
    const apiClient = { post: vi.fn() };
    (require('../api') as any).api = apiClient;

    const mockMessage = { id: '1', content: 'Response' };
    apiClient.post.mockResolvedValue(mockMessage);

    const result = await chatApi.send('session-1', 'Hello');

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/chat', {
      session_id: 'session-1',
      message: 'Hello',
      stream: false,
    });
    expect(result).toEqual(mockMessage);
  });

  it('should parse SSE content events correctly', async () => {
    const mockStream = new ReadableStream({
      start(controller) {
        const encoder = new TextEncoder();
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"Hello "}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"world"}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"done"}\n\n'));
        controller.close();
      },
    });

    mockFetch.mockResolvedValue({
      ok: true,
      body: mockStream,
      json: vi.fn(),
    });

    const events: ChatStreamEvent[] = [];
    await chatApi.stream('session-1', 'Hello', (event) => {
      events.push(event);
    });

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/chat/stream',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token',
        }),
        body: JSON.stringify({
          session_id: 'session-1',
          message: 'Hello',
          stream: true,
        }),
      })
    );

    expect(events).toEqual([
      { type: 'content', content: 'Hello ' },
      { type: 'content', content: 'world' },
      { type: 'done' },
    ]);
  });

  it('should parse SSE error events correctly', async () => {
    const mockStream = new ReadableStream({
      start(controller) {
        const encoder = new TextEncoder();
        controller.enqueue(encoder.encode('data: {"type":"error","error":"Something went wrong"}\n\n'));
        controller.close();
      },
    });

    mockFetch.mockResolvedValue({
      ok: true,
      body: mockStream,
      json: vi.fn(),
    });

    const events: ChatStreamEvent[] = [];
    await chatApi.stream('session-1', 'Hello', (event) => {
      events.push(event);
    });

    expect(events).toEqual([
      { type: 'error', error: 'Something went wrong' },
    ]);
  });

  it('should handle [DONE] sentinel in SSE stream', async () => {
    const mockStream = new ReadableStream({
      start(controller) {
        const encoder = new TextEncoder();
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"Hello"}\n\n'));
        controller.enqueue(encoder.encode('data: [DONE]\n\n'));
        controller.close();
      },
    });

    mockFetch.mockResolvedValue({
      ok: true,
      body: mockStream,
      json: vi.fn(),
    });

    const events: ChatStreamEvent[] = [];
    await chatApi.stream('session-1', 'Hello', (event) => {
      events.push(event);
    });

    expect(events).toEqual([
      { type: 'content', content: 'Hello' },
    ]);
  });

  it('should handle malformed JSON in SSE stream', async () => {
    const mockStream = new ReadableStream({
      start(controller) {
        const encoder = new TextEncoder();
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"Hello"}\n\n'));
        controller.enqueue(encoder.encode('data: {invalid json}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"World"}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"done"}\n\n'));
        controller.close();
      },
    });

    mockFetch.mockResolvedValue({
      ok: true,
      body: mockStream,
      json: vi.fn(),
    });

    const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    const events: ChatStreamEvent[] = [];
    await chatApi.stream('session-1', 'Hello', (event) => {
      events.push(event);
    });

    expect(consoleWarnSpy).toHaveBeenCalledWith('Failed to parse SSE event:', '{invalid json}');
    expect(events).toEqual([
      { type: 'content', content: 'Hello' },
      { type: 'content', content: 'World' },
      { type: 'done' },
    ]);

    consoleWarnSpy.mockRestore();
  });

  it('should handle SSE stream errors', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: vi.fn().mockResolvedValue({ message: 'Internal server error' }),
    });

    await expect(chatApi.stream('session-1', 'Hello', () => {})).rejects.toThrow(
      'Internal server error'
    );
  });

  it('should handle SSE stream errors without json response', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: vi.fn().mockRejectedValue(new Error('Invalid JSON')),
    });

    await expect(chatApi.stream('session-1', 'Hello', () => {})).rejects.toThrow(
      'Chat request failed'
    );
  });

  it('should handle missing response body', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      body: null,
      json: vi.fn(),
    });

    await expect(chatApi.stream('session-1', 'Hello', () => {})).rejects.toThrow(
      'No response body reader'
    );
  });

  it('should stop streaming after done event', async () => {
    const mockStream = new ReadableStream({
      start(controller) {
        const encoder = new TextEncoder();
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"Hello"}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"done"}\n\n'));
        controller.enqueue(encoder.encode('data: {"type":"delta","text":"World"}\n\n'));
        controller.close();
      },
    });

    mockFetch.mockResolvedValue({
      ok: true,
      body: mockStream,
      json: vi.fn(),
    });

    const events: ChatStreamEvent[] = [];
    await chatApi.stream('session-1', 'Hello', (event) => {
      events.push(event);
    });

    expect(events).toEqual([
      { type: 'content', content: 'Hello' },
      { type: 'done' },
    ]);
  });
});

describe('skillsApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list skills without category filter', async () => {
    const mockResponse = {
      items: [{ id: '1', name: 'Skill 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await skillsApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/skills', { params: {} });
    expect(result).toEqual([{ id: '1', name: 'Skill 1' }]);
  });

  it('should list skills with category filter', async () => {
    const mockResponse = {
      items: [{ id: '1', name: 'Skill 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await skillsApi.list('personal');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/skills', { params: { category: 'personal' } });
    expect(result).toEqual([{ id: '1', name: 'Skill 1' }]);
  });

  it('should get a single skill by id', async () => {
    const mockSkill = { id: '1', name: 'Skill 1' };
    apiClient.get.mockResolvedValue(mockSkill);

    const result = await skillsApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/skills/1');
    expect(result).toEqual(mockSkill);
  });

  it('should create a new skill', async () => {
    const mockSkill = { id: '1', name: 'New Skill' };
    const data = { name: 'New Skill', description: 'Test' };
    apiClient.post.mockResolvedValue(mockSkill);

    const result = await skillsApi.create(data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/skills', data);
    expect(result).toEqual(mockSkill);
  });

  it('should update an existing skill', async () => {
    const mockSkill = { id: '1', name: 'Updated Skill' };
    const data = { name: 'Updated Skill' };
    apiClient.put.mockResolvedValue(mockSkill);

    const result = await skillsApi.update('1', data);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/skills/1', data);
    expect(result).toEqual(mockSkill);
  });

  it('should delete a skill', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await skillsApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/skills/1');
  });

  it('should toggle a skill', async () => {
    apiClient.put.mockResolvedValue(undefined);

    await skillsApi.toggle('1', true);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/skills/1/toggle', { enabled: true });
  });
});

describe('artifactsApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list artifacts without search', async () => {
    const mockResponse = {
      items: [{ id: '1', title: 'Artifact 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await artifactsApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/artifacts', {
      params: { page: 1, page_size: 20 },
    });
    expect(result).toEqual(mockResponse);
  });

  it('should list artifacts with search', async () => {
    const mockResponse = {
      items: [{ id: '1', title: 'Artifact 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await artifactsApi.list('test', 2, 10);

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/artifacts', {
      params: { search: 'test', page: 2, page_size: 10 },
    });
    expect(result).toEqual(mockResponse);
  });

  it('should get a single artifact by id', async () => {
    const mockArtifact = { id: '1', title: 'Artifact 1' };
    apiClient.get.mockResolvedValue(mockArtifact);

    const result = await artifactsApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/artifacts/1');
    expect(result).toEqual(mockArtifact);
  });

  it('should create a new artifact', async () => {
    const mockArtifact = { id: '1', title: 'New Artifact' };
    const data = { title: 'New Artifact', content: 'Test content' };
    apiClient.post.mockResolvedValue(mockArtifact);

    const result = await artifactsApi.create(data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/artifacts', data);
    expect(result).toEqual(mockArtifact);
  });

  it('should delete an artifact', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await artifactsApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/artifacts/1');
  });

  it('should get artifact stats', async () => {
    const mockStats = { total: 42 };
    apiClient.get.mockResolvedValue(mockStats);

    const result = await artifactsApi.stats();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/artifacts/stats');
    expect(result).toEqual(mockStats);
  });
});

describe('knowledgeApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list knowledge items', async () => {
    const mockResponse = {
      items: [{ id: '1', name: 'Knowledge 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await knowledgeApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/knowledge');
    expect(result).toEqual([{ id: '1', name: 'Knowledge 1' }]);
  });

  it('should get a single knowledge item by id', async () => {
    const mockKnowledge = { id: '1', name: 'Knowledge 1' };
    apiClient.get.mockResolvedValue(mockKnowledge);

    const result = await knowledgeApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/knowledge/1');
    expect(result).toEqual(mockKnowledge);
  });

  it('should create a new knowledge item', async () => {
    const mockKnowledge = { id: '1', name: 'New Knowledge' };
    const data = { name: 'New Knowledge', description: 'Test' };
    apiClient.post.mockResolvedValue(mockKnowledge);

    const result = await knowledgeApi.create(data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/knowledge', data);
    expect(result).toEqual(mockKnowledge);
  });

  it('should update an existing knowledge item', async () => {
    const mockKnowledge = { id: '1', name: 'Updated Knowledge' };
    const data = { name: 'Updated Knowledge' };
    apiClient.put.mockResolvedValue(mockKnowledge);

    const result = await knowledgeApi.update('1', data);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/knowledge/1', data);
    expect(result).toEqual(mockKnowledge);
  });

  it('should delete a knowledge item', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await knowledgeApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/knowledge/1');
  });
});

describe('schedulesApi', () => {
  let apiClient: any;

  beforeEach(() => {
    apiClient = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() };
    (require('../api') as any).api = apiClient;
  });

  it('should list scheduled tasks', async () => {
    const mockResponse = {
      items: [{ id: '1', title: 'Schedule 1' }],
      total: 1,
    };
    apiClient.get.mockResolvedValue(mockResponse);

    const result = await schedulesApi.list();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedules');
    expect(result).toEqual([{ id: '1', title: 'Schedule 1' }]);
  });

  it('should get a single schedule by id', async () => {
    const mockSchedule = { id: '1', title: 'Schedule 1' };
    apiClient.get.mockResolvedValue(mockSchedule);

    const result = await schedulesApi.get('1');

    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedules/1');
    expect(result).toEqual(mockSchedule);
  });

  it('should create a new schedule', async () => {
    const mockSchedule = { id: '1', title: 'New Schedule' };
    const data = { title: 'New Schedule', prompt: 'Test prompt' };
    apiClient.post.mockResolvedValue(mockSchedule);

    const result = await schedulesApi.create(data);

    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/schedules', data);
    expect(result).toEqual(mockSchedule);
  });

  it('should update an existing schedule', async () => {
    const mockSchedule = { id: '1', title: 'Updated Schedule' };
    const data = { title: 'Updated Schedule' };
    apiClient.put.mockResolvedValue(mockSchedule);

    const result = await schedulesApi.update('1', data);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/schedules/1', data);
    expect(result).toEqual(mockSchedule);
  });

  it('should delete a schedule', async () => {
    apiClient.delete.mockResolvedValue(undefined);

    await schedulesApi.delete('1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/schedules/1');
  });

  it('should toggle a schedule', async () => {
    apiClient.put.mockResolvedValue(undefined);

    await schedulesApi.toggle('1', true);

    expect(apiClient.put).toHaveBeenCalledWith('/api/v1/schedules/1/toggle', { enabled: true });
  });
});
