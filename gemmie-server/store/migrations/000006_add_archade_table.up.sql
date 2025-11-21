-- added archade table
-- created_at is typed text because it coming from client-side
CREATE TABLE IF NOT EXISTS archades (
 id TEXT PRIMARY KEY,
 user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
 code TEXT NOT NULL,
 label TEXT NOT NULL,
 description TEXT,
 code_type TEXT DEFAULT 'html',
 created_at TEXT,
 updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS archades_user_id_idx ON archades(user_id);
