-- Migration 011: expand transaction types and add references for ledger entries

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS reference_type TEXT,
    ADD COLUMN IF NOT EXISTS reference_id UUID;

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
            'debt_adjustment'
        )
    );

CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference_type, reference_id) WHERE deleted_at IS NULL;
