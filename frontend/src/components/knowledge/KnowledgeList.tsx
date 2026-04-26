'use client';

import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';
import { useKnowledgeStore } from '@/stores/knowledgeStore';
import { KnowledgeEditor } from './KnowledgeEditor';
import { Knowledge } from '@/types';
import { Plus, BookOpen, FileText, Trash2 } from 'lucide-react';

type TabKey = 'user' | 'ark';

export function KnowledgeList() {
  const {
    items,
    selectedItem,
    isLoading,
    error,
    fetchKnowledge,
    createKnowledge,
    updateKnowledge,
    deleteKnowledge,
    setSelectedItem,
  } = useKnowledgeStore();

  const [activeTab, setActiveTab] = useState<TabKey>('user');
  const [editorContent, setEditorContent] = useState('');

  useEffect(() => {
    fetchKnowledge();
  }, [fetchKnowledge]);

  useEffect(() => {
    if (selectedItem) {
      setEditorContent(selectedItem.description || '');
    } else {
      setEditorContent('');
    }
  }, [selectedItem]);

  const filteredItems = items.filter((item) => item.source === activeTab);

  const tabs: { key: TabKey; label: string }[] = [
    { key: 'user', label: '你告诉 atom 的' },
    { key: 'ark', label: '来自知识方舟的' },
  ];

  const handleSave = () => {
    if (selectedItem) {
      updateKnowledge(selectedItem.id, { description: editorContent });
    }
  };

  const handleCreate = () => {
    createKnowledge({
      name: '新知识',
      description: '',
      type: 'markdown',
      source: activeTab,
    });
  };

  const handleDelete = (e: React.MouseEvent, item: Knowledge) => {
    e.stopPropagation();
    deleteKnowledge(item.id);
  };

  return (
    <div className="flex flex-col h-full p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-foreground">知识</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          在这里为 atom 灌输知识，让 atom 更符合你的心意
        </p>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 mb-4">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => {
              setActiveTab(tab.key);
              setSelectedItem(null);
            }}
            className={cn(
              'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
              activeTab === tab.key
                ? 'bg-atom-core text-white'
                : 'bg-secondary text-muted-foreground hover:text-foreground hover:bg-secondary/80'
            )}
          >
            {tab.label}
          </button>
        ))}
        <div className="flex-1" />
        <button
          onClick={handleCreate}
          className={cn(
            'inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg',
            'border border-dashed border-atom-core/40 text-atom-core',
            'hover:bg-atom-mist transition-colors'
          )}
        >
          <Plus className="w-4 h-4" />
          添加知识
        </button>
      </div>

      {/* Content area */}
      <div className="flex flex-1 gap-4 min-h-0">
        {/* Left: list */}
        <div className="w-72 shrink-0 overflow-y-auto rounded-lg border border-border bg-card">
          {isLoading && filteredItems.length === 0 ? (
            <div className="flex items-center justify-center h-32 text-sm text-muted-foreground">
              加载中...
            </div>
          ) : filteredItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-48 gap-2 text-muted-foreground">
              <BookOpen className="w-10 h-10 opacity-30" />
              <p className="text-sm">暂无知识</p>
              <p className="text-xs">点击右上角添加</p>
            </div>
          ) : (
            <ul className="divide-y divide-border">
              {filteredItems.map((item) => (
                <li
                  key={item.id}
                  onClick={() => setSelectedItem(item)}
                  className={cn(
                    'flex items-center gap-3 px-4 py-3 cursor-pointer transition-colors',
                    selectedItem?.id === item.id
                      ? 'bg-atom-mist'
                      : 'hover:bg-secondary/60'
                  )}
                >
                  <FileText className="w-4 h-4 shrink-0 text-atom-core" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{item.name}</p>
                    <p className="text-xs text-muted-foreground truncate">
                      {item.description?.slice(0, 40) || '无内容'}
                    </p>
                  </div>
                  <button
                    onClick={(e) => handleDelete(e, item)}
                    className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive transition-all"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Right: editor */}
        <div className="flex-1 min-w-0">
          {selectedItem ? (
            <KnowledgeEditor
              value={editorContent}
              onChange={setEditorContent}
              onSave={handleSave}
            />
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2 rounded-lg border border-dashed border-border">
              <BookOpen className="w-12 h-12 opacity-20" />
              <p className="text-sm">选择左侧的知识条目进行编辑</p>
            </div>
          )}
        </div>
      </div>

      {/* Error display */}
      {error && (
        <div className="mt-4 px-4 py-2 text-sm text-destructive bg-destructive/10 rounded-lg">
          {error}
        </div>
      )}
    </div>
  );
}
