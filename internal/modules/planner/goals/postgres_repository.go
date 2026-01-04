package goals

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
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
	err := r.db.GetContext(ctx, &stats, `SELECT financial_progress_percent, habits_progress_percent, tasks_progress_percent, focus_minutes_last_30, updated_at FROM goal_stats WHERE goal_id=$1`, goalID)
	if err != nil {
		return nil, nil
	}
	return &stats, nil
}

func (r *PostgresRepository) createStats(ctx context.Context, goalID string) error {
	r.db.ExecContext(ctx, `INSERT INTO goal_stats (goal_id, financial_progress_percent, habits_progress_percent, tasks_progress_percent, focus_minutes_last_30, updated_at) VALUES ($1,0,0,0,0,$2)`, goalID, utils.NowUTC())
	return nil
}

func (r *PostgresRepository) updateGoalProgress(ctx context.Context, goalID string, addValue float64) error {
	var g goalRow
	err := r.db.GetContext(ctx, &g, `SELECT current_value, target_value, initial_value FROM goals WHERE id=$1`, goalID)
	if err != nil {
		return err
	}

	newCurrent := g.CurrentValue + addValue
	newProgress := 0.0
	if g.TargetValue.Valid && g.TargetValue.Float64 > 0 {
		initial := 0.0
		if g.InitialValue.Valid {
			initial = g.InitialValue.Float64
		}
		if g.TargetValue.Float64 > initial {
			newProgress = ((newCurrent - initial) / (g.TargetValue.Float64 - initial)) * 100
		}
		if newProgress > 100 {
			newProgress = 100
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
		ProgressPercent: row.ProgressPercent,
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
