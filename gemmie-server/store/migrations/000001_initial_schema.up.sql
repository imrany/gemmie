-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    preferences TEXT,
    work_function TEXT,
    theme TEXT,
    sync_enabled BOOLEAN DEFAULT false,
    plan TEXT,
    plan_name TEXT,
    amount INTEGER DEFAULT 0,
    duration TEXT,
    phone_number TEXT,
    expiry_timestamp BIGINT DEFAULT 0,
    expire_duration BIGINT DEFAULT 0,
    price TEXT,
    response_mode TEXT DEFAULT 'light-response',
    agree_to_terms BOOLEAN DEFAULT false,
    request_count_value INTEGER DEFAULT 0,
    request_count_timestamp BIGINT DEFAULT 0,
    email_verified BOOLEAN DEFAULT false,
    email_subscribed BOOLEAN DEFAULT true,
    verification_token TEXT,
    verification_token_expiry TIMESTAMP,
    unsubscribe_token TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);

-- Create user_data table
CREATE TABLE IF NOT EXISTS user_data (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    chats TEXT,
    link_previews TEXT,
    current_chat_id TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    external_reference TEXT,
    mpesa_receipt_number TEXT UNIQUE,
    checkout_request_id TEXT,
    merchant_request_id TEXT,
    amount INTEGER,
    phone_number TEXT,
    result_code INTEGER,
    result_description TEXT,
    status TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_external_ref ON transactions(external_reference);
CREATE INDEX IF NOT EXISTS idx_transactions_phone ON transactions(phone_number);
CREATE INDEX IF NOT EXISTS idx_transactions_mpesa_receipt ON transactions(mpesa_receipt_number);