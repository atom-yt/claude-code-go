'use client';

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Save } from 'lucide-react';

export function BasicSettings() {
  const [displayName, setDisplayName] = useState('');
  const [language, setLanguage] = useState('zh');
  const [theme, setTheme] = useState('light');

  const labelClass = 'block text-sm font-medium text-foreground mb-1.5';
  const inputClass = cn(
    'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm',
    'placeholder:text-muted-foreground',
    'focus:outline-none focus:ring-2 focus:ring-atom-core/30 focus:border-atom-core',
    'transition-colors'
  );

  const handleSave = () => {
    // TODO: persist settings
    console.log('Save settings:', { displayName, language, theme });
  };

  return (
    <div className="max-w-lg space-y-6">
      {/* Display name */}
      <div>
        <label className={labelClass}>显示名称</label>
        <input
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="你的名字"
          className={inputClass}
        />
      </div>

      {/* Language */}
      <div>
        <label className={labelClass}>语言</label>
        <select
          value={language}
          onChange={(e) => setLanguage(e.target.value)}
          className={inputClass}
        >
          <option value="zh">中文</option>
          <option value="en">English</option>
        </select>
      </div>

      {/* Theme */}
      <div>
        <label className={labelClass}>主题</label>
        <select
          value={theme}
          onChange={(e) => setTheme(e.target.value)}
          className={inputClass}
        >
          <option value="light">浅色</option>
          <option value="dark">深色</option>
        </select>
      </div>

      {/* Save */}
      <div>
        <button
          onClick={handleSave}
          className={cn(
            'inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium',
            'bg-atom-core text-white hover:bg-atom-nucleus',
            'transition-colors'
          )}
        >
          <Save className="w-4 h-4" />
          保存设置
        </button>
      </div>
    </div>
  );
}
