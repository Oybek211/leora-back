package habits

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/lib/pq"
)

const habitSelectFields = `
	id, title, description, icon_id, habit_type, status, show_status,
	goal_id, frequency, days_of_week, times_per_week, time_of_day,
	completion_mode, target_per_day, unit, counting_type, difficulty,
	priority, challenge_length_days, reminder_enabled, reminder_time,
	streak_current, streak_best, completion_rate_30d, finance_rule,
	created_at, updated_at
`

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Habit, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`SELECT %s FROM habits WHERE user_id = $1 AND deleted_at IS NULL AND show_status = 'active' ORDER BY created_at DESC`, habitSelectFields)
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var habits []*Habit
	for rows.Next() {
		var row habitRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		habit := mapRowToHabit(row)

		// Fetch linked goal IDs
		linkedGoalIDs, _ := r.getLinkedGoalIDs(ctx, habit.ID)
		habit.LinkedGoalIDs = linkedGoalIDs

		habits = append(habits, habit)
	}
	return habits, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Habit, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`SELECT %s FROM habits WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, habitSelectFields)
	habit, err := r.fetchHabit(ctx, query, id, userID)
	if err != nil {
		return nil, err
	}

	// Fetch linked goal IDs
	linkedGoalIDs, _ := r.getLinkedGoalIDs(ctx, id)
	habit.LinkedGoalIDs = linkedGoalIDs

	return habit, nil
}

func (r *PostgresRepository) Create(ctx context.Context, habit *Habit) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if habit.ID == "" {
		habit.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	habit.CreatedAt = now
	habit.UpdatedAt = now

	// Set defaults
	if habit.ShowStatus == "" {
		habit.ShowStatus = "active"
	}
	if habit.Frequency == "" {
		habit.Frequency = "daily"
	}
	if habit.CompletionMode == "" {
		habit.CompletionMode = "boolean"
	}
	if habit.CountingType == "" {
		habit.CountingType = "create"
	}
	if habit.Difficulty == "" {
		habit.Difficulty = "medium"
	}
	if habit.Priority == "" {
		habit.Priority = "medium"
	}

	// Serialize FinanceRule to JSON - use nil for null in database
	var financeRuleJSON interface{}
	if habit.FinanceRule != nil {
		jsonBytes, _ := json.Marshal(habit.FinanceRule)
		financeRuleJSON = jsonBytes
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO habits (id, user_id, title, description, icon_id, habit_type, status, show_status,
			goal_id, frequency, days_of_week, times_per_week, time_of_day, completion_mode,
			target_per_day, unit, counting_type, difficulty, priority, challenge_length_days,
			reminder_enabled, reminder_time, streak_current, streak_best, completion_rate_30d,
			finance_rule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
	`, habit.ID, userID, habit.Title, habit.Description, habit.IconID, habit.HabitType, habit.Status, habit.ShowStatus,
		habit.GoalID, habit.Frequency, pq.Array(habit.DaysOfWeek), habit.TimesPerWeek, habit.TimeOfDay, habit.CompletionMode,
		habit.TargetPerDay, habit.Unit, habit.CountingType, habit.Difficulty, habit.Priority, habit.ChallengeLengthDays,
		habit.ReminderEnabled, habit.ReminderTime, habit.StreakCurrent, habit.StreakBest, habit.CompletionRate30d,
		financeRuleJSON, habit.CreatedAt, habit.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	// Save linked goal IDs
	if len(habit.LinkedGoalIDs) > 0 {
		_ = r.saveLinkedGoalIDs(ctx, habit.ID, habit.LinkedGoalIDs)
	}

	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, habit *Habit) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	habit.UpdatedAt = utils.NowUTC()

	// Serialize FinanceRule to JSON - use nil for null in database
	var financeRuleJSON interface{}
	if habit.FinanceRule != nil {
		jsonBytes, _ := json.Marshal(habit.FinanceRule)
		financeRuleJSON = jsonBytes
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE habits SET title=$1, description=$2, icon_id=$3, habit_type=$4, status=$5, show_status=$6,
			goal_id=$7, frequency=$8, days_of_week=$9, times_per_week=$10, time_of_day=$11, completion_mode=$12,
			target_per_day=$13, unit=$14, counting_type=$15, difficulty=$16, priority=$17, challenge_length_days=$18,
			reminder_enabled=$19, reminder_time=$20, streak_current=$21, streak_best=$22, completion_rate_30d=$23,
			finance_rule=$24, updated_at=$25
		WHERE id=$26 AND user_id=$27 AND deleted_at IS NULL
	`, habit.Title, habit.Description, habit.IconID, habit.HabitType, habit.Status, habit.ShowStatus,
		habit.GoalID, habit.Frequency, pq.Array(habit.DaysOfWeek), habit.TimesPerWeek, habit.TimeOfDay, habit.CompletionMode,
		habit.TargetPerDay, habit.Unit, habit.CountingType, habit.Difficulty, habit.Priority, habit.ChallengeLengthDays,
		habit.ReminderEnabled, habit.ReminderTime, habit.StreakCurrent, habit.StreakBest, habit.CompletionRate30d,
		financeRuleJSON, habit.UpdatedAt, habit.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return appErrors.HabitNotFound
	}

	// Update linked goal IDs
	_ = r.saveLinkedGoalIDs(ctx, habit.ID, habit.LinkedGoalIDs)

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `UPDATE habits SET deleted_at=$1, updated_at=$2, show_status='deleted' WHERE id=$3 AND user_id=$4 AND deleted_at IS NULL`, now, now, id, userID)
	if err != nil {
		return appErrors.DatabaseError
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return appErrors.HabitNotFound
	}
	return nil
}

// BulkDelete soft deletes multiple habits by their IDs
func (r *PostgresRepository) BulkDelete(ctx context.Context, ids []string) (int64, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return 0, appErrors.InvalidToken
	}

	if len(ids) == 0 {
		return 0, nil
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE habits
		SET deleted_at = $1, updated_at = $2, show_status = 'deleted'
		WHERE id = ANY($3) AND user_id = $4 AND deleted_at IS NULL
	`, now, now, pq.Array(ids), userID)
	if err != nil {
		return 0, appErrors.DatabaseError
	}

	rows, _ := result.RowsAffected()
	return rows, nil
}

