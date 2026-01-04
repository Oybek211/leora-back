-- Migration 006: Enhance Habits table with all fields from BACKEND_API.md

-- Add missing columns to habits table
ALTER TABLE habits
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS icon_id TEXT,
ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active',
ADD COLUMN IF NOT EXISTS goal_id UUID REFERENCES goals(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS frequency TEXT NOT NULL DEFAULT 'daily',
ADD COLUMN IF NOT EXISTS days_of_week INTEGER[],
ADD COLUMN IF NOT EXISTS times_per_week INTEGER,
ADD COLUMN IF NOT EXISTS time_of_day TEXT,
ADD COLUMN IF NOT EXISTS completion_mode TEXT NOT NULL DEFAULT 'boolean',
ADD COLUMN IF NOT EXISTS target_per_day DECIMAL(19,4),
ADD COLUMN IF NOT EXISTS unit TEXT,
ADD COLUMN IF NOT EXISTS counting_type TEXT NOT NULL DEFAULT 'create',
ADD COLUMN IF NOT EXISTS difficulty TEXT NOT NULL DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS challenge_length_days INTEGER,
ADD COLUMN IF NOT EXISTS reminder_enabled BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS reminder_time TEXT,
ADD COLUMN IF NOT EXISTS streak_current INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS streak_best INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS completion_rate_30d DECIMAL(5,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS finance_rule JSONB;

-- Drop and recreate constraints for habits fields
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_show_status;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_status;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_type;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_frequency;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_completion_mode;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_counting_type;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_difficulty;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_priority;
ALTER TABLE habits DROP CONSTRAINT IF EXISTS check_habit_completion_rate;

ALTER TABLE habits ADD CONSTRAINT check_habit_show_status CHECK (show_status IN ('active', 'archived', 'deleted'));
ALTER TABLE habits ADD CONSTRAINT check_habit_status CHECK (status IN ('active', 'paused', 'archived'));
ALTER TABLE habits ADD CONSTRAINT check_habit_type CHECK (habit_type IN ('health', 'finance', 'productivity', 'education', 'personal', 'custom'));
ALTER TABLE habits ADD CONSTRAINT check_habit_frequency CHECK (frequency IN ('daily', 'weekly', 'custom'));
ALTER TABLE habits ADD CONSTRAINT check_habit_completion_mode CHECK (completion_mode IN ('boolean', 'numeric'));
ALTER TABLE habits ADD CONSTRAINT check_habit_counting_type CHECK (counting_type IN ('create', 'quit'));
ALTER TABLE habits ADD CONSTRAINT check_habit_difficulty CHECK (difficulty IN ('easy', 'medium', 'hard'));
ALTER TABLE habits ADD CONSTRAINT check_habit_priority CHECK (priority IN ('low', 'medium', 'high'));
ALTER TABLE habits ADD CONSTRAINT check_habit_completion_rate CHECK (completion_rate_30d >= 0 AND completion_rate_30d <= 100);

-- Create habit_completions table
CREATE TABLE IF NOT EXISTS habit_completions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    date_key TEXT NOT NULL,
    status TEXT NOT NULL,
    value DECIMAL(19,4),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT check_completion_status CHECK (status IN ('done', 'miss')),
    UNIQUE(habit_id, date_key)
);

CREATE INDEX IF NOT EXISTS idx_completions_habit_id ON habit_completions(habit_id);
CREATE INDEX IF NOT EXISTS idx_completions_date_key ON habit_completions(date_key);
CREATE INDEX IF NOT EXISTS idx_completions_created_at ON habit_completions(created_at DESC);

-- Create habit_goals junction table (many-to-many)
CREATE TABLE IF NOT EXISTS habit_goals (
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    PRIMARY KEY (habit_id, goal_id)
);

CREATE INDEX IF NOT EXISTS idx_habit_goals_habit ON habit_goals(habit_id);
CREATE INDEX IF NOT EXISTS idx_habit_goals_goal ON habit_goals(goal_id);

-- Additional indexes for habits table
CREATE INDEX IF NOT EXISTS idx_habits_show_status ON habits(show_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_habits_habit_type ON habits(habit_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_habits_frequency ON habits(frequency) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_habits_goal_id ON habits(goal_id) WHERE deleted_at IS NULL;
