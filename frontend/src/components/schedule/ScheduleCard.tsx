'use client';

import { cn } from '@/lib/utils';
import { ScheduledTask } from '@/types';
import { Timer } from 'lucide-react';

interface ScheduleCardProps {
  task: ScheduledTask;
  onToggle: (id: string, enabled: boolean) => void;
}

function formatSchedule(task: ScheduledTask): string {
  switch (task.scheduleType) {
    case 'daily':
      return `每天 ${task.scheduleTime}`;
    case 'weekly':
      return `每周 ${task.scheduleTime}`;
    case 'cron':
      return `Cron: ${task.scheduleTime}`;
    default:
      return task.scheduleTime;
  }
}

export function ScheduleCard({ task, onToggle }: ScheduleCardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border border-border bg-card p-4 transition-colors',
        task.enabled ? 'opacity-100' : 'opacity-60'
      )}
    >
      {/* Top row: icon + title + schedule + toggle */}
      <div className="flex items-start gap-3">
        <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-atom-mist text-atom-core">
          <Timer className="w-4 h-4" />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium truncate">{task.title}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {formatSchedule(task)}
          </p>
        </div>
        {/* Toggle switch */}
        <button
          role="switch"
          aria-checked={task.enabled}
          onClick={() => onToggle(task.id, !task.enabled)}
          className={cn(
            'relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors',
            task.enabled ? 'bg-atom-core' : 'bg-muted'
          )}
        >
          <span
            className={cn(
              'pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow-sm transition-transform mt-0.5',
              task.enabled ? 'translate-x-4 ml-0.5' : 'translate-x-0.5'
            )}
          />
        </button>
      </div>

      {/* Prompt preview */}
      {task.prompt && (
        <p className="mt-2 text-xs text-muted-foreground line-clamp-2 pl-11">
          {task.prompt}
        </p>
      )}

      {/* Footer: execution count */}
      <div className="flex justify-end mt-3">
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-secondary text-muted-foreground">
          已执行 {task.executionCount} 次
        </span>
      </div>
    </div>
  );
}
