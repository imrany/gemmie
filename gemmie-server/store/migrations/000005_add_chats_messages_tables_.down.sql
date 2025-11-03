-- Dropped table user_data
-- DROP TABLE IF EXISTS user_data CASCADE;

-- Dropped message_count column from table chats
ALTER TABLE chats DROP COLUMN IF EXISTS message_count;
