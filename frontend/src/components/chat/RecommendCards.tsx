'use client';

import { useState } from 'react';
import { RefreshCw } from 'lucide-react';

interface RecommendCard {
  title: string;
  description: string;
  prompt: string;
}

const allCards: RecommendCard[] = [
  { title: '邮箱整理', description: '查询我今日有哪些邮件还未回复', prompt: '请帮我查询今日未回复的邮件，并按优先级排序' },
  { title: '做定时任务', description: '每早九点给我总结昨天科技圈最火热消息', prompt: '帮我创建一个定时任务：每天早上9点总结昨天科技圈最火热的消息' },
  { title: '趋势查询', description: '用内外搜工具查询 OpenClaw 动态', prompt: '请使用搜索工具查询 OpenClaw 项目的最新动态和技术趋势' },
  { title: '会议总结', description: '对昨天的会议进行总结并整理 todo', prompt: '请帮我总结昨天的会议内容，并整理出待办事项列表' },
  { title: '代码审查', description: '帮我 Review 最近提交的 PR 代码变更', prompt: '请帮我审查最近提交的 Pull Request 代码变更，指出潜在问题' },
  { title: '技术调研', description: '调研 AI Agent 领域最新的开源项目', prompt: '请帮我调研 AI Agent 领域最新的开源项目，列出 Top 10 并分析各自特点' },
  { title: '文档生成', description: '根据代码自动生成 API 文档', prompt: '请帮我根据项目代码自动生成 API 接口文档' },
  { title: '运维巡检', description: '检查服务健康状态和资源使用情况', prompt: '请帮我执行运维巡检，检查各服务的健康状态和资源使用情况' },
];

interface RecommendCardsProps {
  onSelect: (prompt: string) => void;
}

export function RecommendCards({ onSelect }: RecommendCardsProps) {
  const [startIndex, setStartIndex] = useState(0);

  const visibleCards = [];
  for (let i = 0; i < 4; i++) {
    visibleCards.push(allCards[(startIndex + i) % allCards.length]);
  }

  const handleRefresh = () => {
    setStartIndex((prev) => (prev + 4) % allCards.length);
  };

  return (
    <div className="w-full max-w-2xl">
      <div className="grid grid-cols-2 gap-3">
        {visibleCards.map((card, index) => (
          <button
            key={`${startIndex}-${index}`}
            onClick={() => onSelect(card.prompt)}
            className="text-left p-4 rounded-xl border border-border hover:border-atom-spark hover:bg-atom-mist/50 transition-all group"
          >
            <div className="font-medium text-sm text-foreground group-hover:text-atom-nucleus">
              {card.title}
            </div>
            <div className="text-xs text-muted-foreground mt-1 line-clamp-2">
              {card.description}
            </div>
          </button>
        ))}
      </div>
      <div className="flex justify-center mt-3">
        <button
          onClick={handleRefresh}
          className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-atom-core transition-colors"
        >
          <RefreshCw className="w-3 h-3" />
          换一批
        </button>
      </div>
    </div>
  );
}
