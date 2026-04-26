'use client';

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { X } from 'lucide-react';

interface AddScheduleDialogProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: {
    title: string;
    prompt: string;
    scheduleType: 'daily' | 'weekly' | 'cron';
    scheduleTime: string;
    model: string;
    notifyOnDone: boolean;
  }) => void;
}

export function AddScheduleDialog({ open, onClose, onSubmit }: AddScheduleDialogProps) {
  const [title, setTitle] = useState('');
  const [prompt, setPrompt] = useState('');
  const [scheduleType, setScheduleType] = useState<'daily' | 'weekly' | 'cron'>('daily');
  const [scheduleTime, setScheduleTime] = useState('09:00');
  const [model, setModel] = useState('auto');
  const [notifyOnDone, setNotifyOnDone] = useState(true);

  if (!open) return null;

  const handleSubmit = () => {
    if (!title.trim()) return;
    onSubmit({
      title: title.trim(),
      prompt: prompt.trim(),
      scheduleType,
      scheduleTime,
      model,
      notifyOnDone,
    });
    // Reset form
    setTitle('');
    setPrompt('');
    setScheduleType('daily');
    setScheduleTime('09:00');
    setModel('auto');
    setNotifyOnDone(true);
  };

  const labelClass = 'block text-sm font-medium text-foreground mb-1.5';
  const inputClass = cn(
    'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm',
    'placeholder:text-muted-foreground',
    'focus:outline-none focus:ring-2 focus:ring-atom-core/30 focus:border-atom-core',
    'transition-colors'
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />

      {/* Dialog */}
      <div className="relative w-full max-w-lg mx-4 bg-card rounded-xl border border-border shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold">创建定时任务</h2>
          <button
            onClick={onClose}
            className="p-1 rounded-lg hover:bg-secondary transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4 max-h-[60vh] overflow-y-auto">
          {/* Title */}
          <div>
            <label className={labelClass}>做什么（标题）</label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value.slice(0, 30))}
              placeholder="告诉我我要做啥"
              maxLength={30}
              className={inputClass}
            />
            <p className="mt-1 text-xs text-muted-foreground text-right">
              {title.length}/30
            </p>
          </div>

          {/* Prompt */}
          <div>
            <label className={labelClass}>怎么做（提示词）</label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value.slice(0, 5000))}
              placeholder="越详细越好"
              maxLength={5000}
              rows={4}
              className={cn(inputClass, 'resize-none')}
            />
            <p className="mt-1 text-xs text-muted-foreground text-right">
              {prompt.length}/5000
            </p>
          </div>

          {/* Schedule type */}
          <div>
            <label className={labelClass}>啥种任务（类型）</label>
            <select
              value={scheduleType}
              onChange={(e) => setScheduleType(e.target.value as any)}
              className={inputClass}
            >
              <option value="daily">每天</option>
              <option value="weekly">每周</option>
              <option value="cron">自定义 Cron</option>
            </select>
          </div>

          {/* Schedule time */}
          <div>
            <label className={labelClass}>啥时候做（执行时间）</label>
            <input
              type="time"
              value={scheduleTime}
              onChange={(e) => setScheduleTime(e.target.value)}
              className={inputClass}
            />
          </div>

          {/* Model */}
          <div>
            <label className={labelClass}>用啥做（模型）</label>
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className={inputClass}
            >
              <option value="auto">Auto · 自动选择</option>
              <option value="claude-sonnet-4">Claude Sonnet 4</option>
              <option value="gpt-4o">GPT-4o</option>
            </select>
          </div>

          {/* Notify toggle */}
          <div className="flex items-center justify-between">
            <label className="text-sm font-medium text-foreground">完成通知</label>
            <button
              role="switch"
              aria-checked={notifyOnDone}
              onClick={() => setNotifyOnDone(!notifyOnDone)}
              className={cn(
                'relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors',
                notifyOnDone ? 'bg-atom-core' : 'bg-muted'
              )}
            >
              <span
                className={cn(
                  'pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow-sm transition-transform mt-0.5',
                  notifyOnDone ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'
                )}
              />
            </button>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-border">
          <button
            onClick={onClose}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-lg',
              'border border-border text-foreground hover:bg-secondary',
              'transition-colors'
            )}
          >
            取消
          </button>
          <button
            onClick={handleSubmit}
            disabled={!title.trim()}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-lg',
              'bg-atom-core text-white hover:bg-atom-nucleus',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'transition-colors'
            )}
          >
            创建
          </button>
        </div>
      </div>
    </div>
  );
}
