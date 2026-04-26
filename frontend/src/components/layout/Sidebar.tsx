'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Book, Zap, Package, Settings, Plus, MessageSquare, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useSessionsStore } from '@/stores/sessionsStore';
import { useEffect } from 'react';

const navItems = [
  { icon: Book, label: '知识', href: '/knowledge' },
  { icon: Zap, label: '技能', href: '/skills' },
  { icon: Package, label: '产物', href: '/artifacts' },
  { icon: Settings, label: '设置', href: '/settings' },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { sessions, fetchSessions, selectSession, deleteSession, selectedSession } = useSessionsStore();

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleNewChat = () => {
    selectSession(null);
    router.push('/');
  };

  const handleSelectSession = (session: any) => {
    selectSession(session);
    router.push(`/chat/${session.id}`);
  };

  return (
    <aside className="w-64 h-screen bg-white dark:bg-gray-900 border-r border-border flex flex-col shrink-0">
      {/* Logo */}
      <div className="p-4 border-b border-border">
        <Link href="/" className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-full bg-atom-core flex items-center justify-center">
            <span className="text-white text-xs font-bold">a</span>
          </div>
          <span className="text-lg font-bold text-atom-deep dark:text-atom-glow">atom.</span>
        </Link>
      </div>

      {/* New Chat Button */}
      <div className="p-3">
        <button
          onClick={handleNewChat}
          className="w-full flex items-center justify-center gap-2 bg-atom-core hover:bg-atom-nucleus text-white rounded-lg py-2.5 px-4 text-sm font-medium transition-colors"
        >
          <Plus className="w-4 h-4" />
          新的对话
        </button>
      </div>

      {/* Conversation History */}
      <div className="flex-1 overflow-y-auto px-2">
        <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground uppercase tracking-wider">
          我的对话
        </div>
        <div className="space-y-0.5">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={cn(
                'group flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer text-sm transition-colors',
                selectedSession?.id === session.id || pathname === `/chat/${session.id}`
                  ? 'bg-atom-mist text-atom-nucleus dark:bg-atom-deep/30 dark:text-atom-glow'
                  : 'text-foreground/70 hover:bg-muted'
              )}
              onClick={() => handleSelectSession(session)}
            >
              <MessageSquare className="w-4 h-4 shrink-0" />
              <span className="truncate flex-1">{session.title || '新对话'}</span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  deleteSession(session.id);
                }}
                className="opacity-0 group-hover:opacity-100 p-1 hover:text-destructive transition-opacity"
              >
                <Trash2 className="w-3 h-3" />
              </button>
            </div>
          ))}
          {sessions.length === 0 && (
            <div className="px-3 py-4 text-xs text-muted-foreground text-center">
              还没有对话，开始聊天吧
            </div>
          )}
        </div>
      </div>

      {/* Bottom Navigation */}
      <nav className="border-t border-border p-2 space-y-0.5">
        {navItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
                isActive
                  ? 'bg-atom-mist text-atom-nucleus dark:bg-atom-deep/30 dark:text-atom-glow'
                  : 'text-foreground/70 hover:bg-muted'
              )}
            >
              <item.icon className="w-4 h-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
