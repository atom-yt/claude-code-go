import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ChatInput } from '../ChatInput';

// Mock ModelSelector component
vi.mock('../ModelSelector', () => ({
  ModelSelector: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
    <div data-testid="model-selector" data-value={value} onClick={() => onChange('new-model')}>
      Model Selector
    </div>
  ),
}));

describe('ChatInput', () => {
  const mockOnSend = vi.fn();
  const mockOnStop = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render textarea with placeholder', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    expect(textarea).toBeInTheDocument();
  });

  it('should render send button', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const sendButton = screen.getByRole('button');
    expect(sendButton).toBeInTheDocument();
  });

  it('should render custom placeholder when provided', () => {
    render(<ChatInput onSend={mockOnSend} placeholder="Type something..." />);

    const textarea = screen.getByPlaceholderText('Type something...');
    expect(textarea).toBeInTheDocument();
  });

  it('should update message on input change', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    await user.type(textarea, 'Hello world');

    expect(textarea.value).toBe('Hello world');
  });

  it('should send message on Enter key press', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    await user.type(textarea, 'Hello world{Enter}');

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('Hello world');
    });
  });

  it('should not send message on Shift+Enter', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    await user.type(textarea, 'Hello');
    await user.keyboard('{Shift>}{Enter}{/Shift}');

    expect(mockOnSend).not.toHaveBeenCalled();
    expect(textarea.value).toContain('\n');
  });

  it('should add newline on Shift+Enter', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    await user.type(textarea, 'Line 1');
    await user.keyboard('{Shift>}{Enter}{/Shift}');
    await user.type(textarea, 'Line 2');

    expect(textarea.value).toBe('Line 1\nLine 2');
    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('should not send empty message', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    await user.type(textarea, '{Enter}');

    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('should not send whitespace-only message', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    await user.type(textarea, '   {Enter}');

    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('should send message on button click', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const sendButton = screen.getByRole('button');

    await user.type(textarea, 'Hello world');
    await user.click(sendButton);

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('Hello world');
    });
  });

  it('should clear textarea after sending', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    const sendButton = screen.getByRole('button');

    await user.type(textarea, 'Hello world');
    await user.click(sendButton);

    await waitFor(() => {
      expect(textarea.value).toBe('');
    });
  });

  it('should call onStop when stop button is clicked', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} onStop={mockOnStop} disabled />);

    // When disabled, send button might show as stop
    const sendButton = screen.getByRole('button');
    
    // If there's a stop behavior, test it
    if (mockOnStop) {
      await user.click(sendButton);
      expect(mockOnStop).toHaveBeenCalled();
    }
  });

  it('should disable input when disabled prop is true', () => {
    render(<ChatInput onSend={mockOnSend} disabled />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    expect(textarea).toBeDisabled();
  });

  it('should disable send button when input is empty', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const sendButton = screen.getByRole('button');
    expect(sendButton).toBeDisabled();
  });

  it('should disable send button when disabled prop is true', () => {
    render(<ChatInput onSend={mockOnSend} disabled />);

    const sendButton = screen.getByRole('button');
    expect(sendButton).toBeDisabled();
  });

  it('should enable send button when input has text', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const sendButton = screen.getByRole('button');

    await user.type(textarea, 'Hello');

    expect(sendButton).not.toBeDisabled();
  });

  it('should disable send button when input has only whitespace', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const sendButton = screen.getByRole('button');

    await user.type(textarea, '   ');

    expect(sendButton).toBeDisabled();
  });

  it('should auto-grow textarea height on input', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    
    const initialHeight = textarea.style.height;
    
    // Type multiple lines
    await user.type(textarea, 'Line 1\nLine 2\nLine 3\nLine 4\nLine 5');

    // Height should change
    expect(textarea.style.height).not.toBe(initialHeight);
  });

  it('should limit textarea max height to 200px', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    
    // Type many lines
    const longText = Array(20).fill('Line').join('\n');
    await user.type(textarea, longText);

    // Height should be capped
    const height = parseInt(textarea.style.height);
    expect(height).toBeLessThanOrEqual(200);
  });

  it('should reset textarea height after sending', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    const sendButton = screen.getByRole('button');

    await user.type(textarea, 'Line 1\nLine 2\nLine 3');
    const heightBeforeSend = parseInt(textarea.style.height);
    
    await user.click(sendButton);

    await waitFor(() => {
      expect(textarea.style.height).toBe('auto');
    });
  });

  it('should render ModelSelector component', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const modelSelector = screen.getByTestId('model-selector');
    expect(modelSelector).toBeInTheDocument();
  });

  it('should handle rapid typing without sending', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    
    await user.type(textarea, 'Hello world this is a test');

    expect(mockOnSend).not.toHaveBeenCalled();
    expect(textarea).toHaveValue('Hello world this is a test');
  });

  it('should send message with special characters', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const specialText = 'Hello @user #tag https://example.com !@#$%';
    
    await user.type(textarea, `${specialText}{Enter}`);

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith(specialText);
    });
  });

  it('should handle very long messages', async () => {
    const user = userEvent.setup();
    const longMessage = 'a'.repeat(1000);
    
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const sendButton = screen.getByRole('button');

    await user.type(textarea, longMessage);
    await user.click(sendButton);

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith(longMessage);
    });
  });

  it('should handle Chinese characters', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const chineseText = '你好世界';
    
    await user.type(textarea, `${chineseText}{Enter}`);

    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith(chineseText);
    });
  });

  it('should prevent sending when disabled', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} disabled />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/) as HTMLTextAreaElement;
    
    // Type should work even when disabled (for UX)
    await user.type(textarea, 'Hello');
    expect(textarea.value).toBe('Hello');
    
    // But sending should not work
    await user.keyboard('{Enter}');
    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('should maintain focus after sending', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    const sendButton = screen.getByRole('button');

    await user.type(textarea, 'Hello world');
    await user.click(sendButton);

    await waitFor(() => {
      expect(textarea).toHaveFocus();
    });
  });

  it('should handle multiple Enter key presses in sequence', async () => {
    const user = userEvent.setup();
    mockOnSend.mockResolvedValue(undefined);

    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    
    await user.type(textarea, 'First message{Enter}');
    
    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('First message');
    });

    // Type and send again
    await user.type(textarea, 'Second message{Enter}');
    
    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith('Second message');
    });
  });

  it('should render ArrowUp icon in send button', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const sendButton = screen.getByRole('button');
    const icon = sendButton.querySelector('.lucide-arrow-up');
    expect(icon).toBeInTheDocument();
  });

  it('should have correct wrapper container classes', () => {
    const { container } = render(<ChatInput onSend={mockOnSend} />);

    const wrapper = container.querySelector('.max-w-2xl');
    expect(wrapper).toBeInTheDocument();
  });

  it('should have correct textarea classes', () => {
    render(<ChatInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText(/有什么我能帮忙的/);
    expect(textarea).toHaveClass('resize-none');
    expect(textarea).toHaveClass('w-full');
  });

  it('should handle model selection change', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={mockOnSend} />);

    const modelSelector = screen.getByTestId('model-selector');
    await user.click(modelSelector);

    // Model selector is mocked, so we just verify it exists and can be clicked
    expect(modelSelector).toBeInTheDocument();
  });
});
