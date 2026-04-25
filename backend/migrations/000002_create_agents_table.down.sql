-- Drop agents table
DROP INDEX IF EXISTS idx_agents_knowledge_ids;
DROP INDEX IF EXISTS idx_agents_user_id;
DROP TABLE IF EXISTS agents CASCADE;