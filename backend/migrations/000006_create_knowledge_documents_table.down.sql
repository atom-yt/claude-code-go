-- Drop knowledge_documents table
DROP INDEX IF EXISTS idx_knowledge_documents_status;
DROP INDEX IF EXISTS idx_knowledge_documents_kb_id;
DROP TABLE IF EXISTS knowledge_documents CASCADE;