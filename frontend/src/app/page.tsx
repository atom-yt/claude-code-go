'use client';

import { useRouter } from 'next/navigation';
import { ChatHome } from '@/components/chat/ChatHome';
import { useSessionsStore } from '@/stores/sessionsStore';
import { useState } from 'react';

export default function Home() {
  const router = useRouter();
  const { createSession, selectSession } = useSessionsStore();
  const [isLoading, setIsLoading] = useState(false);

  const handleSendMessage = async (message: string) => {
    setIsLoading(true);
    try {
      const session = await createSession({ title: message.slice(0, 50) });
      selectSession(session);
      // Store the pending message to be sent after navigation
      sessionStorage.setItem('pendingMessage', message);
      router.push(`/chat/${session.id}`);
    } catch {
      setIsLoading(false);
    }
  };

  return <ChatHome onSendMessage={handleSendMessage} isLoading={isLoading} />;
}
