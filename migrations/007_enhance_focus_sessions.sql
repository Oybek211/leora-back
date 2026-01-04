-- 007_enhance_focus_sessions.sql
-- Focus Sessions table enhancement based on BACKEND_API.md

-- Add new columns to focus_sessions
ALTER TABLE focus_sessions
    ADD COLUMN IF NOT EXISTS goal_id UUID REFERENCES goals(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS actual_minutes INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ended_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS interruptions_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS notes TEXT;

-- Update old status values to new valid values
UPDATE focus_sessions SET status = 'in_progress' WHERE status = 'active';
UPDATE focus_sessions SET status = 'completed' WHERE status NOT IN ('in_progress', 'completed', 'canceled', 'paused');

-- Add CHECK constraint for status
ALTER TABLE focus_sessions DROP CONSTRAINT IF EXISTS check_focus_session_status;
ALTER TABLE focus_sessions ADD CONSTRAINT check_focus_session_status
    CHECK (status IN ('in_progress', 'completed', 'canceled', 'paused'));

-- Add CHECK constraint for actual_minutes
ALTER TABLE focus_sessions DROP CONSTRAINT IF EXISTS check_actual_minutes;
ALTER TABLE focus_sessions ADD CONSTRAINT check_actual_minutes
    CHECK (actual_minutes >= 0);

-- Add CHECK constraint for interruptions_count
ALTER TABLE focus_sessions DROP CONSTRAINT IF EXISTS check_interruptions_count;
ALTER TABLE focus_sessions ADD CONSTRAINT check_interruptions_count
    CHECK (interruptions_count >= 0);

-- Create index for goal_id
CREATE INDEX IF NOT EXISTS idx_focus_sessions_goal_id ON focus_sessions(goal_id) WHERE deleted_at IS NULL;

-- Create index for started_at
CREATE INDEX IF NOT EXISTS idx_focus_sessions_started_at ON focus_sessions(started_at DESC) WHERE deleted_at IS NULL;
