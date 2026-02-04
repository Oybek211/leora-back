CREATE TABLE IF NOT EXISTS finance_quick_exp_categories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    category_tag TEXT NOT NULL,
    category_name TEXT,
    category_type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_finance_quick_exp_categories_user_type
    ON finance_quick_exp_categories (user_id, category_type);