// ToggleCompletion toggles the completion status for a habit on a specific date
func (r *PostgresRepository) ToggleCompletion(ctx context.Context, habitID, dateKey string) (*HabitCompletion, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	// Verify habit belongs to user
	_, err := r.GetByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	// Check current completion status
	var existing struct {
		ID     string `db:"id"`
		Status string `db:"status"`
	}
	err = r.db.GetContext(ctx, &existing, `
		SELECT id, status
		FROM habit_completions
		WHERE habit_id = $1 AND date_key = $2
	`, habitID, dateKey)

	now := utils.NowUTC()
	completion := &HabitCompletion{
		HabitID:   habitID,
		DateKey:   dateKey,
		CreatedAt: now,
	}

	if err == sql.ErrNoRows {
		// No completion exists - create as 'done'
		completion.ID = uuid.NewString()
		completion.Status = "done"
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO habit_completions (id, habit_id, date_key, status, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, completion.ID, completion.HabitID, completion.DateKey, completion.Status, completion.CreatedAt)
	} else if err != nil {
		return nil, appErrors.DatabaseError
	} else {
		// Completion exists - toggle status
		completion.ID = existing.ID
		if existing.Status == "done" {
			completion.Status = "miss"
		} else {
			completion.Status = "done"
		}
		_, err = r.db.ExecContext(ctx, `
			UPDATE habit_completions SET status = $1 WHERE id = $2
		`, completion.Status, completion.ID)
	}

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	// Update streak and completion rate
	_ = r.updateHabitStats(ctx, habitID)

	return completion, nil
}

