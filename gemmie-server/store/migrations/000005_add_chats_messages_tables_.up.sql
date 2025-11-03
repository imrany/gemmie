-- Instead of storing all chats as JSON, create a proper chats table
CREATE TABLE IF NOT EXISTS chats (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_archived BOOLEAN DEFAULT false,
    message_count INTEGER DEFAULT 0,
    is_private BOOLEAN DEFAULT true,
    last_message_at TIMESTAMP
);

CREATE INDEX idx_chats_user_id ON chats(user_id);
CREATE INDEX idx_chats_user_updated ON chats(user_id, updated_at DESC);
CREATE INDEX idx_chats_user_archived ON chats(user_id, is_archived);

-- Store messages separately
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    prompt TEXT,
    response TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    model TEXT,
    references_ids TEXT
);

CREATE INDEX idx_messages_chat_id ON messages(chat_id, created_at);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- For pagination
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
