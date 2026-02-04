CREATE INDEX IF NOT EXISTS idx_budgets_user_period ON budgets(user_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, date) WHERE deleted_at IS NULL;
