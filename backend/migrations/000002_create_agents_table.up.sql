-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    system_prompt TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    temperature DECIMAL(3,2) DEFAULT 0.7 CHECK (temperature >= 0 AND temperature <= 2),
    max_tokens INTEGER DEFAULT 4096 CHECK (max_tokens > 0),
    tools TEXT[],
    knowledge_ids UUID[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on user_id for faster queries
CREATE INDEX idx_agents_user_id ON agents(user_id);

-- Create index on knowledge_ids for efficient knowledge base lookups
CREATE INDEX idx_agents_knowledge_ids ON agents USING GIN(knowledge_ids);

-- Add comment
COMMENT ON TABLE agents IS 'AI agents configuration';
COMMENT ON COLUMN agents.user_id IS 'Owner of the agent';
COMMENT ON COLUMN agents.system_prompt IS 'System prompt for the agent';
COMMENT ON COLUMN agents.tools IS 'Array of tool names available to the agent';
COMMENT ON COLUMN agents.knowledge_ids IS 'Array of knowledge base IDs the agent can access';