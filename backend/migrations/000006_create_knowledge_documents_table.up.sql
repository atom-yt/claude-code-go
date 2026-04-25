-- Create knowledge_documents table
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content TEXT,
    status VARCHAR(20) DEFAULT 'processing' CHECK (status IN ('processing', 'ready', 'failed')),
    error_message TEXT,
    file_size BIGINT,
    mime_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on kb_id for faster queries
CREATE INDEX idx_knowledge_documents_kb_id ON knowledge_documents(kb_id);
CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(status);

-- Add comment
COMMENT ON TABLE knowledge_documents IS 'Documents uploaded to knowledge bases';
COMMENT ON COLUMN knowledge_documents.kb_id IS 'Reference to the parent knowledge base';
COMMENT ON COLUMN knowledge_documents.status IS 'Processing status: processing, ready, or failed';
COMMENT ON COLUMN knowledge_documents.error_message IS 'Error message if processing failed';