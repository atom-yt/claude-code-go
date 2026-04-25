-- Drop triggers
DROP TRIGGER IF EXISTS trigger_chat_sessions_updated_at ON chat_sessions;
DROP TRIGGER IF EXISTS trigger_agents_updated_at ON agents;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();