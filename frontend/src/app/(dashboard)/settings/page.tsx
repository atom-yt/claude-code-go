'use client';

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { BasicSettings } from '@/components/settings/BasicSettings';
import { ScheduleList } from '@/components/schedule/ScheduleList';
import { Settings, CalendarClock, Palette } from 'lucide-react';

type TabKey = 'basic' | 'schedule' | 'personalization';

const tabs: { key: TabKey; label: string; icon: React.ElementType }[] = [
  { key: 'basic', label: '基础设置', icon: Settings },
  { key: 'schedule', label: '定时服务', icon: CalendarClock },
  { key: 'personalization', label: '个性化配置', icon: Palette },
];

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('basic');

  return (
    <div className="flex flex-col h-full p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-foreground">设置</h1>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 mb-6">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={cn(
                'inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors',
                activeTab === tab.key
                  ? 'bg-atom-core text-white'
                  : 'bg-secondary text-muted-foreground hover:text-foreground hover:bg-secondary/80'
              )}
            >
              <Icon className="w-4 h-4" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <div className="flex-1 min-h-0 overflow-y-auto">
        {activeTab === 'basic' && <BasicSettings />}
        {activeTab === 'schedule' && <ScheduleList />}
        {activeTab === 'personalization' && (
          <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground rounded-lg border border-dashed border-border">
            <Palette className="w-12 h-12 opacity-20" />
            <p className="text-sm">个性化配置功能即将上线</p>
            <p className="text-xs">敬请期待</p>
          </div>
        )}
      </div>
    </div>
  );
}
