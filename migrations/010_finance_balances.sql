-- Migration 010: add stored balances and system opening transaction type

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS current_balance DECIMAL(19,4) NOT NULL DEFAULT 0;

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS is_main BOOLEAN NOT NULL DEFAULT false;

UPDATE accounts
SET current_balance = initial_balance
WHERE current_balance = 0;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type
    CHECK (type IN ('income', 'expense', 'transfer', 'system_opening'));

CREATE INDEX IF NOT EXISTS idx_transactions_from_account_id ON transactions(from_account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_to_account_id ON transactions(to_account_id) WHERE deleted_at IS NULL;