func (r *PostgresRepository) CreateCompletion(ctx context.Context, completion *HabitCompletion) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	// Verify habit belongs to user
	_, err := r.GetByID(ctx, completion.HabitID)
	if err != nil {
		return err
	}

	if completion.ID == "" {
		completion.ID = uuid.NewString()
	}
	completion.CreatedAt = utils.NowUTC()

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO habit_completions (id, habit_id, date_key, status, value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (habit_id, date_key) DO UPDATE SET status = $4, value = $5
	`, completion.ID, completion.HabitID, completion.DateKey, completion.Status, completion.Value, completion.CreatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	// Update streak and completion rate
	_ = r.updateHabitStats(ctx, completion.HabitID)

	return nil
}

func (r *PostgresRepository) GetCompletions(ctx context.Context, habitID string) ([]*HabitCompletion, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	// Verify habit belongs to user
	_, err := r.GetByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryxContext(ctx, `
		SELECT id, habit_id, date_key, status, value, TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
		FROM habit_completions
		WHERE habit_id = $1
		ORDER BY date_key DESC
	`, habitID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var completions []*HabitCompletion
	for rows.Next() {
		var completion HabitCompletion
		if err := rows.StructScan(&completion); err != nil {
			continue // Skip rows with scan errors
		}
		completions = append(completions, &completion)
	}
	return completions, nil
}

func (r *PostgresRepository) GetStats(ctx context.Context, habitID string) (*HabitStats, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	// Verify habit belongs to user
	habit, err := r.GetByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	// Count completions
	var totalCompletions, totalMisses int
	r.db.GetContext(ctx, &totalCompletions, `SELECT COUNT(*) FROM habit_completions WHERE habit_id = $1 AND status = 'done'`, habitID)
	r.db.GetContext(ctx, &totalMisses, `SELECT COUNT(*) FROM habit_completions WHERE habit_id = $1 AND status = 'miss'`, habitID)

	return &HabitStats{
		StreakCurrent:     habit.StreakCurrent,
		StreakBest:        habit.StreakBest,
		CompletionRate30d: habit.CompletionRate30d,
		TotalCompletions:  totalCompletions,
		TotalMisses:       totalMisses,
	}, nil
}

func (r *PostgresRepository) fetchHabit(ctx context.Context, query string, args ...interface{}) (*Habit, error) {
	var row habitRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.HabitNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToHabit(row), nil
}

func (r *PostgresRepository) getLinkedGoalIDs(ctx context.Context, habitID string) ([]string, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT goal_id FROM habit_goals WHERE habit_id = $1`, habitID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var goalIDs []string
	for rows.Next() {
		var goalID string
		_ = rows.Scan(&goalID)
		goalIDs = append(goalIDs, goalID)
	}
	return goalIDs, nil
}

func (r *PostgresRepository) saveLinkedGoalIDs(ctx context.Context, habitID string, goalIDs []string) error {
	// Delete existing links
	r.db.ExecContext(ctx, `DELETE FROM habit_goals WHERE habit_id = $1`, habitID)

	// Insert new links
	for _, goalID := range goalIDs {
		r.db.ExecContext(ctx, `INSERT INTO habit_goals (habit_id, goal_id, created_at) VALUES ($1, $2, $3)`, habitID, goalID, utils.NowUTC())
	}
	return nil
}

func (r *PostgresRepository) updateHabitStats(ctx context.Context, habitID string) error {
	// This is a simplified version - in production you'd want more complex logic
	// to calculate current streak, best streak, and 30-day completion rate

	// Calculate 30-day completion rate
	// date_key is stored as 'YYYY-MM-DD' text, so we compare as DATE
	var doneCount, totalCount int
	r.db.GetContext(ctx, &doneCount, `
		SELECT COUNT(*) FROM habit_completions
		WHERE habit_id = $1 AND date_key::DATE >= (CURRENT_DATE - INTERVAL '30 days') AND status = 'done'
	`, habitID)
	r.db.GetContext(ctx, &totalCount, `
		SELECT COUNT(*) FROM habit_completions
		WHERE habit_id = $1 AND date_key::DATE >= (CURRENT_DATE - INTERVAL '30 days')
	`, habitID)

	var completionRate float64
	if totalCount > 0 {
		completionRate = (float64(doneCount) / float64(totalCount)) * 100
	}

	// Update the habit
	r.db.ExecContext(ctx, `UPDATE habits SET completion_rate_30d = $1, updated_at = $2 WHERE id = $3`, completionRate, utils.NowUTC(), habitID)

	return nil
}

