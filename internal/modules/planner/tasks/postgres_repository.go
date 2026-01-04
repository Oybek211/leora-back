package tasks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const taskSelectFields = `
	id,
	title,
	status,
	show_status,
	priority,
	goal_id,
	habit_id,
	finance_link,
	progress_value,
	progress_unit,
	due_date,
	start_date,
	time_of_day,
	estimated_minutes,
	energy_level,
	context,
	notes,
	last_focus_session_id,
	focus_total_minutes,
	created_at,
	updated_at
`

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, opts ListOptions) ([]*Task, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	whereClauses := []string{"user_id = $1", "deleted_at IS NULL"}
	args := []interface{}{userID}
	argCount := 1

	// Filter by showStatus (default: active)
	if opts.ShowStatus != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("show_status = $%d", argCount))
		args = append(args, opts.ShowStatus)
	} else {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("show_status = $%d", argCount))
		args = append(args, "active")
	}

	// Filter by status
	if opts.Status != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, opts.Status)
	}

	// Filter by priority
	if opts.Priority != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("priority = $%d", argCount))
		args = append(args, opts.Priority)
	}

	// Filter by goalId
	if opts.GoalID != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("goal_id = $%d", argCount))
		args = append(args, opts.GoalID)
	}

	// Filter by habitId
	if opts.HabitID != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("habit_id = $%d", argCount))
		args = append(args, opts.HabitID)
	}

	// Filter by exact dueDate
	if opts.DueDate != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("due_date = $%d", argCount))
		args = append(args, opts.DueDate)
	}

	// Filter by dueDate range
	if opts.DueDateFrom != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("due_date >= $%d", argCount))
		args = append(args, opts.DueDateFrom)
	}
	if opts.DueDateTo != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("due_date <= $%d", argCount))
		args = append(args, opts.DueDateTo)
	}

	// Search by title
	if opts.Search != "" {
		argCount++
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(title) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+opts.Search+"%")
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if opts.SortBy != "" {
		sortField := opts.SortBy
		switch opts.SortBy {
		case "dueDate":
			sortField = "due_date"
		case "priority":
			sortField = "priority"
		case "createdAt":
			sortField = "created_at"
		default:
			sortField = "created_at"
		}

		sortOrder := "DESC"
		if opts.SortOrder == "asc" {
			sortOrder = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", sortField, sortOrder)
	}

	whereClause := fmt.Sprintf("WHERE %s", joinClauses(whereClauses))
	query := fmt.Sprintf(`
		SELECT %s FROM tasks
		%s
		ORDER BY %s
	`, taskSelectFields, whereClause, orderBy)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var row taskRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		tasks = append(tasks, mapRowToTask(row))
	}

	return tasks, nil
}

func joinClauses(clauses []string) string {
	result := ""
	for i, clause := range clauses {
		if i > 0 {
			result += " AND "
		}
		result += clause
	}
	return result
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Task, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM tasks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, taskSelectFields)

	task, err := r.fetchTask(ctx, query, id, userID)
	if err != nil {
		return nil, err
	}

	// Fetch checklist items
	checklist, err := r.getChecklistItems(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Checklist = checklist

	return task, nil
}

