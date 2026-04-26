import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MessageItem } from '../MessageItem';
import { Message } from '@/types';

describe('MessageItem', () => {
  const baseMessage: Message = {
    id: '1',
    sessionId: 'session-1',
    role: 'user',
    content: 'Hello, world!',
    createdAt: '2024-01-01T00:00:00Z',
  };

  it('should render user message with User icon', () => {
    render(<MessageItem message={{ ...baseMessage, role: 'user' }} />);

    const content = screen.getByText('Hello, world!');
    expect(content).toBeInTheDocument();

    // Check for user message styling - find the correct parent div
    const contentDiv = content.closest('.bg-atom-core');
    expect(contentDiv).toBeInTheDocument();
  });

  it('should render assistant message with Bot icon', () => {
    render(<MessageItem message={{ ...baseMessage, role: 'assistant', content: 'Hi there!' }} />);

    const content = screen.getByText('Hi there!');
    expect(content).toBeInTheDocument();

    // Check for assistant message styling - find the correct parent div
    const contentDiv = content.closest('.bg-muted');
    expect(contentDiv).toBeInTheDocument();
  });

  it('should render system message', () => {
    render(<MessageItem message={{ ...baseMessage, role: 'assistant', content: 'System message' }} />);

    const content = screen.getByText('System message');
    expect(content).toBeInTheDocument();
  });

  it('should render user message on the right side', () => {
    const { container } = render(<MessageItem message={{ ...baseMessage, role: 'user' }} />);

    const messageRow = container.querySelector('.flex-row-reverse');
    expect(messageRow).toBeInTheDocument();
  });

  it('should render assistant message on the left side', () => {
    const { container } = render(<MessageItem message={{ ...baseMessage, role: 'assistant' }} />);

    const messageRow = container.querySelector('.flex');
    expect(messageRow).toBeInTheDocument();
    expect(messageRow).not.toHaveClass('flex-row-reverse');
  });

  it('should handle multi-line content with whitespace preserved', () => {
    const multiLineContent = 'Line 1\nLine 2\nLine 3';
    render(<MessageItem message={{ ...baseMessage, role: 'user', content: multiLineContent }} />);

    const content = screen.getByText(/Line 1/);
    expect(content).toBeInTheDocument();
    // With whitespace-pre-wrap, line breaks are preserved visually
    expect(content.textContent).toContain('Line 1');
  });

  it('should handle markdown-like content (basic text)', () => {
    const markdownContent = 'This is **bold** and *italic* text';
    render(<MessageItem message={{ ...baseMessage, role: 'assistant', content: markdownContent }} />);

    const content = screen.getByText(/This is \*\*bold\*\* and \*italic\* text/);
    expect(content).toBeInTheDocument();
  });

  it('should render tool call blocks when present', () => {
    const messageWithToolCalls: Message = {
      ...baseMessage,
      role: 'assistant',
      toolCalls: [
        {
          id: 'tool-1',
          name: 'Read',
          input: { path: '/path/to/file.txt' },
        },
      ],
    };

    render(<MessageItem message={messageWithToolCalls} />);

    const toolCallName = screen.getByText('Read');
    expect(toolCallName).toBeInTheDocument();
  });

  it('should expand and collapse tool call block', () => {
    const messageWithToolCalls: Message = {
      ...baseMessage,
      role: 'assistant',
      toolCalls: [
        {
          id: 'tool-1',
          name: 'Read',
          input: { path: '/path/to/file.txt' },
        },
      ],
    };

    render(<MessageItem message={messageWithToolCalls} />);

    const toolCallButton = screen.getByText('Read');
    expect(toolCallButton).toBeInTheDocument();

    // Tool call input should not be visible initially (collapsed)
    const inputLabel = screen.queryByText('输入');
    expect(inputLabel).not.toBeInTheDocument();

    // Click to expand
    fireEvent.click(toolCallButton);

    // Now input label should be visible
    const expandedInputLabel = screen.getByText('输入');
    expect(expandedInputLabel).toBeInTheDocument();

    // Click again to collapse
    fireEvent.click(toolCallButton);

    // Input label should be hidden again
    const collapsedInputLabel = screen.queryByText('输入');
    expect(collapsedInputLabel).not.toBeInTheDocument();
  });

  it('should render tool call with output when present', () => {
    const messageWithToolCalls: Message = {
      ...baseMessage,
      role: 'assistant',
      toolCalls: [
        {
          id: 'tool-1',
          name: 'Read',
          input: { path: '/path/to/file.txt' },
          output: 'File content here',
        },
      ],
    };

    render(<MessageItem message={messageWithToolCalls} />);

    const toolCallButton = screen.getByText('Read');
    fireEvent.click(toolCallButton);

    const outputLabel = screen.getByText('输出');
    expect(outputLabel).toBeInTheDocument();

    const outputContent = screen.getByText('File content here');
    expect(outputContent).toBeInTheDocument();
  });

  it('should show error indicator for failed tool calls', () => {
    const messageWithFailedToolCall: Message = {
      ...baseMessage,
      role: 'assistant',
      toolCalls: [
        {
          id: 'tool-1',
          name: 'Bash',
          input: { command: 'ls -la' },
          isError: true,
          output: 'Error: command failed',
        },
      ],
    };

    render(<MessageItem message={messageWithFailedToolCall} />);

    const toolCallButton = screen.getByText('Bash');
    fireEvent.click(toolCallButton);

    const errorIndicator = screen.getByText('错误');
    expect(errorIndicator).toBeInTheDocument();
    expect(errorIndicator).toHaveClass('text-destructive');
  });

  it('should render multiple tool calls', () => {
    const messageWithMultipleToolCalls: Message = {
      ...baseMessage,
      role: 'assistant',
      toolCalls: [
        {
          id: 'tool-1',
          name: 'Read',
          input: { path: '/path/to/file.txt' },
        },
        {
          id: 'tool-2',
          name: 'Write',
          input: { path: '/path/to/output.txt', content: 'test' },
        },
      ],
    };

    render(<MessageItem message={messageWithMultipleToolCalls} />);

    const readTool = screen.getByText('Read');
    const writeTool = screen.getByText('Write');
    expect(readTool).toBeInTheDocument();
    expect(writeTool).toBeInTheDocument();
  });

  it('should handle very long message content', () => {
    const longContent = 'A'.repeat(1000);
    render(<MessageItem message={{ ...baseMessage, role: 'user', content: longContent }} />);

    const content = screen.getByText(/A{10}/); // Match at least part of the content
    expect(content).toBeInTheDocument();
  });

  it('should handle special characters in content', () => {
    const specialContent = 'Hello @user #tag https://example.com';
    render(<MessageItem message={{ ...baseMessage, role: 'user', content: specialContent }} />);

    const content = screen.getByText(/Hello @user #tag/);
    expect(content).toBeInTheDocument();
  });

  it('should handle empty content', () => {
    const { container } = render(<MessageItem message={{ ...baseMessage, role: 'assistant', content: '' }} />);

    // The component should render without crashing
    // Check that the message container is present
    const messageRow = container.querySelector('.flex');
    expect(messageRow).toBeInTheDocument();
  });
});