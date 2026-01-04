package widgets

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const widgetSelectFields = `
	id,
	title,
	config,
	created_at,
	updated_at
`

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Widget, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM widgets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, widgetSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var widgets []*Widget
	for rows.Next() {
		var row widgetRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		widget, err := mapRowToWidget(row)
		if err != nil {
			return nil, appErrors.DatabaseError
		}
		widgets = append(widgets, widget)
	}

	return widgets, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Widget, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM widgets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, widgetSelectFields)

	return r.fetchWidget(ctx, query, id, userID)
}

func (r *PostgresRepository) Create(ctx context.Context, widget *Widget) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if widget.ID == "" {
		widget.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	widget.CreatedAt = now
	widget.UpdatedAt = now

	configJSON, err := json.Marshal(widget.Config)
	if err != nil {
		return appErrors.InvalidWidgetData
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO widgets (id, user_id, title, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, widget.ID, userID, widget.Title, configJSON, widget.CreatedAt, widget.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, widget *Widget) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	widget.UpdatedAt = utils.NowUTC()

	configJSON, err := json.Marshal(widget.Config)
	if err != nil {
		return appErrors.InvalidWidgetData
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE widgets
		SET title = $1, config = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL
	`, widget.Title, configJSON, widget.UpdatedAt, widget.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.WidgetNotFound
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
		UPDATE widgets
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
		return appErrors.WidgetNotFound
	}

	return nil
}

func (r *PostgresRepository) fetchWidget(ctx context.Context, query string, args ...interface{}) (*Widget, error) {
	var row widgetRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.WidgetNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToWidget(row)
}

type widgetRow struct {
	ID        string `db:"id"`
	Title     string `db:"title"`
	Config    string `db:"config"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func mapRowToWidget(row widgetRow) (*Widget, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
		return nil, err
	}

	return &Widget{
		ID:        row.ID,
		Title:     row.Title,
		Config:    config,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
