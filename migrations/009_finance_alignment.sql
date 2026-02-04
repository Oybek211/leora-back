-- Migration 009: Align finance schema with backend docs

-- Accounts
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS initial_balance DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS linked_goal_id UUID REFERENCES goals(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS custom_type_id TEXT,
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active';

ALTER TABLE accounts DROP CONSTRAINT IF EXISTS check_account_show_status;
ALTER TABLE accounts ADD CONSTRAINT check_account_show_status
    CHECK (show_status IN ('active', 'archived', 'deleted'));

-- Transactions
ALTER TABLE transactions
    ALTER COLUMN account_id DROP NOT NULL;

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'expense',
    ADD COLUMN IF NOT EXISTS from_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS to_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS base_currency TEXT,
    ADD COLUMN IF NOT EXISTS rate_used_to_base DECIMAL(19,6) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS converted_amount_to_base DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS to_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS to_currency TEXT,
    ADD COLUMN IF NOT EXISTS effective_rate_from_to DECIMAL(19,6) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS fee_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee_category_id TEXT,
    ADD COLUMN IF NOT EXISTS category_id TEXT,
    ADD COLUMN IF NOT EXISTS subcategory_id TEXT,
    ADD COLUMN IF NOT EXISTS name TEXT,
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS date DATE,
    ADD COLUMN IF NOT EXISTS time TEXT,
    ADD COLUMN IF NOT EXISTS counterparty_id UUID,
    ADD COLUMN IF NOT EXISTS recurring_id UUID,
    ADD COLUMN IF NOT EXISTS attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS is_balance_adjustment BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS skip_budget_matching BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS related_budget_id UUID,
    ADD COLUMN IF NOT EXISTS related_debt_id UUID,
    ADD COLUMN IF NOT EXISTS planned_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS original_currency TEXT,
    ADD COLUMN IF NOT EXISTS original_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS conversion_rate DECIMAL(19,6) NOT NULL DEFAULT 1;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_type;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_type
    CHECK (type IN ('income', 'expense', 'transfer'));

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS check_transaction_show_status;
ALTER TABLE transactions ADD CONSTRAINT check_transaction_show_status
    CHECK (show_status IN ('active', 'archived', 'deleted'));

CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_show_status ON transactions(show_status) WHERE deleted_at IS NULL;

-- Budgets
ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS budget_type TEXT NOT NULL DEFAULT 'category',
    ADD COLUMN IF NOT EXISTS category_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS transaction_type TEXT,
    ADD COLUMN IF NOT EXISTS period_type TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS start_date DATE,
    ADD COLUMN IF NOT EXISTS end_date DATE,
    ADD COLUMN IF NOT EXISTS spent_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS remaining_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS percent_used DECIMAL(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_overspent BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS rollover_mode TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS notify_on_exceed BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS contribution_total DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS current_balance DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active';

ALTER TABLE budgets DROP CONSTRAINT IF EXISTS check_budget_show_status;
ALTER TABLE budgets ADD CONSTRAINT check_budget_show_status
    CHECK (show_status IN ('active', 'archived', 'deleted'));

ALTER TABLE budgets DROP CONSTRAINT IF EXISTS check_budget_period_type;
ALTER TABLE budgets ADD CONSTRAINT check_budget_period_type
    CHECK (period_type IN ('none', 'weekly', 'monthly', 'custom_range'));

-- Debts
ALTER TABLE debts
    ADD COLUMN IF NOT EXISTS direction TEXT NOT NULL DEFAULT 'i_owe',
    ADD COLUMN IF NOT EXISTS counterparty_id UUID,
    ADD COLUMN IF NOT EXISTS counterparty_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS principal_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS principal_currency TEXT NOT NULL DEFAULT 'UZS',
    ADD COLUMN IF NOT EXISTS principal_original_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS principal_original_currency TEXT,
    ADD COLUMN IF NOT EXISTS base_currency TEXT NOT NULL DEFAULT 'UZS',
    ADD COLUMN IF NOT EXISTS rate_on_start DECIMAL(19,6) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS principal_base_value DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS repayment_currency TEXT,
    ADD COLUMN IF NOT EXISTS repayment_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS repayment_rate_on_start DECIMAL(19,6) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS is_fixed_repayment_amount BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS start_date DATE,
    ADD COLUMN IF NOT EXISTS due_date DATE,
    ADD COLUMN IF NOT EXISTS interest_mode TEXT,
    ADD COLUMN IF NOT EXISTS interest_rate_annual DECIMAL(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS schedule_hint TEXT,
    ADD COLUMN IF NOT EXISTS funding_account_id UUID,
    ADD COLUMN IF NOT EXISTS funding_transaction_id UUID,
    ADD COLUMN IF NOT EXISTS lent_from_account_id UUID,
    ADD COLUMN IF NOT EXISTS return_to_account_id UUID,
    ADD COLUMN IF NOT EXISTS received_to_account_id UUID,
    ADD COLUMN IF NOT EXISTS pay_from_account_id UUID,
    ADD COLUMN IF NOT EXISTS custom_rate_used DECIMAL(19,6) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reminder_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS reminder_time TEXT,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS settled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS final_rate_used DECIMAL(19,6) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS final_profit_loss DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS final_profit_loss_currency TEXT,
    ADD COLUMN IF NOT EXISTS total_paid_in_repayment_currency DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS remaining_amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_paid DECIMAL(19,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS percent_paid DECIMAL(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active';

ALTER TABLE debts DROP CONSTRAINT IF EXISTS check_debt_show_status;
ALTER TABLE debts ADD CONSTRAINT check_debt_show_status
    CHECK (show_status IN ('active', 'archived', 'deleted'));

ALTER TABLE debts DROP CONSTRAINT IF EXISTS check_debt_status;
ALTER TABLE debts ADD CONSTRAINT check_debt_status
    CHECK (status IN ('active', 'paid', 'overdue', 'canceled'));

-- Counterparties
CREATE TABLE IF NOT EXISTS counterparties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    phone_number TEXT,
    comment TEXT,
    search_keywords TEXT,
    show_status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_counterparties_user_id ON counterparties(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_counterparties_name ON counterparties(display_name) WHERE deleted_at IS NULL;

-- FX rates
CREATE TABLE IF NOT EXISTS fx_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rate_date DATE NOT NULL,
    from_currency TEXT NOT NULL,
    to_currency TEXT NOT NULL,
    rate DECIMAL(19,6) NOT NULL DEFAULT 0,
    rate_mid DECIMAL(19,6) NOT NULL DEFAULT 0,
    rate_bid DECIMAL(19,6) NOT NULL DEFAULT 0,
    rate_ask DECIMAL(19,6) NOT NULL DEFAULT 0,
    nominal DECIMAL(19,6) NOT NULL DEFAULT 1,
    spread_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'manual',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fx_rates_date ON fx_rates(rate_date);
CREATE INDEX IF NOT EXISTS idx_fx_rates_pair ON fx_rates(from_currency, to_currency, rate_date);

-- Debt payments
CREATE TABLE IF NOT EXISTS debt_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE CASCADE,
    amount DECIMAL(19,4) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    base_currency TEXT NOT NULL,
    rate_used_to_base DECIMAL(19,6) NOT NULL DEFAULT 1,
    converted_amount_to_base DECIMAL(19,4) NOT NULL DEFAULT 0,
    rate_used_to_debt DECIMAL(19,6) NOT NULL DEFAULT 1,
    converted_amount_to_debt DECIMAL(19,4) NOT NULL DEFAULT 0,
    payment_date DATE NOT NULL,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    note TEXT,
    related_transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_debt_payments_debt_id ON debt_payments(debt_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_debt_payments_date ON debt_payments(payment_date) WHERE deleted_at IS NULL;
