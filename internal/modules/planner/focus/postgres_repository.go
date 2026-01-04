package focus

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const sessionSelectFields = `
	id, task_id, goal_id, planned_minutes, actual_minutes, status,
	started_at, ended_at, interruptions_count, notes, created_at, updated_at
`

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Session, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM focus_sessions
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, sessionSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var row sessionRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		sessions = append(sessions, mapRowToSession(row))
	}

	return sessions, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM focus_sessions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, sessionSelectFields)

	return r.fetchSession(ctx, query, id, userID)
}

func (r *PostgresRepository) Create(ctx context.Context, session *Session) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if session.ID == "" {
		session.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	session.CreatedAt = now
	session.UpdatedAt = now

	if session.Status == "" {
		session.Status = "in_progress"
	}
	if session.StartedAt == nil {
		session.StartedAt = &now
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO focus_sessions (id, user_id, task_id, goal_id, planned_minutes, actual_minutes,
			status, started_at, ended_at, interruptions_count, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, session.ID, userID, session.TaskID, session.GoalID, session.PlannedMinutes, session.ActualMinutes,
		session.Status, session.StartedAt, session.EndedAt, session.InterruptionsCount, session.Notes,
		session.CreatedAt, session.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, session *Session) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	session.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE focus_sessions
		SET task_id = $1, goal_id = $2, planned_minutes = $3, actual_minutes = $4,
			status = $5, started_at = $6, ended_at = $7, interruptions_count = $8,
			notes = $9, updated_at = $10
		WHERE id = $11 AND user_id = $12 AND deleted_at IS NULL
	`, session.TaskID, session.GoalID, session.PlannedMinutes, session.ActualMinutes,
		session.Status, session.StartedAt, session.EndedAt, session.InterruptionsCount,
		session.Notes, session.UpdatedAt, session.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.FocusSessionNotFound
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
		UPDATE focus_sessions
		SET deleted_at = $1, updated_at = $2
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
		return appErrors.FocusSessionNotFound
	}

	return nil
}

func (r *PostgresRepository) GetStats(ctx context.Context) (*SessionStats, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	stats := &SessionStats{}

	// Total sessions
	r.db.GetContext(ctx, &stats.TotalSessions, `
		SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)

	// Total minutes
	r.db.GetContext(ctx, &stats.TotalMinutes, `
		SELECT COALESCE(SUM(actual_minutes), 0) FROM focus_sessions WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)

	// Completed sessions
	r.db.GetContext(ctx, &stats.CompletedSessions, `
		SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND status = 'completed' AND deleted_at IS NULL
	`, userID)

	// Average minutes
	if stats.CompletedSessions > 0 {
		var avgMinutes float64
		r.db.GetContext(ctx, &avgMinutes, `
			SELECT COALESCE(AVG(actual_minutes), 0) FROM focus_sessions WHERE user_id = $1 AND status = 'completed' AND deleted_at IS NULL
		`, userID)
		stats.AverageMinutes = avgMinutes
	}

	// Total interruptions
	r.db.GetContext(ctx, &stats.TotalInterruptions, `
		SELECT COALESCE(SUM(interruptions_count), 0) FROM focus_sessions WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)

	return stats, nil
}

func (r *PostgresRepository) fetchSession(ctx context.Context, query string, args ...interface{}) (*Session, error) {
	var row sessionRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.FocusSessionNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToSession(row), nil
}

type sessionRow struct {
	ID                 string         `db:"id"`
	TaskID             sql.NullString `db:"task_id"`
	GoalID             sql.NullString `db:"goal_id"`
	PlannedMinutes     int            `db:"planned_minutes"`
	ActualMinutes      int            `db:"actual_minutes"`
	Status             string         `db:"status"`
	StartedAt          sql.NullString `db:"started_at"`
	EndedAt            sql.NullString `db:"ended_at"`
	InterruptionsCount int            `db:"interruptions_count"`
	Notes              sql.NullString `db:"notes"`
	CreatedAt          string         `db:"created_at"`
	UpdatedAt          string         `db:"updated_at"`
}

func mapRowToSession(row sessionRow) *Session {
	session := &Session{
		ID:                 row.ID,
		PlannedMinutes:     row.PlannedMinutes,
		ActualMinutes:      row.ActualMinutes,
		Status:             row.Status,
		InterruptionsCount: row.InterruptionsCount,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}

	if row.TaskID.Valid {
		session.TaskID = &row.TaskID.String
	}
	if row.GoalID.Valid {
		session.GoalID = &row.GoalID.String
	}
	if row.StartedAt.Valid {
		session.StartedAt = &row.StartedAt.String
	}
	if row.EndedAt.Valid {
		session.EndedAt = &row.EndedAt.String
	}
	if row.Notes.Valid {
		session.Notes = &row.Notes.String
	}

	return session
}
