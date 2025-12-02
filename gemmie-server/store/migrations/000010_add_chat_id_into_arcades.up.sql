 -- added chat_id into arcades
 ALTER TABLE arcades ADD COLUMN chat_id TEXT REFERENCES chats(id);
 UPDATE arcades SET chat_id = NULL WHERE chat_id = '';
 ALTER TABLE arcades ALTER COLUMN chat_id DROP DEFAULT;

  -- added a unique constraint to make sure each chat only has one arcade
  ALTER TABLE arcades ADD CONSTRAINT unique_chat_id UNIQUE (chat_id);
