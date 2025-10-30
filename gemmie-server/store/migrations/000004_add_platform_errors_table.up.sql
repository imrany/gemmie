-- added platform_error table
-- created_at is typed text because it coming from client-side
CREATE TABLE IF NOT EXISTS platform_errors (
 id TEXT PRIMARY KEY,
 user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
 message TEXT,
 description TEXT,
 action TEXT,
 status TEXT,
 context TEXT,
 severity TEXT,
 created_at TEXT,
 updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
