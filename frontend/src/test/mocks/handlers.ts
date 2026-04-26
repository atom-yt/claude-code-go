import { http, HttpResponse } from 'msw';

// Helper function to check authentication
function isAuthenticated(request: Request): boolean {
  const authHeader = request.headers.get('authorization');
  return authHeader === 'Bearer mock-jwt-token';
}

// Mock data
const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  username: 'testuser',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
};

const mockAgents = [
  {
    id: '00000000-0000-0000-0000-000000000001',
    userId: 'user-1',
    name: 'Default Agent',
    description: 'Default system agent',
    systemPrompt: 'You are a helpful assistant.',
    tools: ['Read', 'Write', 'Grep', 'Bash'],
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '00000000-0000-0000-0000-000000000002',
    userId: 'user-1',
    name: 'Code Agent',
    description: 'Specialized for code assistance',
    systemPrompt: 'You are a coding assistant.',
    tools: ['Read', 'Write', 'Grep', 'Bash', 'Glob'],
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const mockSessions = [
  {
    id: 'session-1',
    userId: 'user-1',
    agentId: '00000000-0000-0000-0000-000000000001',
    title: 'Test Session 1',
    status: 'active' as const,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'session-2',
    userId: 'user-1',
    agentId: '00000000-0000-0000-0000-000000000002',
    title: 'Test Session 2',
    status: 'active' as const,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const mockMessages = [
  {
    id: 'msg-1',
    sessionId: 'session-1',
    role: 'user' as const,
    content: 'Hello, how are you?',
    createdAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'msg-2',
    sessionId: 'session-1',
    role: 'assistant' as const,
    content: 'I am doing well, thank you!',
    createdAt: '2024-01-01T00:01:00Z',
  },
];

const mockSkills = [
  {
    id: 'skill-1',
    userId: 'user-1',
    name: 'Web Search',
    description: 'Search the web for information',
    category: 'personal' as const,
    icon: 'search',
    enabled: true,
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 'skill-2',
    userId: 'user-1',
    name: 'Code Analysis',
    description: 'Analyze code patterns',
    category: 'personal' as const,
    icon: 'code',
    enabled: false,
    config: {},
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const mockArtifacts = [
  {
    id: 'artifact-1',
    userId: 'user-1',
    sessionId: 'session-1',
    title: 'Test Artifact',
    content: 'Artifact content here',
    fileType: 'text/plain',
    tags: ['test', 'example'],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const mockKnowledge = [
  {
    id: 'knowledge-1',
    userId: 'user-1',
    name: 'Project Documentation',
    description: 'Documentation for the project',
    type: 'file',
    source: 'user' as const,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

const mockSchedules = [
  {
    id: 'schedule-1',
    userId: 'user-1',
    title: 'Daily Standup',
    prompt: 'Generate a daily standup summary',
    scheduleType: 'daily' as const,
    scheduleTime: '09:00',
    model: 'claude-3-5-sonnet-20241022',
    enabled: true,
    notifyOnDone: true,
    executionCount: 5,
    lastRunAt: '2024-01-01T09:00:00Z',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// Mock handlers for all API endpoints
export const handlers = [
  // Auth endpoints
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string };

    if (body.email === 'test@example.com' && body.password === 'password') {
      return HttpResponse.json({
        user: mockUser,
        access_token: 'mock-jwt-token',
        refresh_token: 'mock-refresh-token',
      });
    }

    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    );
  }),

  http.post('/api/v1/auth/register', async ({ request }) => {
    const body = await request.json() as { email: string; password: string; display_name: string };

    return HttpResponse.json({
      user: {
        ...mockUser,
        id: 'new-user',
        email: body.email,
        username: body.display_name,
      },
      access_token: 'mock-jwt-token',
      refresh_token: 'mock-refresh-token',
    });
  }),

  http.post('/api/v1/auth/logout', () => {
    return HttpResponse.json({ success: true });
  }),

  http.get('/api/v1/auth/me', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({ user: mockUser });
  }),

  // Agents endpoints
  http.get('/api/v1/agents', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockAgents,
      total: mockAgents.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/agents/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const agent = mockAgents.find((a) => a.id === params.id);
    if (!agent) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(agent);
  }),

  http.post('/api/v1/agents', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newAgent = {
      id: 'new-agent-id',
      userId: 'user-1',
      ...body,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockAgents.push(newAgent);
    return HttpResponse.json(newAgent);
  }),

  http.put('/api/v1/agents/:id', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const index = mockAgents.findIndex((a) => a.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockAgents[index] = { ...mockAgents[index], ...body, updatedAt: new Date().toISOString() };
    return HttpResponse.json(mockAgents[index]);
  }),

  http.delete('/api/v1/agents/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockAgents.findIndex((a) => a.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockAgents.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  // Sessions endpoints
  http.get('/api/v1/sessions', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockSessions,
      total: mockSessions.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/sessions/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const session = mockSessions.find((s) => s.id === params.id);
    if (!session) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(session);
  }),

  http.post('/api/v1/sessions', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newSession = {
      id: 'new-session-id',
      userId: 'user-1',
      ...body,
      status: 'active',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockSessions.unshift(newSession);
    return HttpResponse.json(newSession);
  }),

  http.delete('/api/v1/sessions/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockSessions.findIndex((s) => s.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockSessions.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  // Messages endpoints
  http.get('/api/v1/sessions/:sessionId/messages', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const messages = mockMessages.filter((m) => m.sessionId === params.sessionId);
    return HttpResponse.json({
      items: messages,
      total: messages.length,
      page: 1,
      page_size: 50,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/sessions/:sessionId/messages/recent', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const messages = mockMessages.filter((m) => m.sessionId === params.sessionId);
    return HttpResponse.json(messages);
  }),

  http.post('/api/v1/sessions/:sessionId/messages', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newMessage = {
      id: 'new-msg-id',
      sessionId: params.sessionId,
      ...body,
      createdAt: new Date().toISOString(),
    };
    mockMessages.push(newMessage);
    return HttpResponse.json(newMessage);
  }),

  http.delete('/api/v1/messages/:messageId', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockMessages.findIndex((m) => m.id === params.messageId);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockMessages.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  // Skills endpoints
  http.get('/api/v1/skills', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockSkills,
      total: mockSkills.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/skills/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const skill = mockSkills.find((s) => s.id === params.id);
    if (!skill) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(skill);
  }),

  http.post('/api/v1/skills', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newSkill = {
      id: 'new-skill-id',
      userId: 'user-1',
      ...body,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockSkills.push(newSkill);
    return HttpResponse.json(newSkill);
  }),

  http.put('/api/v1/skills/:id', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const index = mockSkills.findIndex((s) => s.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockSkills[index] = { ...mockSkills[index], ...body, updatedAt: new Date().toISOString() };
    return HttpResponse.json(mockSkills[index]);
  }),

  http.delete('/api/v1/skills/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockSkills.findIndex((s) => s.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockSkills.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  http.put('/api/v1/skills/:id/toggle', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json() as { enabled: boolean };
    const skill = mockSkills.find((s) => s.id === params.id);
    if (!skill) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    skill.enabled = body.enabled;
    return HttpResponse.json({ success: true });
  }),

  // Artifacts endpoints
  http.get('/api/v1/artifacts', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockArtifacts,
      total: mockArtifacts.length,
    });
  }),

  http.get('/api/v1/artifacts/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const artifact = mockArtifacts.find((a) => a.id === params.id);
    if (!artifact) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(artifact);
  }),

  http.post('/api/v1/artifacts', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newArtifact = {
      id: 'new-artifact-id',
      userId: 'user-1',
      ...body,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockArtifacts.push(newArtifact);
    return HttpResponse.json(newArtifact);
  }),

  http.delete('/api/v1/artifacts/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockArtifacts.findIndex((a) => a.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockArtifacts.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  http.get('/api/v1/artifacts/stats', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({ total: mockArtifacts.length });
  }),

  // Knowledge endpoints
  http.get('/api/v1/knowledge', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockKnowledge,
      total: mockKnowledge.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/knowledge/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const knowledge = mockKnowledge.find((k) => k.id === params.id);
    if (!knowledge) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(knowledge);
  }),

  http.post('/api/v1/knowledge', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newKnowledge = {
      id: 'new-knowledge-id',
      userId: 'user-1',
      ...body,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockKnowledge.push(newKnowledge);
    return HttpResponse.json(newKnowledge);
  }),

  http.put('/api/v1/knowledge/:id', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const index = mockKnowledge.findIndex((k) => k.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockKnowledge[index] = { ...mockKnowledge[index], ...body, updatedAt: new Date().toISOString() };
    return HttpResponse.json(mockKnowledge[index]);
  }),

  http.delete('/api/v1/knowledge/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockKnowledge.findIndex((k) => k.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockKnowledge.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  // Schedules endpoints
  http.get('/api/v1/schedules', ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    return HttpResponse.json({
      items: mockSchedules,
      total: mockSchedules.length,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });
  }),

  http.get('/api/v1/schedules/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const schedule = mockSchedules.find((s) => s.id === params.id);
    if (!schedule) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    return HttpResponse.json(schedule);
  }),

  http.post('/api/v1/schedules', async ({ request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const newSchedule = {
      id: 'new-schedule-id',
      userId: 'user-1',
      ...body,
      executionCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mockSchedules.push(newSchedule);
    return HttpResponse.json(newSchedule);
  }),

  http.put('/api/v1/schedules/:id', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json();
    const index = mockSchedules.findIndex((s) => s.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockSchedules[index] = { ...mockSchedules[index], ...body, updatedAt: new Date().toISOString() };
    return HttpResponse.json(mockSchedules[index]);
  }),

  http.delete('/api/v1/schedules/:id', ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const index = mockSchedules.findIndex((s) => s.id === params.id);
    if (index === -1) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    mockSchedules.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  http.put('/api/v1/schedules/:id/toggle', async ({ params, request }) => {
    if (!isAuthenticated(request)) {
      return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }
    const body = await request.json() as { enabled: boolean };
    const schedule = mockSchedules.find((s) => s.id === params.id);
    if (!schedule) {
      return HttpResponse.json({ error: 'Not found' }, { status: 404 });
    }
    schedule.enabled = body.enabled;
    return HttpResponse.json({ success: true });
  }),
];