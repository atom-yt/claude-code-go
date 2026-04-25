// User types
export interface User {
  id: string;
  email: string;
  username: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthToken {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType: string;
}

export interface AuthResponse {
  user: User;
  token: AuthToken;
}

// Agent types
export interface Agent {
  id: string;
  userId: string;
  name: string;
  description: string;
  systemPrompt: string;
  tools: string[];
  config?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

// Session types
export interface Session {
  id: string;
  userId: string;
  agentId: string;        // 关联的 Agent ID
  title: string;
  status: 'active' | 'archived' | 'deleted';  // Session 状态
  createdAt: string;
  updatedAt: string;
}

// Message types
export interface Message {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant' | 'tool';
  content: string;
  toolCalls?: ToolCall[];
  createdAt: string;
}

export interface ToolCall {
  id: string;
  name: string;
  input: Record<string, any>;
  output?: string;
  isError?: boolean;
}

// API request/response types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface CreateSessionRequest {
  title?: string;
  agentId: string;        // 必填：关联的 Agent ID
}

export interface SendMessageRequest {
  content: string;
  stream?: boolean;
}

export interface CreateAgentRequest {
  name: string;
  description: string;
  systemPrompt: string;
  tools: string[];
  config?: Record<string, any>;
}

export interface ErrorResponse {
  error: string;
  message: string;
  code: number;
}

export interface HealthResponse {
  status: string;
  version: string;
  database: string;
}

export interface ApiError {
  message: string;
  status: number;
  code?: string;
}

// Store types
export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface AgentsState {
  agents: Agent[];
  selectedAgent: Agent | null;
  isLoading: boolean;
  error: string | null;
}

export interface SessionsState {
  sessions: Session[];
  selectedSession: Session | null;
  messages: Message[];
  isLoading: boolean;
  error: string | null;
}

export interface ChatState {
  isStreaming: boolean;
  currentContent: string;
  toolCalls: ToolCall[];
}