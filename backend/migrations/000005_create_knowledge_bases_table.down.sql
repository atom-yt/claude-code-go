-- Drop knowledge_bases table
DROP INDEX IF EXISTS idx_knowledge_bases_user_id;
DROP TABLE IF EXISTS knowledge_bases CASCADE;