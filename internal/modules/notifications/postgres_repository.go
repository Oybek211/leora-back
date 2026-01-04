package notifications

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const notificationSelectFields = `
	id,
	title,
	message,
	created_at,
	updated_at
`

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Notification, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM notifications
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, notificationSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var row notificationRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		notifications = append(notifications, mapRowToNotification(row))
	}

	return notifications, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Notification, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM notifications
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, notificationSelectFields)

	return r.fetchNotification(ctx, query, id, userID)
}

func (r *PostgresRepository) Create(ctx context.Context, notification *Notification) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if notification.ID == "" {
		notification.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	notification.CreatedAt = now
	notification.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, title, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, notification.ID, userID, notification.Title, notification.Message, notification.CreatedAt, notification.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, notification *Notification) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	notification.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET title = $1, message = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL
	`, notification.Title, notification.Message, notification.UpdatedAt, notification.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.NotificationNotFound
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
		UPDATE notifications
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
		return appErrors.NotificationNotFound
	}

	return nil
}

func (r *PostgresRepository) fetchNotification(ctx context.Context, query string, args ...interface{}) (*Notification, error) {
	var row notificationRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.NotificationNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToNotification(row), nil
}

type notificationRow struct {
	ID        string `db:"id"`
	Title     string `db:"title"`
	Message   string `db:"message"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func mapRowToNotification(row notificationRow) *Notification {
	return &Notification{
		ID:        row.ID,
		Title:     row.Title,
		Message:   row.Message,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
