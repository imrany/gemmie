-- Assuming message_id is VARCHAR, ensure it's large enough (e.g., VARCHAR(255))
ALTER TABLE arcades ADD COLUMN message_id VARCHAR(255);

-- Generate a unique UUID for every existing row where message_id is NULL
UPDATE arcades
SET message_id = gen_random_uuid()::VARCHAR(255)
WHERE message_id IS NULL;

-- Now add the UNIQUE and NOT NULL constraints
ALTER TABLE arcades ADD CONSTRAINT arcades_message_id_key UNIQUE (message_id);
ALTER TABLE arcades ALTER COLUMN message_id SET NOT NULL;

-- Continue with the rest of your migration (uncommented if needed)
CREATE INDEX unique_code ON arcades USING HASH (code);
