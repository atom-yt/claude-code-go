-- Enable pgvector extension if not already enabled
CREATE EXTENSION IF NOT EXISTS vector;

-- Create knowledge_chunks table
CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (document_id, chunk_index)
);

-- Create indexes for faster queries
CREATE INDEX idx_knowledge_chunks_document_id ON knowledge_chunks(document_id);

-- Create HNSW index on embedding for vector similarity search
CREATE INDEX idx_knowledge_chunks_embedding ON knowledge_chunks USING hnsw (embedding vector_cosine_ops);

-- Create GIN index on metadata for efficient queries
CREATE INDEX idx_knowledge_chunks_metadata ON knowledge_chunks USING GIN(metadata);

-- Add comment
COMMENT ON TABLE knowledge_chunks IS 'Chunks of knowledge documents with embeddings for semantic search';
COMMENT ON COLUMN knowledge_chunks.chunk_index IS 'Position of this chunk in the document';
COMMENT ON COLUMN knowledge_chunks.embedding IS 'Vector embedding (1536 dimensions) using pgvector';
COMMENT ON COLUMN knowledge_chunks.metadata IS 'Additional metadata about the chunk';