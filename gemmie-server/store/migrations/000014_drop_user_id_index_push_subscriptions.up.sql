-- DROP INDEX idx_push_subscriptions_user_id;
DROP INDEX IF EXISTS idx_push_subscriptions_user_id;
ALTER TABLE push_subscriptions DROP CONSTRAINT push_subscriptions_pkey;
