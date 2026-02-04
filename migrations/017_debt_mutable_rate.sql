-- 017: Add mutable exchange rate to debts and applied_rate to debt_payments
-- exchange_rate_current: single source of truth for repayment conversions (mutable)
-- applied_rate: audit trail for which rate was used per payment

ALTER TABLE debts ADD COLUMN IF NOT EXISTS exchange_rate_current DECIMAL(19,10) DEFAULT 0;

ALTER TABLE debt_payments ADD COLUMN IF NOT EXISTS applied_rate DECIMAL(19,10) DEFAULT 0;
