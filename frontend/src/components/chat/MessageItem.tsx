'use client';

import { cn } from '@/lib/utils';
import { User, Bot, ChevronDown, ChevronRight } from 'lucide-react';
import { Message, ToolCall } from '@/types';
import { useState } from 'react';

interface MessageItemProps {
  message: Message;
}

function ToolCallBlock({ toolCall }: { toolCall: ToolCall }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="mt-2 border border-border rounded-lg overflow-hidden text-xs">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-3 py-2 bg-muted/50 hover:bg-muted transition-colors"
      >
        {expanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
        <span className="font-mono font-medium text-atom-nucleus">{toolCall.name}</span>
        {toolCall.isError && <span className="text-destructive ml-auto">错误</span>}
      </button>
      {expanded && (
        <div className="p-3 space-y-2">
          <div>
            <div className="text-muted-foreground mb-1">输入</div>
            <pre className="bg-muted p-2 rounded text-xs overflow-x-auto">
              {JSON.stringify(toolCall.input, null, 2)}
            </pre>
          </div>
          {toolCall.output && (
            <div>
              <div className="text-muted-foreground mb-1">输出</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-x-auto max-h-40 overflow-y-auto">
                {toolCall.output}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function MessageItem({ message }: MessageItemProps) {
  const isUser = message.role === 'user';

  return (
    <div className={cn('flex gap-3 px-4 py-3', isUser ? 'flex-row-reverse' : '')}>
      <div
        className={cn(
          'w-8 h-8 rounded-full flex items-center justify-center shrink-0',
          isUser ? 'bg-foreground/10' : 'bg-atom-core'
        )}
      >
        {isUser ? (
          <User className="w-4 h-4 text-foreground/60" />
        ) : (
          <Bot className="w-4 h-4 text-white" />
        )}
      </div>
      <div className={cn('max-w-[75%] min-w-0', isUser ? 'text-right' : '')}>
        <div
          className={cn(
            'inline-block rounded-2xl px-4 py-2.5 text-sm',
            isUser
              ? 'bg-atom-core text-white rounded-br-md'
              : 'bg-muted rounded-bl-md'
          )}
        >
          <div className="whitespace-pre-wrap break-words">{message.content}</div>
        </div>
        {message.toolCalls && message.toolCalls.length > 0 && (
          <div className="mt-1 text-left">
            {message.toolCalls.map((tc) => (
              <ToolCallBlock key={tc.id} toolCall={tc} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
