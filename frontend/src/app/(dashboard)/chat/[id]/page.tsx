'use client';

import { useEffect } from 'react';
import { useParams } from 'next/navigation';
import { Chat } from '@/components/chat/Chat';
import { useSessionsStore } from '@/stores/sessionsStore';

export default function ChatPage() {
  const params = useParams();
  const sessionId = params.id as string;
  const { selectSession, sessions, fetchSessions, fetchMessages } = useSessionsStore();

  useEffect(() => {
    if (sessions.length === 0) {
      fetchSessions();
    }
  }, [sessions.length, fetchSessions]);

  useEffect(() => {
    const session = sessions.find((s) => s.id === sessionId);
    if (session) {
      selectSession(session);
    } else {
      // Fetch messages directly even if session list hasn't loaded yet
      fetchMessages(sessionId);
    }
  }, [sessionId, sessions, selectSession, fetchMessages]);

  // Check for pending message from home page
  useEffect(() => {
    const pending = sessionStorage.getItem('pendingMessage');
    if (pending) {
      sessionStorage.removeItem('pendingMessage');
      // The Chat component will handle sending this via its own state
      // We pass it as initial message via a custom event
      window.dispatchEvent(new CustomEvent('pendingMessage', { detail: pending }));
    }
  }, []);

  return <Chat sessionId={sessionId} />;
}
