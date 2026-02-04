-- 018: Social login support — auth_identities table
-- Links external providers (google, apple) to internal users

CREATE TABLE IF NOT EXISTS auth_identities (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider    TEXT NOT NULL,           -- 'google', 'apple'
    provider_id TEXT NOT NULL,           -- sub claim from ID token
    email       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(provider, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_auth_identities_user ON auth_identities(user_id);
