package goals

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
	plannerHabits "github.com/leora/leora-server/internal/modules/planner/habits"
	plannerTasks "github.com/leora/leora-server/internal/modules/planner/tasks"
	"github.com/lib/pq"
)

const goalSelectFields = `
	id, title, description, goal_type, status, show_status,
	metric_type, direction, unit, initial_value, target_value,
	progress_target_value, current_value, finance_mode, currency,
	linked_budget_id, linked_debt_id, start_date, target_date,
	completed_date, progress_percent, created_at, updated_at
`

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetStats(ctx context.Context, goalID string) (*GoalStats, error) {
	return r.getStats(ctx, goalID)
}

func (r *PostgresRepository) ListTasksByGoal(ctx context.Context, goalID string) ([]*TaskSummary, error) {
	repo := plannerTasks.NewPostgresRepository(r.db)
	return repo.List(ctx, plannerTasks.ListOptions{GoalID: goalID})
}

func (r *PostgresRepository) ListHabitsByGoal(ctx context.Context, goalID string) ([]*HabitSummary, error) {
	repo := plannerHabits.NewPostgresRepository(r.db)
	habits, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]*HabitSummary, 0, len(habits))
	for _, habit := range habits {
		if habit.GoalID != nil && *habit.GoalID == goalID {
			filtered = append(filtered, habit)
		}
	}
	return filtered, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Goal, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`SELECT %s FROM goals WHERE user_id = $1 AND deleted_at IS NULL AND show_status = 'active' ORDER BY created_at DESC`, goalSelectFields)
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var goals []*Goal
	for rows.Next() {
		var row goalRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		goals = append(goals, mapRowToGoal(row))
	}
	return goals, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Goal, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`SELECT %s FROM goals WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, goalSelectFields)
	goal, err := r.fetchGoal(ctx, query, id, userID)
	if err != nil {
		return nil, err
	}

	milestones, _ := r.getMilestones(ctx, id)
	goal.Milestones = milestones

	stats, _ := r.getStats(ctx, id)
	goal.Stats = stats

	return goal, nil
}

func (r *PostgresRepository) Create(ctx context.Context, goal *Goal) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if goal.ID == "" {
		goal.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	goal.CreatedAt = now
	goal.UpdatedAt = now

	if goal.ShowStatus == "" {
		goal.ShowStatus = "active"
	}
	if goal.MetricType == "" {
		goal.MetricType = "none"
	}
	if goal.Direction == "" {
		goal.Direction = "neutral"
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO goals (id, user_id, title, description, goal_type, status, show_status,
			metric_type, direction, unit, initial_value, target_value, progress_target_value,
			current_value, finance_mode, currency, linked_budget_id, linked_debt_id,
			start_date, target_date, completed_date, progress_percent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
	`, goal.ID, userID, goal.Title, goal.Description, goal.GoalType, goal.Status, goal.ShowStatus,
		goal.MetricType, goal.Direction, goal.Unit, goal.InitialValue, goal.TargetValue, goal.ProgressTargetValue,
		goal.CurrentValue, goal.FinanceMode, goal.Currency, goal.LinkedBudgetID, goal.LinkedDebtID,
		goal.StartDate, goal.TargetDate, goal.CompletedDate, goal.ProgressPercent, goal.CreatedAt, goal.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	if len(goal.Milestones) > 0 {
		_ = r.saveMilestones(ctx, goal.ID, goal.Milestones)
	}

	_ = r.createStats(ctx, goal.ID)
	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, goal *Goal) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	goal.UpdatedAt = utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE goals SET title=$1, description=$2, goal_type=$3, status=$4, show_status=$5,
			metric_type=$6, direction=$7, unit=$8, initial_value=$9, target_value=$10,
			progress_target_value=$11, current_value=$12, finance_mode=$13, currency=$14,
			linked_budget_id=$15, linked_debt_id=$16, start_date=$17, target_date=$18,
			completed_date=$19, progress_percent=$20, updated_at=$21
		WHERE id=$22 AND user_id=$23 AND deleted_at IS NULL
	`, goal.Title, goal.Description, goal.GoalType, goal.Status, goal.ShowStatus,
		goal.MetricType, goal.Direction, goal.Unit, goal.InitialValue, goal.TargetValue,
		goal.ProgressTargetValue, goal.CurrentValue, goal.FinanceMode, goal.Currency,
		goal.LinkedBudgetID, goal.LinkedDebtID, goal.StartDate, goal.TargetDate,
		goal.CompletedDate, goal.ProgressPercent, goal.UpdatedAt, goal.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return appErrors.GoalNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `UPDATE goals SET deleted_at=$1, updated_at=$2, show_status='deleted' WHERE id=$3 AND user_id=$4 AND deleted_at IS NULL`, now, now, id, userID)
	if err != nil {
		return appErrors.DatabaseError
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return appErrors.GoalNotFound
	}
	return nil
}

// BulkDelete soft deletes multiple goals and unlinks any associated budgets/debts
// Returns: (deletedCount, unlinkedBudgetIDs, unlinkedDebtIDs, error)
func (r *PostgresRepository) BulkDelete(ctx context.Context, ids []string) (int64, []string, []string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return 0, nil, nil, appErrors.InvalidToken
	}

	if len(ids) == 0 {
		return 0, nil, nil, nil
	}

	// First, get any linked budget/debt IDs for unlinking
	var linkedBudgetIDs []string
	var linkedDebtIDs []string
	rows, err := r.db.QueryxContext(ctx, `
		SELECT linked_budget_id, linked_debt_id
		FROM goals
		WHERE id = ANY($1) AND user_id = $2 AND deleted_at IS NULL
	`, pq.Array(ids), userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var budgetID, debtID sql.NullString
			rows.Scan(&budgetID, &debtID)
			if budgetID.Valid && budgetID.String != "" {
				linkedBudgetIDs = append(linkedBudgetIDs, budgetID.String)
			}
			if debtID.Valid && debtID.String != "" {
				linkedDebtIDs = append(linkedDebtIDs, debtID.String)
			}
		}
	}

	now := utils.NowUTC()
	// Soft delete the goals
	result, err := r.db.ExecContext(ctx, `
		UPDATE goals
		SET deleted_at = $1, updated_at = $2, show_status = 'deleted', linked_budget_id = NULL, linked_debt_id = NULL
		WHERE id = ANY($3) AND user_id = $4 AND deleted_at IS NULL
	`, now, now, pq.Array(ids), userID)
	if err != nil {
		return 0, nil, nil, appErrors.DatabaseError
	}

	deletedCount, _ := result.RowsAffected()

	// Unlink budgets (keep budget data, just remove the link)
	if len(linkedBudgetIDs) > 0 {
		r.db.ExecContext(ctx, `UPDATE budgets SET linked_goal_id = NULL, updated_at = $1 WHERE id = ANY($2)`, now, pq.Array(linkedBudgetIDs))
	}

	// Unlink debts (keep debt data, just remove the link)
	if len(linkedDebtIDs) > 0 {
		r.db.ExecContext(ctx, `UPDATE debts SET linked_goal_id = NULL, updated_at = $1 WHERE id = ANY($2)`, now, pq.Array(linkedDebtIDs))
	}

	return deletedCount, linkedBudgetIDs, linkedDebtIDs, nil
}

func (r *PostgresRepository) CreateCheckIn(ctx context.Context, checkIn *CheckIn) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	_, err := r.GetByID(ctx, checkIn.GoalID)
	if err != nil {
		return err
	}

	if checkIn.ID == "" {
		checkIn.ID = uuid.NewString()
	}
	checkIn.CreatedAt = utils.NowUTC()

	_, err = r.db.ExecContext(ctx, `INSERT INTO goal_check_ins (id, goal_id, value, note, source_type, source_id, date_key, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		checkIn.ID, checkIn.GoalID, checkIn.Value, checkIn.Note, checkIn.SourceType, checkIn.SourceID, checkIn.DateKey, checkIn.CreatedAt)
	if err != nil {
		return appErrors.DatabaseError
	}

	_ = r.updateGoalProgress(ctx, checkIn.GoalID, checkIn.Value)
	return nil
}

