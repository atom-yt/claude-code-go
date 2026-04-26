'use client';

import { RecommendCards } from './RecommendCards';
import { ChatInput } from './ChatInput';

interface ChatHomeProps {
  onSendMessage: (message: string) => void;
  isLoading?: boolean;
}

export function ChatHome({ onSendMessage, isLoading }: ChatHomeProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full px-4">
      <div className="flex flex-col items-center gap-6 -mt-16">
        {/* Atom Logo */}
        <div className="w-20 h-20 rounded-full bg-gradient-to-br from-atom-spark to-atom-core flex items-center justify-center shadow-lg shadow-atom-core/20">
          <div className="w-14 h-14 rounded-full bg-white/20 flex items-center justify-center">
            <span className="text-white text-2xl font-bold">a</span>
          </div>
        </div>

        {/* Branding */}
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground">
            I&apos;m atom.
          </h1>
          <p className="text-lg text-atom-core font-medium mt-1">
            小 原 子 , 大 能 量
          </p>
          <p className="text-sm text-muted-foreground mt-2">
            有什么我能帮忙的？试试输入 / 召唤技能
          </p>
        </div>

        {/* Recommend Cards */}
        <RecommendCards onSelect={onSendMessage} />
      </div>

      {/* Input */}
      <div className="w-full max-w-2xl mt-8">
        <ChatInput onSend={onSendMessage} disabled={isLoading} />
      </div>
    </div>
  );
}
