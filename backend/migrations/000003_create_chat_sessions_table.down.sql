-- Drop chat_sessions table
DROP INDEX IF EXISTS idx_chat_sessions_status;
DROP INDEX IF EXISTS idx_chat_sessions_agent_id;
DROP INDEX IF EXISTS idx_chat_sessions_user_id;
DROP TABLE IF EXISTS chat_sessions CASCADE;