func (r *PostgresRepository) GetCheckIns(ctx context.Context, goalID string) ([]*CheckIn, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	_, err := r.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryxContext(ctx, `SELECT id, goal_id, value, note, source_type, source_id, date_key, created_at FROM goal_check_ins WHERE goal_id=$1 ORDER BY created_at DESC`, goalID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var checkIns []*CheckIn
	for rows.Next() {
		var checkIn CheckIn
		_ = rows.StructScan(&checkIn)
		checkIns = append(checkIns, &checkIn)
	}
	return checkIns, nil
}

func (r *PostgresRepository) fetchGoal(ctx context.Context, query string, args ...interface{}) (*Goal, error) {
	var row goalRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.GoalNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToGoal(row), nil
}

func (r *PostgresRepository) getMilestones(ctx context.Context, goalID string) ([]*Milestone, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT id, goal_id, title, description, target_percent, due_date, completed_at, item_order, created_at, updated_at FROM goal_milestones WHERE goal_id=$1 ORDER BY item_order ASC`, goalID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		var m Milestone
		_ = rows.StructScan(&m)
		milestones = append(milestones, &m)
	}
	return milestones, nil
}

func (r *PostgresRepository) saveMilestones(ctx context.Context, goalID string, milestones []*Milestone) error {
	r.db.ExecContext(ctx, `DELETE FROM goal_milestones WHERE goal_id=$1`, goalID)
	for i, m := range milestones {
		if m.ID == "" {
			m.ID = uuid.NewString()
		}
		m.GoalID = goalID
		m.Order = i
		now := utils.NowUTC()
		m.CreatedAt = now
		m.UpdatedAt = now
		r.db.ExecContext(ctx, `INSERT INTO goal_milestones (id, goal_id, title, description, target_percent, due_date, completed_at, item_order, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			m.ID, m.GoalID, m.Title, m.Description, m.TargetPercent, m.DueDate, m.CompletedAt, m.Order, m.CreatedAt, m.UpdatedAt)
	}
	return nil
}

