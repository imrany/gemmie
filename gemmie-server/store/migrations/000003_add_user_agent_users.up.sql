-- 000003_add_user_agent_users.up.sql
ALTER TABLE users ADD COLUMN user_agent TEXT DEFAULT '';
