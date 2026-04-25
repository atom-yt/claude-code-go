-- Drop messages table
DROP INDEX IF EXISTS idx_messages_tool_calls;
DROP INDEX IF EXISTS idx_messages_content;
DROP INDEX IF EXISTS idx_messages_session_created;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP TABLE IF EXISTS messages CASCADE;