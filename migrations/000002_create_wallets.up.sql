CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    balance NUMERIC(20,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'RUB',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
