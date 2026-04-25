/**
 * Model and Provider Configuration Constants
 */

import { ModelConfig, ProviderConfig, ModelPreset, AgentConfig } from './types';

/**
 * Available AI Providers
 */
export const PROVIDERS: ProviderConfig[] = [
  {
    id: 'anthropic',
    name: 'Anthropic',
    baseURL: 'https://api.anthropic.com',
    envKey: 'ANTHROPIC_API_KEY',
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'openai',
    name: 'OpenAI',
    baseURL: 'https://api.openai.com/v1',
    envKey: 'OPENAI_API_KEY',
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'codex',
    name: 'Codex',
    baseURL: 'https://coder.api.visioncoder.cn/v1',
    envKey: 'CODEX_API_KEY',
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'kimi',
    name: 'Kimi',
    envKey: 'KIMI_API_KEY',
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    envKey: 'DEEPSEEK_API_KEY',
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'qwen',
    name: 'Qwen',
    envKey: 'QWEN_API_KEY',
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'ark',
    name: 'ARK',
    envKey: 'ARK_API_KEY',
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
];

/**
 * Available Models by Provider
 */
export const MODELS: ModelConfig[] = [
  // Anthropic Models
  {
    id: 'claude-sonnet-4-6',
    name: 'Claude Sonnet 4.6',
    provider: 'anthropic',
    description: 'Balanced performance for most tasks, optimized for coding',
    contextWindow: 200000,
    maxTokens: 8192,
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'claude-opus-4-6',
    name: 'Claude Opus 4.6',
    provider: 'anthropic',
    description: 'Most capable model for complex reasoning and creative tasks',
    contextWindow: 200000,
    maxTokens: 8192,
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'claude-haiku-4-6',
    name: 'Claude Haiku 4.6',
    provider: 'anthropic',
    description: 'Fast and efficient for simple tasks',
    contextWindow: 200000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  // OpenAI Models
  {
    id: 'gpt-4-turbo',
    name: 'GPT-4 Turbo',
    provider: 'openai',
    description: 'Advanced model with vision capabilities',
    contextWindow: 128000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'gpt-4',
    name: 'GPT-4',
    provider: 'openai',
    description: 'Powerful model for complex tasks',
    contextWindow: 8192,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: true, streaming: true },
  },
  {
    id: 'gpt-3.5-turbo',
    name: 'GPT-3.5 Turbo',
    provider: 'openai',
    description: 'Fast and cost-effective for simple tasks',
    contextWindow: 16385,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  // Codex Models
  {
    id: 'codex-o3',
    name: 'Codex O3',
    provider: 'codex',
    description: 'Advanced reasoning model',
    contextWindow: 128000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'codex-r1',
    name: 'Codex R1',
    provider: 'codex',
    description: 'Reasoning-focused model',
    contextWindow: 128000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  // Kimi Models
  {
    id: 'moonshot-v1-8k',
    name: 'Moonshot V1 8K',
    provider: 'kimi',
    description: 'Fast model for general tasks',
    contextWindow: 8000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'moonshot-v1-32k',
    name: 'Moonshot V1 32K',
    provider: 'kimi',
    description: 'Extended context for longer conversations',
    contextWindow: 32000,
    maxTokens: 8192,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  // DeepSeek Models
  {
    id: 'deepseek-chat',
    name: 'DeepSeek Chat',
    provider: 'deepseek',
    description: 'Chat-optimized model',
    contextWindow: 128000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'deepseek-coder',
    name: 'DeepSeek Coder',
    provider: 'deepseek',
    description: 'Code-optimized model',
    contextWindow: 128000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  // Qwen Models
  {
    id: 'qwen-turbo',
    name: 'Qwen Turbo',
    provider: 'qwen',
    description: 'Fast and efficient',
    contextWindow: 8000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'qwen-plus',
    name: 'Qwen Plus',
    provider: 'qwen',
    description: 'Balanced performance',
    contextWindow: 32000,
    maxTokens: 4096,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
  {
    id: 'qwen-max',
    name: 'Qwen Max',
    provider: 'qwen',
    description: 'Most capable Qwen model',
    contextWindow: 128000,
    maxTokens: 8192,
    capabilities: { toolUse: true, vision: false, streaming: true },
  },
];

/**
 * Model Presets for Quick Configuration
 */
export const MODEL_PRESETS: ModelPreset[] = [
  {
    id: 'balanced',
    name: 'Balanced',
    description: 'Good performance for most tasks (Claude Sonnet 4.6)',
    config: {
      model: 'claude-sonnet-4-6',
      provider: 'anthropic',
    },
  },
  {
    id: 'coding',
    name: 'Coding',
    description: 'Optimized for development tasks (DeepSeek Coder)',
    config: {
      model: 'deepseek-coder',
      provider: 'deepseek',
    },
  },
  {
    id: 'creative',
    name: 'Creative',
    description: 'Best for complex reasoning (Claude Opus 4.6)',
    config: {
      model: 'claude-opus-4-6',
      provider: 'anthropic',
    },
  },
  {
    id: 'fast',
    name: 'Fast',
    description: 'Quick responses for simple tasks (Claude Haiku 4.6)',
    config: {
      model: 'claude-haiku-4-6',
      provider: 'anthropic',
    },
  },
  {
    id: 'vision',
    name: 'Vision',
    description: 'Supports image analysis (GPT-4 Turbo)',
    config: {
      model: 'gpt-4-turbo',
      provider: 'openai',
    },
  },
];

/**
 * Helper Functions
 */

export function getProviderById(id: string): ProviderConfig | undefined {
  return PROVIDERS.find((p) => p.id === id);
}

export function getModelsByProvider(providerId: string): ModelConfig[] {
  return MODELS.filter((m) => m.provider === providerId);
}

export function getModelById(id: string): ModelConfig | undefined {
  return MODELS.find((m) => m.id === id);
}

export function getPresetById(id: string): ModelPreset | undefined {
  return MODEL_PRESETS.find((p) => p.id === id);
}

export function getDefaultModelForProvider(providerId: string): ModelConfig | undefined {
  const models = getModelsByProvider(providerId);
  return models[0];
}

export function validateModelProviderPair(modelId: string, providerId: string): boolean {
  const model = getModelById(modelId);
  return model?.provider === providerId;
}
