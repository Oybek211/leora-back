-- Finance categories and localization support

CREATE TABLE IF NOT EXISTS finance_categories (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    name_i18n JSONB NOT NULL DEFAULT '{}'::jsonb,
    icon_name TEXT NOT NULL,
    color TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_finance_categories_type_active ON finance_categories (type, is_active);
CREATE INDEX IF NOT EXISTS idx_finance_categories_sort ON finance_categories (sort_order);

CREATE TABLE IF NOT EXISTS meta_languages (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meta_languages_active ON meta_languages (is_active);

CREATE TABLE IF NOT EXISTS error_translations (
    code TEXT NOT NULL,
    lang_code TEXT NOT NULL REFERENCES meta_languages(code) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (code, lang_code)
);