func (r *PostgresRepository) getStats(ctx context.Context, goalID string) (*GoalStats, error) {
	var stats GoalStats
	err := r.db.GetContext(ctx, &stats, `
		SELECT
			COALESCE((SELECT COUNT(*) FROM tasks WHERE goal_id = $1 AND deleted_at IS NULL), 0) as total_tasks,
			COALESCE((SELECT COUNT(*) FROM tasks WHERE goal_id = $1 AND status = 'completed' AND deleted_at IS NULL), 0) as completed_tasks,
			COALESCE((SELECT COUNT(*) FROM habits WHERE goal_id = $1 AND deleted_at IS NULL), 0) as total_habits,
			COALESCE((SELECT SUM(actual_minutes) FROM focus_sessions WHERE goal_id = $1 AND deleted_at IS NULL), 0) as focus_minutes
	`, goalID)
	if err != nil {
		return nil, nil
	}
	return &stats, nil
}

func (r *PostgresRepository) createStats(ctx context.Context, goalID string) error {
	// deprecated: keep row to satisfy existing migrations, but stats are computed dynamically
	r.db.ExecContext(ctx, `INSERT INTO goal_stats (goal_id, financial_progress_percent, habits_progress_percent, tasks_progress_percent, focus_minutes_last_30, updated_at) VALUES ($1,0,0,0,0,$2)`, goalID, utils.NowUTC())
	return nil
}

func (r *PostgresRepository) updateGoalProgress(ctx context.Context, goalID string, addValue float64) error {
	var g goalRow
	err := r.db.GetContext(ctx, &g, `SELECT current_value, target_value, initial_value, direction FROM goals WHERE id=$1`, goalID)
	if err != nil {
		return err
	}

	newCurrent := g.CurrentValue + addValue
	newProgress := 0.0

	if g.TargetValue.Valid {
		initial := 0.0
		if g.InitialValue.Valid {
			initial = g.InitialValue.Float64
		}
		target := g.TargetValue.Float64

		// Handle decrease direction (e.g., weight loss)
		if g.Direction == "decrease" {
			if initial != target {
				newProgress = (initial - newCurrent) / (initial - target)
			} else if newCurrent <= target {
				newProgress = 1.0
			}
		} else {
			// Default: increase direction
			if target > initial {
				newProgress = (newCurrent - initial) / (target - initial)
			} else if newCurrent >= target {
				newProgress = 1.0
			}
		}

		// Clamp between 0 and 1
		if newProgress > 1 {
			newProgress = 1
		}
		if newProgress < 0 {
			newProgress = 0
		}
	}

	r.db.ExecContext(ctx, `UPDATE goals SET current_value=$1, progress_percent=$2, updated_at=$3 WHERE id=$4`, newCurrent, newProgress, utils.NowUTC(), goalID)
	return nil
}

type goalRow struct {
	ID                  string          `db:"id"`
	Title               string          `db:"title"`
	Description         sql.NullString  `db:"description"`
	GoalType            string          `db:"goal_type"`
	Status              string          `db:"status"`
	ShowStatus          string          `db:"show_status"`
	MetricType          string          `db:"metric_type"`
	Direction           string          `db:"direction"`
	Unit                sql.NullString  `db:"unit"`
	InitialValue        sql.NullFloat64 `db:"initial_value"`
	TargetValue         sql.NullFloat64 `db:"target_value"`
	ProgressTargetValue sql.NullFloat64 `db:"progress_target_value"`
	CurrentValue        float64         `db:"current_value"`
	FinanceMode         sql.NullString  `db:"finance_mode"`
	Currency            sql.NullString  `db:"currency"`
	LinkedBudgetID      sql.NullString  `db:"linked_budget_id"`
	LinkedDebtID        sql.NullString  `db:"linked_debt_id"`
	StartDate           sql.NullString  `db:"start_date"`
	TargetDate          sql.NullString  `db:"target_date"`
	CompletedDate       sql.NullString  `db:"completed_date"`
	ProgressPercent     float64         `db:"progress_percent"`
	CreatedAt           string          `db:"created_at"`
	UpdatedAt           string          `db:"updated_at"`
}

