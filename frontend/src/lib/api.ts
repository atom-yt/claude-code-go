import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { AuthResponse, ApiError, Agent, Session, Message, Skill, Artifact, Knowledge, ScheduledTask } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// 分页响应类型
interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        if (typeof window !== 'undefined') {
          const token = localStorage.getItem('token');
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor to handle auth errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Clear token and redirect to login
          if (typeof window !== 'undefined') {
            localStorage.removeItem('token');
            window.location.href = '/login';
          }
        }
        return Promise.reject(this.handleError(error));
      }
    );
  }

  private handleError(error: any): ApiError {
    if (error.response) {
      return {
        message: error.response.data?.message || 'An error occurred',
        status: error.response.status,
        code: error.response.data?.code,
      };
    } else if (error.request) {
      return {
        message: 'No response from server',
        status: 0,
        code: 'NETWORK_ERROR',
      };
    } else {
      return {
        message: error.message || 'An unknown error occurred',
        status: 0,
        code: 'UNKNOWN_ERROR',
      };
    }
  }

  async request<T>(config: AxiosRequestConfig): Promise<AxiosResponse<T>> {
    return this.client.request<T>(config);
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }
}

export const api = new ApiClient();

// Auth API
export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    return api.post('/api/v1/auth/login', { email, password });
  },

  register: async (email: string, password: string, name: string): Promise<AuthResponse> => {
    return api.post('/api/v1/auth/register', { email, password, name });
  },

  logout: async (): Promise<void> => {
    return api.post('/api/v1/auth/logout');
  },

  me: async (): Promise<AuthResponse> => {
    return api.get('/api/v1/auth/me');
  },
};

// Agents API
export const agentsApi = {
  list: async (): Promise<Agent[]> => {
    const response = await api.get<ListResponse<Agent>>('/api/v1/agents');
    return response.items;
  },

  get: async (id: string): Promise<Agent> => {
    return api.get(`/api/v1/agents/${id}`);
  },

  create: async (data: any): Promise<Agent> => {
    return api.post('/api/v1/agents', data);
  },

  update: async (id: string, data: any): Promise<Agent> => {
    return api.put(`/api/v1/agents/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/agents/${id}`);
  },
};

// Sessions API
export const sessionsApi = {
  list: async (agentId?: string): Promise<Session[]> => {
    const params = agentId ? { agentId } : {};
    const response = await api.get<ListResponse<Session>>('/api/v1/sessions', { params });
    return response.items;
  },

  get: async (id: string): Promise<Session> => {
    return api.get(`/api/v1/sessions/${id}`);
  },

  create: async (data: any): Promise<Session> => {
    return api.post('/api/v1/sessions', data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/sessions/${id}`);
  },
};

// Messages API
export const messagesApi = {
  get: async (sessionId: string, page = 1, pageSize = 50): Promise<Message[]> => {
    const response = await api.get<ListResponse<Message>>(`/api/v1/sessions/${sessionId}/messages`, {
      params: { page, page_size: pageSize }
    });
    return response.items;
  },

  getRecent: async (sessionId: string): Promise<Message[]> => {
    const response = await api.get<Message[]>(`/api/v1/sessions/${sessionId}/messages/recent`);
    return response;
  },

  create: async (sessionId: string, data: any): Promise<Message> => {
    return api.post(`/api/v1/sessions/${sessionId}/messages`, data);
  },

  delete: async (messageId: string): Promise<void> => {
    return api.delete(`/api/v1/messages/${messageId}`);
  },
};

// Chat API (streaming)
export interface ChatStreamEvent {
  type: 'content' | 'tool_call' | 'tool_result' | 'error' | 'done';
  content?: string;
  toolCall?: {
    id: string;
    name: string;
    input: Record<string, any>;
  };
  toolResult?: {
    toolCallId: string;
    output: string;
    isError?: boolean;
  };
  error?: string;
}

export const chatApi = {
  stream: async (
    sessionId: string,
    message: string,
    onEvent: (event: ChatStreamEvent) => void
  ): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/v1/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${typeof window !== 'undefined' ? localStorage.getItem('token') : ''}`,
      },
      body: JSON.stringify({ session_id: sessionId, message, stream: true }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || 'Chat request failed');
    }

    const reader = response.body?.getReader();
    const decoder = new TextDecoder();

    if (!reader) {
      throw new Error('No response body reader');
    }

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6).trim();
            if (!data || data === '[DONE]') continue;

            try {
              const event = JSON.parse(data) as ChatStreamEvent;
              onEvent(event);

              if (event.type === 'done') {
                return;
              }
            } catch (e) {
              console.warn('Failed to parse SSE event:', data);
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  },

  send: async (sessionId: string, message: string): Promise<Message> => {
    return api.post(`/api/v1/chat`, { session_id: sessionId, message, stream: false });
  },
};

// Skills API
export const skillsApi = {
  list: async (category?: string): Promise<Skill[]> => {
    const params: Record<string, string> = {};
    if (category) params.category = category;
    const response = await api.get<ListResponse<Skill>>('/api/v1/skills', { params });
    return response.items;
  },

  get: async (id: string): Promise<Skill> => {
    return api.get(`/api/v1/skills/${id}`);
  },

  create: async (data: any): Promise<Skill> => {
    return api.post('/api/v1/skills', data);
  },

  update: async (id: string, data: any): Promise<Skill> => {
    return api.put(`/api/v1/skills/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/skills/${id}`);
  },

  toggle: async (id: string, enabled: boolean): Promise<void> => {
    return api.put(`/api/v1/skills/${id}/toggle`, { enabled });
  },
};

// Artifacts API
export const artifactsApi = {
  list: async (search?: string, page = 1, pageSize = 20): Promise<{ items: Artifact[]; total: number }> => {
    const params: Record<string, any> = { page, page_size: pageSize };
    if (search) params.search = search;
    return api.get('/api/v1/artifacts', { params });
  },

  get: async (id: string): Promise<Artifact> => {
    return api.get(`/api/v1/artifacts/${id}`);
  },

  create: async (data: any): Promise<Artifact> => {
    return api.post('/api/v1/artifacts', data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/artifacts/${id}`);
  },

  stats: async (): Promise<{ total: number }> => {
    return api.get('/api/v1/artifacts/stats');
  },
};

// Knowledge API
export const knowledgeApi = {
  list: async (): Promise<Knowledge[]> => {
    const response = await api.get<ListResponse<Knowledge>>('/api/v1/knowledge');
    return response.items;
  },

  get: async (id: string): Promise<Knowledge> => {
    return api.get(`/api/v1/knowledge/${id}`);
  },

  create: async (data: any): Promise<Knowledge> => {
    return api.post('/api/v1/knowledge', data);
  },

  update: async (id: string, data: any): Promise<Knowledge> => {
    return api.put(`/api/v1/knowledge/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/knowledge/${id}`);
  },
};

// Schedules API
export const schedulesApi = {
  list: async (): Promise<ScheduledTask[]> => {
    const response = await api.get<ListResponse<ScheduledTask>>('/api/v1/schedules');
    return response.items;
  },

  get: async (id: string): Promise<ScheduledTask> => {
    return api.get(`/api/v1/schedules/${id}`);
  },

  create: async (data: any): Promise<ScheduledTask> => {
    return api.post('/api/v1/schedules', data);
  },

  update: async (id: string, data: any): Promise<ScheduledTask> => {
    return api.put(`/api/v1/schedules/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/schedules/${id}`);
  },

  toggle: async (id: string, enabled: boolean): Promise<void> => {
    return api.put(`/api/v1/schedules/${id}/toggle`, { enabled });
  },
};