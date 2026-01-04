-- Migration 005: Enhance Goals table with all fields from BACKEND_API.md

-- Add missing columns to goals table
ALTER TABLE goals
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active',
ADD COLUMN IF NOT EXISTS metric_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN IF NOT EXISTS direction TEXT NOT NULL DEFAULT 'neutral',
ADD COLUMN IF NOT EXISTS unit TEXT,
ADD COLUMN IF NOT EXISTS initial_value DECIMAL(19,4),
ADD COLUMN IF NOT EXISTS target_value DECIMAL(19,4),
ADD COLUMN IF NOT EXISTS progress_target_value DECIMAL(19,4),
ADD COLUMN IF NOT EXISTS current_value DECIMAL(19,4) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS finance_mode TEXT,
ADD COLUMN IF NOT EXISTS currency TEXT,
ADD COLUMN IF NOT EXISTS linked_budget_id UUID REFERENCES budgets(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS linked_debt_id UUID REFERENCES debts(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS start_date DATE,
ADD COLUMN IF NOT EXISTS target_date DATE,
ADD COLUMN IF NOT EXISTS completed_date DATE,
ADD COLUMN IF NOT EXISTS progress_percent DECIMAL(5,2) NOT NULL DEFAULT 0;

-- Drop and recreate constraints for goals fields
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_show_status;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_status;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_type;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_metric_type;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_direction;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_finance_mode;
ALTER TABLE goals DROP CONSTRAINT IF EXISTS check_goal_progress_percent;

ALTER TABLE goals ADD CONSTRAINT check_goal_show_status CHECK (show_status IN ('active', 'archived', 'deleted'));
ALTER TABLE goals ADD CONSTRAINT check_goal_status CHECK (status IN ('active', 'paused', 'completed', 'archived'));
ALTER TABLE goals ADD CONSTRAINT check_goal_type CHECK (goal_type IN ('financial', 'health', 'education', 'productivity', 'personal'));
ALTER TABLE goals ADD CONSTRAINT check_goal_metric_type CHECK (metric_type IN ('none', 'amount', 'weight', 'count', 'duration', 'custom'));
ALTER TABLE goals ADD CONSTRAINT check_goal_direction CHECK (direction IN ('increase', 'decrease', 'neutral'));
ALTER TABLE goals ADD CONSTRAINT check_goal_finance_mode CHECK (finance_mode IS NULL OR finance_mode IN ('save', 'spend', 'debt_close'));
ALTER TABLE goals ADD CONSTRAINT check_goal_progress_percent CHECK (progress_percent >= 0 AND progress_percent <= 100);

-- Create goal_milestones table
CREATE TABLE IF NOT EXISTS goal_milestones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    target_percent DECIMAL(5,2) NOT NULL,
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    item_order INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT check_milestone_target_percent CHECK (target_percent >= 0 AND target_percent <= 100),
    CONSTRAINT check_milestone_order CHECK (item_order >= 0)
);

CREATE INDEX IF NOT EXISTS idx_milestones_goal_id ON goal_milestones(goal_id);
CREATE INDEX IF NOT EXISTS idx_milestones_order ON goal_milestones(goal_id, item_order);

-- Create goal_check_ins table
CREATE TABLE IF NOT EXISTS goal_check_ins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    value DECIMAL(19,4) NOT NULL,
    note TEXT,
    source_type TEXT NOT NULL,
    source_id UUID,
    date_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT check_checkin_source_type CHECK (source_type IN ('manual', 'task', 'habit', 'finance'))
);

CREATE INDEX IF NOT EXISTS idx_checkins_goal_id ON goal_check_ins(goal_id);
CREATE INDEX IF NOT EXISTS idx_checkins_date_key ON goal_check_ins(date_key);
CREATE INDEX IF NOT EXISTS idx_checkins_created_at ON goal_check_ins(created_at DESC);

-- Create goal_stats table (embedded stats)
CREATE TABLE IF NOT EXISTS goal_stats (
    goal_id UUID PRIMARY KEY REFERENCES goals(id) ON DELETE CASCADE,
    financial_progress_percent DECIMAL(5,2) DEFAULT 0,
    habits_progress_percent DECIMAL(5,2) DEFAULT 0,
    tasks_progress_percent DECIMAL(5,2) DEFAULT 0,
    focus_minutes_last_30 INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Create goal_habits junction table (many-to-many)
CREATE TABLE IF NOT EXISTS goal_habits (
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    PRIMARY KEY (goal_id, habit_id)
);

CREATE INDEX IF NOT EXISTS idx_goal_habits_goal ON goal_habits(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_habits_habit ON goal_habits(habit_id);

-- Additional indexes for goals table
CREATE INDEX IF NOT EXISTS idx_goals_show_status ON goals(show_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_goals_goal_type ON goals(goal_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_goals_target_date ON goals(target_date) WHERE deleted_at IS NULL AND show_status = 'active';
CREATE INDEX IF NOT EXISTS idx_goals_linked_budget ON goals(linked_budget_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_goals_linked_debt ON goals(linked_debt_id) WHERE deleted_at IS NULL;