// clampProgress ensures progress is between 0 and 1
func clampProgress(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

func mapRowToGoal(row goalRow) *Goal {
	goal := &Goal{
		ID:              row.ID,
		Title:           row.Title,
		GoalType:        row.GoalType,
		Status:          row.Status,
		ShowStatus:      row.ShowStatus,
		MetricType:      row.MetricType,
		Direction:       row.Direction,
		CurrentValue:    row.CurrentValue,
		ProgressPercent: clampProgress(row.ProgressPercent), // Clamp to 0-1 range
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.Description.Valid {
		goal.Description = &row.Description.String
	}
	if row.Unit.Valid {
		goal.Unit = &row.Unit.String
	}
	if row.InitialValue.Valid {
		goal.InitialValue = &row.InitialValue.Float64
	}
	if row.TargetValue.Valid {
		goal.TargetValue = &row.TargetValue.Float64
	}
	if row.ProgressTargetValue.Valid {
		goal.ProgressTargetValue = &row.ProgressTargetValue.Float64
	}
	if row.FinanceMode.Valid {
		goal.FinanceMode = &row.FinanceMode.String
	}
	if row.Currency.Valid {
		goal.Currency = &row.Currency.String
	}
	if row.LinkedBudgetID.Valid {
		goal.LinkedBudgetID = &row.LinkedBudgetID.String
	}
	if row.LinkedDebtID.Valid {
		goal.LinkedDebtID = &row.LinkedDebtID.String
	}
	if row.StartDate.Valid {
		goal.StartDate = &row.StartDate.String
	}
	if row.TargetDate.Valid {
		goal.TargetDate = &row.TargetDate.String
	}
	if row.CompletedDate.Valid {
		goal.CompletedDate = &row.CompletedDate.String
	}
	return goal
}

// UpdateBudgetGoalLink updates the linked_goal_id in the budgets table (bidirectional link)
func (r *PostgresRepository) UpdateBudgetGoalLink(ctx context.Context, budgetID, goalID string) error {
	var query string
	var err error
	if goalID == "" {
		query = `UPDATE budgets SET linked_goal_id = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
		_, err = r.db.ExecContext(ctx, query, utils.NowUTC(), budgetID)
	} else {
		query = `UPDATE budgets SET linked_goal_id = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		_, err = r.db.ExecContext(ctx, query, goalID, utils.NowUTC(), budgetID)
	}
	return err
}

// UpdateDebtGoalLink updates the linked_goal_id in the debts table (bidirectional link)
func (r *PostgresRepository) UpdateDebtGoalLink(ctx context.Context, debtID, goalID string) error {
	var query string
	var err error
	if goalID == "" {
		query = `UPDATE debts SET linked_goal_id = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
		_, err = r.db.ExecContext(ctx, query, utils.NowUTC(), debtID)
	} else {
		query = `UPDATE debts SET linked_goal_id = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		_, err = r.db.ExecContext(ctx, query, goalID, utils.NowUTC(), debtID)
	}
	return err
}

// GetBudgetProgress retrieves budget spending information for finance progress calculation
func (r *PostgresRepository) GetBudgetProgress(ctx context.Context, budgetID string) (*BudgetProgress, error) {
	var progress BudgetProgress
	err := r.db.GetContext(ctx, &progress, `
		SELECT
			COALESCE(limit_amount, 0) as limit_amount,
			COALESCE((
				SELECT SUM(amount)
				FROM transactions
				WHERE budget_id = $1
				AND type = 'expense'
				AND deleted_at IS NULL
			), 0) as spent_amount
		FROM budgets
		WHERE id = $1 AND deleted_at IS NULL
	`, budgetID)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetDebtProgress retrieves debt payment information for finance progress calculation
func (r *PostgresRepository) GetDebtProgress(ctx context.Context, debtID string) (*DebtProgress, error) {
	var progress DebtProgress
	err := r.db.GetContext(ctx, &progress, `
		SELECT
			COALESCE(principal_amount, 0) as principal_amount,
			COALESCE((
				SELECT SUM(converted_amount_to_debt)
				FROM debt_payments
				WHERE debt_id = $1
				AND deleted_at IS NULL
			), 0) as paid_amount
		FROM debts
		WHERE id = $1 AND deleted_at IS NULL
	`, debtID)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}
