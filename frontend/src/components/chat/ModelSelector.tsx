'use client';

import { useState } from 'react';
import { ChevronDown } from 'lucide-react';

const models = [
  { id: 'claude-sonnet-4', name: 'Claude Sonnet 4', provider: 'Anthropic' },
  { id: 'claude-opus-4', name: 'Claude Opus 4', provider: 'Anthropic' },
  { id: 'gpt-4o', name: 'GPT-4o', provider: 'OpenAI' },
  { id: 'deepseek-chat', name: 'DeepSeek Chat', provider: 'DeepSeek' },
  { id: 'qwen-max', name: 'Qwen Max', provider: 'Qwen' },
  { id: 'moonshot-v1-128k', name: 'Moonshot 128K', provider: 'Kimi' },
];

interface ModelSelectorProps {
  value: string;
  onChange: (model: string) => void;
}

export function ModelSelector({ value, onChange }: ModelSelectorProps) {
  const [open, setOpen] = useState(false);
  const selected = models.find((m) => m.id === value) || models[0];

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded-md hover:bg-muted"
      >
        {selected.name}
        <ChevronDown className="w-3 h-3" />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute bottom-full left-0 mb-1 bg-popover border border-border rounded-lg shadow-lg py-1 z-50 min-w-[200px]">
            {models.map((model) => (
              <button
                key={model.id}
                onClick={() => {
                  onChange(model.id);
                  setOpen(false);
                }}
                className={`w-full text-left px-3 py-2 text-sm hover:bg-muted transition-colors ${
                  model.id === value ? 'text-atom-core font-medium' : 'text-foreground'
                }`}
              >
                <div>{model.name}</div>
                <div className="text-xs text-muted-foreground">{model.provider}</div>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
