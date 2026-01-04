-- Users jadvaliga yetishmayotgan ustunlarni qo'shish
ALTER TABLE users
ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS region TEXT NOT NULL DEFAULT 'US',
ADD COLUMN IF NOT EXISTS primary_currency TEXT NOT NULL DEFAULT 'USD',
ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user',
ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Performance uchun indexlar
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;

-- Validation constraintlar
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_role;
ALTER TABLE users ADD CONSTRAINT check_role CHECK (role IN ('user', 'admin', 'premium'));

ALTER TABLE users DROP CONSTRAINT IF EXISTS check_status;
ALTER TABLE users ADD CONSTRAINT check_status CHECK (status IN ('active', 'suspended', 'deleted'));
