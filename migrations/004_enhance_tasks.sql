-- Migration 004: Enhance Tasks table with all fields from BACKEND_API.md

-- First, update existing status values to match new constraints
UPDATE tasks SET status = 'inbox' WHERE status = 'pending';
UPDATE tasks SET priority = 'medium' WHERE priority NOT IN ('low', 'medium', 'high');

-- Add missing columns to tasks table
ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS show_status TEXT NOT NULL DEFAULT 'active',
ADD COLUMN IF NOT EXISTS goal_id UUID REFERENCES goals(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS habit_id UUID REFERENCES habits(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS finance_link TEXT,
ADD COLUMN IF NOT EXISTS progress_value DECIMAL(19,4),
ADD COLUMN IF NOT EXISTS progress_unit TEXT,
ADD COLUMN IF NOT EXISTS due_date DATE,
ADD COLUMN IF NOT EXISTS start_date DATE,
ADD COLUMN IF NOT EXISTS time_of_day TEXT,
ADD COLUMN IF NOT EXISTS estimated_minutes INTEGER,
ADD COLUMN IF NOT EXISTS energy_level INTEGER,
ADD COLUMN IF NOT EXISTS context TEXT,
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS last_focus_session_id UUID REFERENCES focus_sessions(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS focus_total_minutes INTEGER NOT NULL DEFAULT 0;

-- Add constraints for task fields
ALTER TABLE tasks ADD CONSTRAINT check_task_show_status CHECK (show_status IN ('active', 'archived', 'deleted'));
ALTER TABLE tasks ADD CONSTRAINT check_task_status CHECK (status IN ('inbox', 'planned', 'in_progress', 'completed', 'canceled', 'moved', 'overdue'));
ALTER TABLE tasks ADD CONSTRAINT check_task_priority CHECK (priority IN ('low', 'medium', 'high'));
ALTER TABLE tasks ADD CONSTRAINT check_task_finance_link CHECK (finance_link IS NULL OR finance_link IN ('record_expenses', 'pay_debt', 'review_budget', 'transfer_money', 'none'));
ALTER TABLE tasks ADD CONSTRAINT check_task_energy_level CHECK (energy_level IS NULL OR (energy_level >= 1 AND energy_level <= 3));

-- Create task_checklist_items table
CREATE TABLE IF NOT EXISTS task_checklist_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    item_order INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT check_item_order CHECK (item_order >= 0)
);

CREATE INDEX IF NOT EXISTS idx_checklist_task_id ON task_checklist_items(task_id);
CREATE INDEX IF NOT EXISTS idx_checklist_order ON task_checklist_items(task_id, item_order);

-- Create task_dependencies table
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT check_dependency_status CHECK (status IN ('pending', 'met')),
    CONSTRAINT check_no_self_dependency CHECK (task_id != depends_on_task_id),
    UNIQUE(task_id, depends_on_task_id)
);

CREATE INDEX IF NOT EXISTS idx_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_depends_on ON task_dependencies(depends_on_task_id);

-- Add additional indexes for tasks table
CREATE INDEX IF NOT EXISTS idx_tasks_goal_id ON tasks(goal_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_habit_id ON tasks(habit_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date) WHERE deleted_at IS NULL AND show_status = 'active';
CREATE INDEX IF NOT EXISTS idx_tasks_show_status ON tasks(show_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_context ON tasks(context) WHERE deleted_at IS NULL AND context IS NOT NULL;
