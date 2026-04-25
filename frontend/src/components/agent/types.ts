/**
 * Agent Configuration Types
 */

export interface AgentConfig {
  model?: string;
  provider?: string;
  api_key?: string;
  base_url?: string;
  system_prompt?: string;
}

export interface ModelConfig {
  id: string;
  name: string;
  provider: string;
  description: string;
  contextWindow?: number;
  maxTokens?: number;
  capabilities?: {
    toolUse?: boolean;
    vision?: boolean;
    streaming?: boolean;
  };
}

export interface ProviderConfig {
  id: string;
  name: string;
  baseURL?: string;
  envKey: string;
  capabilities: {
    toolUse: boolean;
    vision: boolean;
    streaming: boolean;
  };
}

export interface ModelPreset {
  id: string;
  name: string;
  description: string;
  config: AgentConfig;
}

export interface FormErrors {
  model?: string;
  provider?: string;
  api_key?: string;
  base_url?: string;
  system_prompt?: string;
}

export interface AgentConfigFormProps {
  initialConfig?: AgentConfig;
  onSave: (config: AgentConfig) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
}

export interface ModelSelectProps {
  value?: string;
  onChange: (modelId: string) => void;
  provider?: string;
  disabled?: boolean;
}

export interface ProviderSelectProps {
  value?: string;
  onChange: (providerId: string) => void;
  disabled?: boolean;
}
