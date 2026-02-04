-- Migration 014: transaction audit fields, statuses, and expanded types

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'completed',
    ADD COLUMN IF NOT EXISTS metadata JSONB,
    ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMP WITH TIME ZONE DEFAULT now();

-- Backfill occurred_at from created_at where possible
UPDATE transactions
SET occurred_at = COALESCE(occurred_at, created_at, now())
WHERE occurred_at IS NULL;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_status;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_status
    CHECK (status IN ('pending', 'completed', 'failed'));

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type
    CHECK (
        type IN (
            'income',
            'expense',
            'transfer',
            'transfer_in',
            'transfer_out',
            'system_opening',
            'system_adjustment',
            'system_archive',
            'debt_create',
            'debt_payment',
            'debt_adjustment',
            'account_create_funding',
            'account_delete_withdrawal',
            'budget_add_value',
            'debt_add_value',
            'debt_full_payment'
        )
    );

CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at) WHERE deleted_at IS NULL;
