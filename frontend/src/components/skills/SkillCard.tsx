'use client';

import { Skill } from '@/types';
import { cn } from '@/lib/utils';

interface SkillCardProps {
  skill: Skill;
  onToggle: (id: string, enabled: boolean) => void;
}

export default function SkillCard({ skill, onToggle }: SkillCardProps) {
  const firstLetter = skill.name.charAt(0).toUpperCase();

  return (
    <div className="flex items-center gap-3 rounded-xl border border-border bg-card p-4">
      {/* Icon */}
      <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-atom-mist text-atom-deep font-semibold text-sm">
        {firstLetter}
      </div>

      {/* Content */}
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-foreground truncate">{skill.name}</p>
        <p className="text-xs text-muted-foreground line-clamp-2">{skill.description}</p>
      </div>

      {/* Toggle */}
      <button
        type="button"
        role="switch"
        aria-checked={skill.enabled}
        onClick={() => onToggle(skill.id, !skill.enabled)}
        className={cn(
          'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out',
          skill.enabled ? 'bg-green-500' : 'bg-gray-300'
        )}
      >
        <span
          className={cn(
            'pointer-events-none inline-block h-5 w-5 translate-y-0.5 rounded-full bg-white shadow ring-0 transition-transform duration-200 ease-in-out',
            skill.enabled ? 'translate-x-[22px]' : 'translate-x-[2px]'
          )}
        />
      </button>
    </div>
  );
}
