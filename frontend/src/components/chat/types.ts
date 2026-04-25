// Streaming event types from backend
export type StreamEventType = 'delta' | 'tool_call' | 'tool_result' | 'error' | 'done';

export interface StreamEvent {
  type: StreamEventType;
  data: {
    content?: string;
    toolCall?: {
      id: string;
      name: string;
      input: Record<string, any>;
    };
    toolResult?: {
      id: string;
      output: string;
      isError: boolean;
    };
    error?: {
      message: string;
      code?: string;
    };
  };
}

export interface ChatConnectionType {
  type: 'sse' | 'websocket';
}

export interface ChatMessage extends Message {
  isStreaming?: boolean;
  toolCallsDisplay?: ToolCall[];
}