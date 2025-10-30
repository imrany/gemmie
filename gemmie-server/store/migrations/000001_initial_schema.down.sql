-- Dropped current_chat_id column from table user_data
ALTER TABLE user_data DROP COLUMN IF EXISTS current_chat_id;