// GetTransactionCountForHabit counts transactions linked to this habit for a given date
func (r *PostgresRepository) GetTransactionCountForHabit(ctx context.Context, habitID string, dateKey string) (int, float64, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return 0, 0, appErrors.InvalidToken
	}

	// Verify habit belongs to user
	_, err := r.GetByID(ctx, habitID)
	if err != nil {
		return 0, 0, err
	}

	var count int
	var totalAmount float64

	// Count transactions linked to this habit on the given date
	err = r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM transactions
		WHERE habit_id = $1 AND user_id = $2 AND DATE(created_at) = $3 AND deleted_at IS NULL
	`, habitID, userID, dateKey)
	if err != nil {
		return 0, 0, appErrors.DatabaseError
	}

	// Sum transaction amounts
	r.db.GetContext(ctx, &totalAmount, `
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE habit_id = $1 AND user_id = $2 AND DATE(created_at) = $3 AND deleted_at IS NULL
	`, habitID, userID, dateKey)

	return count, totalAmount, nil
}

type habitRow struct {
	ID                  string          `db:"id"`
	Title               string          `db:"title"`
	Description         sql.NullString  `db:"description"`
	IconID              sql.NullString  `db:"icon_id"`
	HabitType           string          `db:"habit_type"`
	Status              string          `db:"status"`
	ShowStatus          string          `db:"show_status"`
	GoalID              sql.NullString  `db:"goal_id"`
	Frequency           string          `db:"frequency"`
	DaysOfWeek          pq.Int64Array   `db:"days_of_week"`
	TimesPerWeek        sql.NullInt64   `db:"times_per_week"`
	TimeOfDay           sql.NullString  `db:"time_of_day"`
	CompletionMode      string          `db:"completion_mode"`
	TargetPerDay        sql.NullFloat64 `db:"target_per_day"`
	Unit                sql.NullString  `db:"unit"`
	CountingType        string          `db:"counting_type"`
	Difficulty          string          `db:"difficulty"`
	Priority            string          `db:"priority"`
	ChallengeLengthDays sql.NullInt64   `db:"challenge_length_days"`
	ReminderEnabled     bool            `db:"reminder_enabled"`
	ReminderTime        sql.NullString  `db:"reminder_time"`
	StreakCurrent       int             `db:"streak_current"`
	StreakBest          int             `db:"streak_best"`
	CompletionRate30d   float64         `db:"completion_rate_30d"`
	FinanceRule         []byte          `db:"finance_rule"`
	CreatedAt           string          `db:"created_at"`
	UpdatedAt           string          `db:"updated_at"`
}

func mapRowToHabit(row habitRow) *Habit {
	habit := &Habit{
		ID:                row.ID,
		Title:             row.Title,
		HabitType:         row.HabitType,
		Status:            row.Status,
		ShowStatus:        row.ShowStatus,
		Frequency:         row.Frequency,
		CompletionMode:    row.CompletionMode,
		CountingType:      row.CountingType,
		Difficulty:        row.Difficulty,
		Priority:          row.Priority,
		ReminderEnabled:   row.ReminderEnabled,
		StreakCurrent:     row.StreakCurrent,
		StreakBest:        row.StreakBest,
		CompletionRate30d: row.CompletionRate30d,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}

	if row.Description.Valid {
		habit.Description = &row.Description.String
	}
	if row.IconID.Valid {
		habit.IconID = &row.IconID.String
	}
	if row.GoalID.Valid {
		habit.GoalID = &row.GoalID.String
	}
	if row.TimesPerWeek.Valid {
		timesPerWeek := int(row.TimesPerWeek.Int64)
		habit.TimesPerWeek = &timesPerWeek
	}
	if row.TimeOfDay.Valid {
		habit.TimeOfDay = &row.TimeOfDay.String
	}
	if row.TargetPerDay.Valid {
		habit.TargetPerDay = &row.TargetPerDay.Float64
	}
	if row.Unit.Valid {
		habit.Unit = &row.Unit.String
	}
	if row.ChallengeLengthDays.Valid {
		challengeLengthDays := int(row.ChallengeLengthDays.Int64)
		habit.ChallengeLengthDays = &challengeLengthDays
	}
	if row.ReminderTime.Valid {
		habit.ReminderTime = &row.ReminderTime.String
	}

	// Parse DaysOfWeek array
	if len(row.DaysOfWeek) > 0 {
		habit.DaysOfWeek = make([]int, len(row.DaysOfWeek))
		for i, v := range row.DaysOfWeek {
			habit.DaysOfWeek[i] = int(v)
		}
	}

	// Parse FinanceRule JSON
	if len(row.FinanceRule) > 0 {
		var financeRule FinanceRule
		if err := json.Unmarshal(row.FinanceRule, &financeRule); err == nil {
			habit.FinanceRule = &financeRule
		}
	}

	return habit
}
