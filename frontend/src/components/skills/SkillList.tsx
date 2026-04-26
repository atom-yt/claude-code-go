'use client';

import { useEffect } from 'react';
import { Plus, Zap } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useSkillsStore } from '@/stores/skillsStore';
import SkillCard from './SkillCard';

const TABS = [
  { key: 'personal', label: '自己加的' },
  { key: 'team', label: '团队加的' },
  { key: 'builtin', label: 'atom 自带的' },
] as const;

export default function SkillList() {
  const { skills, category, isLoading, fetchSkills, toggleSkill, setCategory } =
    useSkillsStore();

  useEffect(() => {
    fetchSkills(category || undefined);
  }, []);

  const filteredSkills = category
    ? skills.filter((s) => s.category === category)
    : skills;

  return (
    <div className="mx-auto max-w-3xl px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-bold text-foreground">技能</h1>
          <span className="inline-flex items-center rounded-full bg-atom-mist px-2.5 py-0.5 text-xs font-medium text-atom-deep">
            {skills.length}
          </span>
        </div>
        <button
          className="inline-flex items-center gap-1.5 rounded-lg bg-black px-4 py-2 text-sm font-medium text-white hover:bg-black/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          新技能
        </button>
      </div>

      {/* Description */}
      <p className="mt-2 text-sm text-muted-foreground">
        统一管理赋予 atom 更强大的能力
      </p>

      {/* Tabs */}
      <div className="mt-6 flex items-center gap-2">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setCategory(category === tab.key ? '' : tab.key)}
            className={cn(
              'rounded-lg px-3 py-1.5 text-sm font-medium transition-colors',
              category === tab.key
                ? 'bg-atom-mist text-atom-deep'
                : 'text-muted-foreground hover:bg-muted'
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Grid */}
      {isLoading ? (
        <div className="mt-8 flex items-center justify-center py-12 text-muted-foreground text-sm">
          加载中...
        </div>
      ) : filteredSkills.length === 0 ? (
        <div className="mt-8 flex flex-col items-center justify-center py-16 text-center">
          <Zap className="h-10 w-10 text-muted-foreground/40" />
          <p className="mt-3 text-sm text-muted-foreground">暂无技能</p>
          <p className="mt-1 text-xs text-muted-foreground/60">
            点击右上角「+ 新技能」添加
          </p>
        </div>
      ) : (
        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
          {filteredSkills.map((skill) => (
            <SkillCard key={skill.id} skill={skill} onToggle={toggleSkill} />
          ))}
        </div>
      )}
    </div>
  );
}
