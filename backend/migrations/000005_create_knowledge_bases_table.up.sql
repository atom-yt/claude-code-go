-- Create knowledge_bases table
CREATE TABLE IF NOT EXISTS knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on user_id for faster queries
CREATE INDEX idx_knowledge_bases_user_id ON knowledge_bases(user_id);

-- Add comment
COMMENT ON TABLE knowledge_bases IS 'Knowledge bases for RAG capabilities';
COMMENT ON COLUMN knowledge_bases.name IS 'Name of the knowledge base';
COMMENT ON COLUMN knowledge_bases.description IS 'Description of the knowledge base contents';