'use client';

import { useCallback, useState } from 'react';
import { MessageList } from './MessageList';
import { ChatInput } from './ChatInput';
import { Message } from '@/types';
import { chatApi, ChatStreamEvent } from '@/lib/api';
import { useSessionsStore } from '@/stores/sessionsStore';

interface ChatProps {
  sessionId: string;
}

export function Chat({ sessionId }: ChatProps) {
  const { messages, addMessage } = useSessionsStore();
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamContent, setStreamContent] = useState('');

  const handleSend = useCallback(async (content: string) => {
    // Add user message
    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      sessionId,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };
    addMessage(userMessage);

    setIsStreaming(true);
    setStreamContent('');

    let accumulated = '';

    try {
      await chatApi.stream(sessionId, content, (event: ChatStreamEvent) => {
        if (event.type === 'content' && event.content) {
          accumulated += event.content;
          setStreamContent(accumulated);
        } else if (event.type === 'done') {
          const assistantMessage: Message = {
            id: `msg-${Date.now() + 1}`,
            sessionId,
            role: 'assistant',
            content: accumulated,
            createdAt: new Date().toISOString(),
          };
          addMessage(assistantMessage);
          setIsStreaming(false);
          setStreamContent('');
        } else if (event.type === 'error') {
          const errorMessage: Message = {
            id: `msg-${Date.now() + 1}`,
            sessionId,
            role: 'assistant',
            content: `错误: ${event.error || '请求失败'}`,
            createdAt: new Date().toISOString(),
          };
          addMessage(errorMessage);
          setIsStreaming(false);
          setStreamContent('');
        }
      });
    } catch {
      setIsStreaming(false);
      setStreamContent('');
      const errorMessage: Message = {
        id: `msg-${Date.now() + 1}`,
        sessionId,
        role: 'assistant',
        content: '连接已断开，请重试',
        createdAt: new Date().toISOString(),
      };
      addMessage(errorMessage);
    }
  }, [sessionId, addMessage]);

  // Build display messages including streaming content
  const displayMessages = [...messages];
  if (isStreaming && streamContent) {
    displayMessages.push({
      id: 'streaming',
      sessionId,
      role: 'assistant',
      content: streamContent,
      createdAt: new Date().toISOString(),
    });
  }

  return (
    <div className="flex flex-col h-full">
      <MessageList messages={displayMessages} isStreaming={isStreaming && !streamContent} />
      <div className="border-t border-border p-4">
        <ChatInput
          onSend={handleSend}
          disabled={isStreaming}
          placeholder="继续对话..."
        />
      </div>
    </div>
  );
}
