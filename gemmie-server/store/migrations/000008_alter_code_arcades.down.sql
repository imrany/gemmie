-- DROP INDEX IF EXISTS unique_code;
-- ALTER TABLE arcades DROP CONSTRAINT IF EXISTS unique_code;

-- drops message_id from arcades table
ALTER TABLE arcades DROP COLUMN message_id;
