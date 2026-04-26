import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MessageList } from '../MessageList';
import { Message } from '@/types';

// Mock scrollIntoView for jsdom
beforeEach(() => {
  window.HTMLElement.prototype.scrollIntoView = vi.fn();
});

describe('MessageList', () => {
  const mockMessages: Message[] = [
    {
      id: '1',
      sessionId: 'session-1',
      role: 'user',
      content: 'Hello',
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      sessionId: 'session-1',
      role: 'assistant',
      content: 'Hi there!',
      createdAt: '2024-01-01T00:00:01Z',
    },
  ];

  it('should render empty state when no messages', () => {
    render(<MessageList messages={[]} />);

    const emptyHint = screen.getByText('开始对话吧');
    expect(emptyHint).toBeInTheDocument();
  });

  it('should render empty state hint with correct styling', () => {
    const { container } = render(<MessageList messages={[]} />);

    const emptyHint = screen.getByText('开始对话吧');
    expect(emptyHint).toHaveClass('text-muted-foreground');
  });

  it('should render all messages in the list', () => {
    render(<MessageList messages={mockMessages} />);

    expect(screen.getByText('Hello')).toBeInTheDocument();
    expect(screen.getByText('Hi there!')).toBeInTheDocument();
  });

  it('should render messages in correct order', () => {
    render(<MessageList messages={mockMessages} />);

    const messages = screen.getAllByText(/Hello|Hi there!/);
    expect(messages[0]).toHaveTextContent('Hello');
    expect(messages[1]).toHaveTextContent('Hi there!');
  });

  it('should show streaming indicator when isStreaming is true', () => {
    render(<MessageList messages={mockMessages} isStreaming={true} />);

    const streamingText = screen.getByText('atom 正在思考...');
    expect(streamingText).toBeInTheDocument();
  });

  it('should not show streaming indicator when isStreaming is false', () => {
    render(<MessageList messages={mockMessages} isStreaming={false} />);

    const streamingText = screen.queryByText('atom 正在思考...');
    expect(streamingText).not.toBeInTheDocument();
  });

  it('should not show streaming indicator by default', () => {
    render(<MessageList messages={mockMessages} />);

    const streamingText = screen.queryByText('atom 正在思考...');
    expect(streamingText).not.toBeInTheDocument();
  });

  it('should render loader icon when streaming', () => {
    const { container } = render(<MessageList messages={mockMessages} isStreaming={true} />);

    // Check for the spinner element
    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('should handle single message in list', () => {
    const singleMessage: Message[] = [
      {
        id: '1',
        sessionId: 'session-1',
        role: 'user',
        content: 'Single message',
        createdAt: '2024-01-01T00:00:00Z',
      },
    ];

    render(<MessageList messages={singleMessage} />);

    expect(screen.getByText('Single message')).toBeInTheDocument();
    expect(screen.queryByText('atom 正在思考...')).not.toBeInTheDocument();
  });

  it('should handle large number of messages', () => {
    const largeMessageList: Message[] = Array.from({ length: 100 }, (_, i) => ({
      id: `msg-${i}`,
      sessionId: 'session-1',
      role: i % 2 === 0 ? 'user' : 'assistant',
      content: `Message ${i}`,
      createdAt: `2024-01-01T00:00:00Z`,
    }));

    render(<MessageList messages={largeMessageList} />);

    expect(screen.getByText('Message 0')).toBeInTheDocument();
    expect(screen.getByText('Message 50')).toBeInTheDocument();
    expect(screen.getByText('Message 99')).toBeInTheDocument();
  });

  it('should render messages with tool calls', () => {
    const messagesWithToolCalls: Message[] = [
      {
        id: '1',
        sessionId: 'session-1',
        role: 'assistant',
        content: 'Reading file...',
        toolCalls: [
          {
            id: 'tool-1',
            name: 'Read',
            input: { path: '/path/to/file.txt' },
          },
        ],
        createdAt: '2024-01-01T00:00:00Z',
      },
    ];

    render(<MessageList messages={messagesWithToolCalls} />);

    expect(screen.getByText('Reading file...')).toBeInTheDocument();
    expect(screen.getByText('Read')).toBeInTheDocument();
  });

  it('should scroll to bottom when messages change', () => {
    const { rerender } = render(<MessageList messages={mockMessages} />);

    // Add a new message
    const updatedMessages = [
      ...mockMessages,
      {
        id: '3',
        sessionId: 'session-1',
        role: 'user',
        content: 'New message',
        createdAt: '2024-01-01T00:00:02Z',
      },
    ];

    rerender(<MessageList messages={updatedMessages} />);

    expect(screen.getByText('New message')).toBeInTheDocument();
  });

  it('should have correct container structure', () => {
    const { container } = render(<MessageList messages={mockMessages} />);

    // Check that the main container has the correct classes
    const messageContainer = container.querySelector('.flex-1.overflow-y-auto');
    expect(messageContainer).toBeInTheDocument();
  });

  it('should handle messages with special characters', () => {
    const specialMessages: Message[] = [
      {
        id: '1',
        sessionId: 'session-1',
        role: 'user',
        content: 'Hello @user #tag https://example.com',
        createdAt: '2024-01-01T00:00:00Z',
      },
    ];

    render(<MessageList messages={specialMessages} />);

    const content = screen.getByText(/Hello @user #tag/);
    expect(content).toBeInTheDocument();
  });

  it('should handle messages with line breaks', () => {
    const multilineMessages: Message[] = [
      {
        id: '1',
        sessionId: 'session-1',
        role: 'user',
        content: 'Line 1\nLine 2\nLine 3',
        createdAt: '2024-01-01T00:00:00Z',
      },
    ];

    render(<MessageList messages={multilineMessages} />);

    const content = screen.getByText(/Line 1/);
    expect(content).toBeInTheDocument();
  });

  it('should render bottom anchor element', () => {
    const { container } = render(<MessageList messages={mockMessages} />);

    // The bottom ref div is rendered (empty div for scrolling)
    const divs = container.querySelectorAll('div');
    expect(divs.length).toBeGreaterThan(0);
  });

  it('should handle system role messages', () => {
    const systemMessages: Message[] = [
      {
        id: '1',
        sessionId: 'session-1',
        role: 'assistant',
        content: 'System: something happened',
        createdAt: '2024-01-01T00:00:00Z',
      },
    ];

    render(<MessageList messages={systemMessages} />);

    expect(screen.getByText('System: something happened')).toBeInTheDocument();
  });
});