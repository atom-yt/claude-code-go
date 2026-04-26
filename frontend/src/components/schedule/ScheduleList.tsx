'use client';

import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';
import { useScheduleStore } from '@/stores/scheduleStore';
import { ScheduleCard } from './ScheduleCard';
import { AddScheduleDialog } from './AddScheduleDialog';
import { Plus, CalendarClock } from 'lucide-react';

export function ScheduleList() {
  const { tasks, isLoading, error, fetchTasks, createTask, toggleTask } = useScheduleStore();
  const [dialogOpen, setDialogOpen] = useState(false);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleSubmit = async (data: any) => {
    await createTask(data);
    setDialogOpen(false);
  };

  return (
    <div className="flex flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">定时任务</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            在指定时间自动执行预设的 AI 任务
          </p>
        </div>
        <button
          onClick={() => setDialogOpen(true)}
          className={cn(
            'inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg',
            'bg-atom-core text-white hover:bg-atom-nucleus',
            'transition-colors'
          )}
        >
          <Plus className="w-4 h-4" />
          添加任务
        </button>
      </div>

      {/* Task list */}
      {isLoading && tasks.length === 0 ? (
        <div className="flex items-center justify-center h-32 text-sm text-muted-foreground">
          加载中...
        </div>
      ) : tasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground rounded-lg border border-dashed border-border">
          <CalendarClock className="w-12 h-12 opacity-20" />
          <p className="text-sm">暂无定时任务</p>
          <p className="text-xs">点击上方按钮创建你的第一个定时任务</p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          {tasks.map((task) => (
            <ScheduleCard key={task.id} task={task} onToggle={toggleTask} />
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="px-4 py-2 text-sm text-destructive bg-destructive/10 rounded-lg">
          {error}
        </div>
      )}

      {/* Dialog */}
      <AddScheduleDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
