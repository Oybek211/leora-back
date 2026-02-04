-- Finance-Planner Integration: Add linking columns
-- This migration adds bidirectional links between Finance and Planner modules

-- Budget -> Goal link (bidirectional with Goal.linkedBudgetId)
ALTER TABLE budgets ADD COLUMN IF NOT EXISTS linked_goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_budgets_linked_goal ON budgets(linked_goal_id) WHERE linked_goal_id IS NOT NULL AND deleted_at IS NULL;

-- Debt -> Goal link (bidirectional with Goal.linkedDebtId)
ALTER TABLE debts ADD COLUMN IF NOT EXISTS linked_goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_debts_linked_goal ON debts(linked_goal_id) WHERE linked_goal_id IS NOT NULL AND deleted_at IS NULL;

-- Debt -> Budget link
ALTER TABLE debts ADD COLUMN IF NOT EXISTS linked_budget_id UUID REFERENCES budgets(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_debts_linked_budget ON debts(linked_budget_id) WHERE linked_budget_id IS NOT NULL AND deleted_at IS NULL;

-- Transaction -> Budget link (for budget tracking)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS budget_id UUID REFERENCES budgets(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_budget ON transactions(budget_id) WHERE budget_id IS NOT NULL AND deleted_at IS NULL;

-- Transaction -> Habit link (for habit finance rules)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS habit_id UUID REFERENCES habits(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_habit ON transactions(habit_id) WHERE habit_id IS NOT NULL AND deleted_at IS NULL;

-- Transaction -> Goal link (if not exists)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS linked_goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_linked_goal ON transactions(linked_goal_id) WHERE linked_goal_id IS NOT NULL AND deleted_at IS NULL;

-- Transaction -> Debt link (if not exists)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS linked_debt_id UUID REFERENCES debts(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_linked_debt ON transactions(linked_debt_id) WHERE linked_debt_id IS NOT NULL AND deleted_at IS NULL;