func (r *PostgresRepository) Create(ctx context.Context, task *Task) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if task.ID == "" {
		task.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Set default values
	if task.ShowStatus == "" {
		task.ShowStatus = "active"
	}
	if task.FocusTotalMinutes == 0 {
		task.FocusTotalMinutes = 0
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (
			id, user_id, title, status, show_status, priority,
			goal_id, habit_id, finance_link, progress_value, progress_unit,
			due_date, start_date, time_of_day, estimated_minutes, energy_level,
			context, notes, last_focus_session_id, focus_total_minutes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
	`, task.ID, userID, task.Title, task.Status, task.ShowStatus, task.Priority,
		task.GoalID, task.HabitID, task.FinanceLink, task.ProgressValue, task.ProgressUnit,
		task.DueDate, task.StartDate, task.TimeOfDay, task.EstimatedMinutes, task.EnergyLevel,
		task.Context, task.Notes, task.LastFocusSessionID, task.FocusTotalMinutes,
		task.CreatedAt, task.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	// Save checklist items if provided
	if len(task.Checklist) > 0 {
		if err := r.saveChecklistItems(ctx, task.ID, task.Checklist); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, task *Task) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	task.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = $1, status = $2, show_status = $3, priority = $4,
			goal_id = $5, habit_id = $6, finance_link = $7,
			progress_value = $8, progress_unit = $9,
			due_date = $10, start_date = $11, time_of_day = $12,
			estimated_minutes = $13, energy_level = $14, context = $15,
			notes = $16, last_focus_session_id = $17, focus_total_minutes = $18,
			updated_at = $19
		WHERE id = $20 AND user_id = $21 AND deleted_at IS NULL
	`, task.Title, task.Status, task.ShowStatus, task.Priority,
		task.GoalID, task.HabitID, task.FinanceLink,
		task.ProgressValue, task.ProgressUnit,
		task.DueDate, task.StartDate, task.TimeOfDay,
		task.EstimatedMinutes, task.EnergyLevel, task.Context,
		task.Notes, task.LastFocusSessionID, task.FocusTotalMinutes,
		task.UpdatedAt, task.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.TaskNotFound
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET deleted_at = $1, updated_at = $2, show_status = 'deleted'
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
	`, now, now, id, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.TaskNotFound
	}

	return nil
}

func (r *PostgresRepository) fetchTask(ctx context.Context, query string, args ...interface{}) (*Task, error) {
	var row taskRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.TaskNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToTask(row), nil
}

type taskRow struct {
	ID                 string          `db:"id"`
	Title              string          `db:"title"`
	Status             string          `db:"status"`
	ShowStatus         string          `db:"show_status"`
	Priority           string          `db:"priority"`
	GoalID             sql.NullString  `db:"goal_id"`
	HabitID            sql.NullString  `db:"habit_id"`
	FinanceLink        sql.NullString  `db:"finance_link"`
	ProgressValue      sql.NullFloat64 `db:"progress_value"`
	ProgressUnit       sql.NullString  `db:"progress_unit"`
	DueDate            sql.NullString  `db:"due_date"`
	StartDate          sql.NullString  `db:"start_date"`
	TimeOfDay          sql.NullString  `db:"time_of_day"`
	EstimatedMinutes   sql.NullInt64   `db:"estimated_minutes"`
	EnergyLevel        sql.NullInt64   `db:"energy_level"`
	Context            sql.NullString  `db:"context"`
	Notes              sql.NullString  `db:"notes"`
	LastFocusSessionID sql.NullString  `db:"last_focus_session_id"`
	FocusTotalMinutes  int             `db:"focus_total_minutes"`
	CreatedAt          string          `db:"created_at"`
	UpdatedAt          string          `db:"updated_at"`
}

func mapRowToTask(row taskRow) *Task {
	task := &Task{
		ID:                row.ID,
		Title:             row.Title,
		Status:            row.Status,
		ShowStatus:        row.ShowStatus,
		Priority:          row.Priority,
		FocusTotalMinutes: row.FocusTotalMinutes,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}

	if row.GoalID.Valid {
		task.GoalID = &row.GoalID.String
	}
	if row.HabitID.Valid {
		task.HabitID = &row.HabitID.String
	}
	if row.FinanceLink.Valid {
		task.FinanceLink = &row.FinanceLink.String
	}
	if row.ProgressValue.Valid {
		task.ProgressValue = &row.ProgressValue.Float64
	}
	if row.ProgressUnit.Valid {
		task.ProgressUnit = &row.ProgressUnit.String
	}
	if row.DueDate.Valid {
		task.DueDate = &row.DueDate.String
	}
	if row.StartDate.Valid {
		task.StartDate = &row.StartDate.String
	}
	if row.TimeOfDay.Valid {
		task.TimeOfDay = &row.TimeOfDay.String
	}
	if row.EstimatedMinutes.Valid {
		minutes := int(row.EstimatedMinutes.Int64)
		task.EstimatedMinutes = &minutes
	}
	if row.EnergyLevel.Valid {
		level := int(row.EnergyLevel.Int64)
		task.EnergyLevel = &level
	}
	if row.Context.Valid {
		task.Context = &row.Context.String
	}
	if row.Notes.Valid {
		task.Notes = &row.Notes.String
	}
	if row.LastFocusSessionID.Valid {
		task.LastFocusSessionID = &row.LastFocusSessionID.String
	}

	return task
}

// getChecklistItems fetches all checklist items for a task
func (r *PostgresRepository) getChecklistItems(ctx context.Context, taskID string) ([]*ChecklistItem, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT id, task_id, title, completed, item_order, created_at, updated_at
		FROM task_checklist_items
		WHERE task_id = $1
		ORDER BY item_order ASC
	`, taskID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var items []*ChecklistItem
	for rows.Next() {
		var item ChecklistItem
		if err := rows.StructScan(&item); err != nil {
			return nil, appErrors.DatabaseError
		}
		items = append(items, &item)
	}

	return items, nil
}

// saveChecklistItems saves checklist items for a task (replaces all existing items)
func (r *PostgresRepository) saveChecklistItems(ctx context.Context, taskID string, items []*ChecklistItem) error {
	// Delete existing items
	_, err := r.db.ExecContext(ctx, `DELETE FROM task_checklist_items WHERE task_id = $1`, taskID)
	if err != nil {
		return appErrors.DatabaseError
	}

	// Insert new items
	for i, item := range items {
		if item.ID == "" {
			item.ID = uuid.NewString()
		}
		item.TaskID = taskID
		item.Order = i

		now := utils.NowUTC()
		item.CreatedAt = now
		item.UpdatedAt = now

		_, err := r.db.ExecContext(ctx, `
			INSERT INTO task_checklist_items (id, task_id, title, completed, item_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.ID, item.TaskID, item.Title, item.Completed, item.Order, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return appErrors.DatabaseError
		}
	}

	return nil
}

// UpdateChecklistItem updates a single checklist item
func (r *PostgresRepository) UpdateChecklistItem(ctx context.Context, taskID, itemID string, completed bool) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	// Verify task belongs to user
	_, err := r.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE task_checklist_items
		SET completed = $1, updated_at = $2
		WHERE id = $3 AND task_id = $4
	`, completed, utils.NowUTC(), itemID, taskID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.TaskNotFound
	}

	return nil
}
