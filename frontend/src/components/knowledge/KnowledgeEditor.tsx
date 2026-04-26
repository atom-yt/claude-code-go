'use client';

import { cn } from '@/lib/utils';
import { Save } from 'lucide-react';

interface KnowledgeEditorProps {
  value: string;
  onChange: (value: string) => void;
  onSave?: () => void;
}

export function KnowledgeEditor({ value, onChange, onSave }: KnowledgeEditorProps) {
  return (
    <div className="flex flex-col h-full gap-3">
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="在这里编写 Markdown 格式的知识内容..."
        className={cn(
          'flex-1 w-full resize-none rounded-lg border border-border bg-background p-4',
          'font-mono text-sm leading-relaxed',
          'placeholder:text-muted-foreground',
          'focus:outline-none focus:ring-2 focus:ring-atom-core/30 focus:border-atom-core',
          'transition-colors'
        )}
      />
      {onSave && (
        <div className="flex justify-end">
          <button
            onClick={onSave}
            className={cn(
              'inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium',
              'bg-atom-core text-white hover:bg-atom-nucleus',
              'transition-colors'
            )}
          >
            <Save className="w-4 h-4" />
            保存知识
          </button>
        </div>
      )}
    </div>
  );
}
