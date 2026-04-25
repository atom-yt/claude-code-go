-- Drop knowledge_chunks table
DROP INDEX IF EXISTS idx_knowledge_chunks_metadata;
DROP INDEX IF EXISTS idx_knowledge_chunks_embedding;
DROP INDEX IF EXISTS idx_knowledge_chunks_document_id;
DROP TABLE IF EXISTS knowledge_chunks CASCADE;

-- Note: We do NOT drop the vector extension as other tables might depend on it