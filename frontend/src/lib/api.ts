import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { AuthResponse, ApiError } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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
  list: async (): Promise<any[]> => {
    return api.get('/api/v1/agents');
  },

  get: async (id: string): Promise<any> => {
    return api.get(`/api/v1/agents/${id}`);
  },

  create: async (data: any): Promise<any> => {
    return api.post('/api/v1/agents', data);
  },

  update: async (id: string, data: any): Promise<any> => {
    return api.put(`/api/v1/agents/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/agents/${id}`);
  },
};

// Sessions API
export const sessionsApi = {
  list: async (agentId?: string): Promise<any[]> => {
    const params = agentId ? { agentId } : {};
    return api.get('/api/v1/sessions', { params });
  },

  get: async (id: string): Promise<any> => {
    return api.get(`/api/v1/sessions/${id}`);
  },

  create: async (data: any): Promise<any> => {
    return api.post('/api/v1/sessions', data);
  },

  delete: async (id: string): Promise<void> => {
    return api.delete(`/api/v1/sessions/${id}`);
  },
};