package finance

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const (
	accountSelectFields     = `id, name, currency, account_type AS type, created_at, updated_at`
	transactionSelectFields = `id, account_id, amount, currency, category, created_at, updated_at`
	budgetSelectFields      = `id, name, currency, limit_amount, created_at, updated_at`
	debtSelectFields        = `id, name, balance, created_at, updated_at`
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ========== ACCOUNTS ==========

func (r *PostgresRepository) ListAccounts(ctx context.Context) ([]*Account, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, accountSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		var row accountRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		accounts = append(accounts, mapRowToAccount(row))
	}

	return accounts, nil
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM accounts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, accountSelectFields)

	var row accountRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.AccountNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToAccount(row), nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *Account) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if account.ID == "" {
		account.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO accounts (id, user_id, name, currency, account_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, account.ID, userID, account.Name, account.Currency, account.Type, account.CreatedAt, account.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdateAccount(ctx context.Context, account *Account) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	account.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE accounts
		SET name = $1, currency = $2, account_type = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`, account.Name, account.Currency, account.Type, account.UpdatedAt, account.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.AccountNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteAccount(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE accounts
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
		return appErrors.AccountNotFound
	}

	return nil
}

// ========== TRANSACTIONS ==========

func (r *PostgresRepository) ListTransactions(ctx context.Context) ([]*Transaction, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, transactionSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var row transactionRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		transactions = append(transactions, mapRowToTransaction(row))
	}

	return transactions, nil
}

func (r *PostgresRepository) GetTransactionByID(ctx context.Context, id string) (*Transaction, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM transactions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, transactionSelectFields)

	var row transactionRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.TransactionNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToTransaction(row), nil
}

func (r *PostgresRepository) CreateTransaction(ctx context.Context, txn *Transaction) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if txn.ID == "" {
		txn.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	txn.CreatedAt = now
	txn.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO transactions (id, user_id, account_id, amount, currency, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, txn.ID, userID, txn.AccountID, txn.Amount, txn.Currency, txn.Category, txn.CreatedAt, txn.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdateTransaction(ctx context.Context, txn *Transaction) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	txn.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE transactions
		SET account_id = $1, amount = $2, currency = $3, category = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, txn.AccountID, txn.Amount, txn.Currency, txn.Category, txn.UpdatedAt, txn.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.TransactionNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteTransaction(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE transactions
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
		return appErrors.TransactionNotFound
	}

	return nil
}

// ========== BUDGETS ==========

func (r *PostgresRepository) ListBudgets(ctx context.Context) ([]*Budget, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM budgets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, budgetSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var budgets []*Budget
	for rows.Next() {
		var row budgetRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		budgets = append(budgets, mapRowToBudget(row))
	}

	return budgets, nil
}

func (r *PostgresRepository) GetBudgetByID(ctx context.Context, id string) (*Budget, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM budgets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, budgetSelectFields)

	var row budgetRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.BudgetNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToBudget(row), nil
}

func (r *PostgresRepository) CreateBudget(ctx context.Context, budget *Budget) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if budget.ID == "" {
		budget.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	budget.CreatedAt = now
	budget.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO budgets (id, user_id, name, currency, limit_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, budget.ID, userID, budget.Name, budget.Currency, budget.Limit, budget.CreatedAt, budget.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdateBudget(ctx context.Context, budget *Budget) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	budget.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE budgets
		SET name = $1, currency = $2, limit_amount = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`, budget.Name, budget.Currency, budget.Limit, budget.UpdatedAt, budget.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.BudgetNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteBudget(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE budgets
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
		return appErrors.BudgetNotFound
	}

	return nil
}

// ========== DEBTS ==========

func (r *PostgresRepository) ListDebts(ctx context.Context) ([]*Debt, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM debts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, debtSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var debts []*Debt
	for rows.Next() {
		var row debtRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		debts = append(debts, mapRowToDebt(row))
	}

	return debts, nil
}

func (r *PostgresRepository) GetDebtByID(ctx context.Context, id string) (*Debt, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM debts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, debtSelectFields)

	var row debtRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.DebtNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToDebt(row), nil
}

func (r *PostgresRepository) CreateDebt(ctx context.Context, debt *Debt) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if debt.ID == "" {
		debt.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	debt.CreatedAt = now
	debt.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO debts (id, user_id, name, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, debt.ID, userID, debt.Name, debt.Balance, debt.CreatedAt, debt.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) UpdateDebt(ctx context.Context, debt *Debt) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	debt.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE debts
		SET name = $1, balance = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL
	`, debt.Name, debt.Balance, debt.UpdatedAt, debt.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.DebtNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteDebt(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE debts
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
		return appErrors.DebtNotFound
	}

	return nil
}

// ========== ROW STRUCTS AND MAPPERS ==========

type accountRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Currency  string `db:"currency"`
	Type      string `db:"type"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func mapRowToAccount(row accountRow) *Account {
	return &Account{
		ID:        row.ID,
		Name:      row.Name,
		Currency:  row.Currency,
		Type:      row.Type,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

type transactionRow struct {
	ID        string  `db:"id"`
	AccountID string  `db:"account_id"`
	Amount    float64 `db:"amount"`
	Currency  string  `db:"currency"`
	Category  string  `db:"category"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

func mapRowToTransaction(row transactionRow) *Transaction {
	return &Transaction{
		ID:        row.ID,
		AccountID: row.AccountID,
		Amount:    row.Amount,
		Currency:  row.Currency,
		Category:  row.Category,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

type budgetRow struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	Currency  string  `db:"currency"`
	Limit     float64 `db:"limit_amount"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

func mapRowToBudget(row budgetRow) *Budget {
	return &Budget{
		ID:        row.ID,
		Name:      row.Name,
		Currency:  row.Currency,
		Limit:     row.Limit,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

type debtRow struct {
	ID        string  `db:"id"`
	Name      string  `db:"name"`
	Balance   float64 `db:"balance"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

func mapRowToDebt(row debtRow) *Debt {
	return &Debt{
		ID:        row.ID,
		Name:      row.Name,
		Balance:   row.Balance,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
