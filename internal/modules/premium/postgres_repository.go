package premium

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

const (
	planSelectFields         = `id, name, interval, prices, created_at, updated_at`
	subscriptionSelectFields = `id, tier, status, cancel_at_period_end, created_at, updated_at`
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ========== PLANS ==========

func (r *PostgresRepository) ListPlans(ctx context.Context) ([]*Plan, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM premium_plans
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`, planSelectFields)

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		var row planRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		plan, err := mapRowToPlan(row)
		if err != nil {
			return nil, appErrors.DatabaseError
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func (r *PostgresRepository) GetPlanByID(ctx context.Context, id string) (*Plan, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM premium_plans
		WHERE id = $1 AND deleted_at IS NULL
	`, planSelectFields)

	var row planRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.PlanNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToPlan(row)
}

func (r *PostgresRepository) CreatePlan(ctx context.Context, plan *Plan) error {
	if plan.ID == "" {
		plan.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	pricesJSON, err := json.Marshal(plan.Prices)
	if err != nil {
		return appErrors.InvalidSubscriptionData
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO premium_plans (id, name, interval, prices, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, plan.ID, plan.Name, plan.Interval, pricesJSON, plan.CreatedAt, plan.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdatePlan(ctx context.Context, plan *Plan) error {
	plan.UpdatedAt = utils.NowUTC()

	pricesJSON, err := json.Marshal(plan.Prices)
	if err != nil {
		return appErrors.InvalidSubscriptionData
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE premium_plans
		SET name = $1, interval = $2, prices = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`, plan.Name, plan.Interval, pricesJSON, plan.UpdatedAt, plan.ID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.PlanNotFound
	}

	return nil
}

func (r *PostgresRepository) DeletePlan(ctx context.Context, id string) error {
	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE premium_plans
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, now, now, id)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.PlanNotFound
	}

	return nil
}

// ========== SUBSCRIPTIONS ==========

func (r *PostgresRepository) ListSubscriptions(ctx context.Context) ([]*Subscription, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, subscriptionSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var row subscriptionRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		subscriptions = append(subscriptions, mapRowToSubscription(row))
	}

	return subscriptions, nil
}

func (r *PostgresRepository) GetSubscriptionByID(ctx context.Context, id string) (*Subscription, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM subscriptions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, subscriptionSelectFields)

	var row subscriptionRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.SubscriptionNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToSubscription(row), nil
}

func (r *PostgresRepository) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if subscription.ID == "" {
		subscription.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO subscriptions (id, user_id, tier, status, cancel_at_period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, subscription.ID, userID, subscription.Tier, subscription.Status, subscription.CancelAtPeriodEnd, subscription.CreatedAt, subscription.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdateSubscription(ctx context.Context, subscription *Subscription) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	subscription.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET tier = $1, status = $2, cancel_at_period_end = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`, subscription.Tier, subscription.Status, subscription.CancelAtPeriodEnd, subscription.UpdatedAt, subscription.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.SubscriptionNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteSubscription(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
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
		return appErrors.SubscriptionNotFound
	}

	return nil
}

// ========== ROW STRUCTS AND MAPPERS ==========

type planRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Interval  string `db:"interval"`
	Prices    string `db:"prices"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func mapRowToPlan(row planRow) (*Plan, error) {
	var prices map[string]float64
	if err := json.Unmarshal([]byte(row.Prices), &prices); err != nil {
		return nil, err
	}

	return &Plan{
		ID:        row.ID,
		Name:      row.Name,
		Interval:  row.Interval,
		Prices:    prices,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

type subscriptionRow struct {
	ID                string `db:"id"`
	Tier              string `db:"tier"`
	Status            string `db:"status"`
	CancelAtPeriodEnd bool   `db:"cancel_at_period_end"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

func mapRowToSubscription(row subscriptionRow) *Subscription {
	return &Subscription{
		ID:                row.ID,
		Tier:              row.Tier,
		Status:            row.Status,
		CancelAtPeriodEnd: row.CancelAtPeriodEnd,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
