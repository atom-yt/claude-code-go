'use client';

import { useState, useRef, KeyboardEvent } from 'react';
import { ArrowUp } from 'lucide-react';
import { ModelSelector } from './ModelSelector';

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export function ChatInput({ onSend, disabled, placeholder }: ChatInputProps) {
  const [message, setMessage] = useState('');
  const [model, setModel] = useState('claude-sonnet-4');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = () => {
    const trimmed = message.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setMessage('');
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleInput = () => {
    const el = textareaRef.current;
    if (el) {
      el.style.height = 'auto';
      el.style.height = Math.min(el.scrollHeight, 200) + 'px';
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto">
      <div className="relative border border-border rounded-2xl bg-background shadow-sm focus-within:border-atom-spark focus-within:ring-1 focus-within:ring-atom-spark/30 transition-all">
        <textarea
          ref={textareaRef}
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyDown={handleKeyDown}
          onInput={handleInput}
          placeholder={placeholder || '有什么我能帮忙的？试试输入 / 召唤技能'}
          disabled={disabled}
          rows={1}
          className="w-full resize-none bg-transparent px-4 pt-3 pb-10 text-sm outline-none placeholder:text-muted-foreground disabled:opacity-50"
        />
        <div className="absolute bottom-2 left-2 right-2 flex items-center justify-between">
          <ModelSelector value={model} onChange={setModel} />
          <button
            onClick={handleSend}
            disabled={!message.trim() || disabled}
            className="p-1.5 rounded-lg bg-atom-core text-white hover:bg-atom-nucleus disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ArrowUp className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
