-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content JSONB NOT NULL,
    tool_calls JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for faster queries
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_session_created ON messages(session_id, created_at DESC);

-- Create GIN index on JSONB content for efficient queries
CREATE INDEX idx_messages_content ON messages USING GIN(content);

-- Create GIN index on tool_calls for efficient queries
CREATE INDEX idx_messages_tool_calls ON messages USING GIN(tool_calls);

-- Add comment
COMMENT ON TABLE messages IS 'Messages in chat sessions';
COMMENT ON COLUMN messages.role IS 'Message role: user, assistant, or system';
COMMENT ON COLUMN messages.content IS 'Message content as JSONB (text, images, etc.)';
COMMENT ON COLUMN messages.tool_calls IS 'Tool use blocks as JSONB';