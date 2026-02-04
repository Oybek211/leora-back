package finance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

const (
	accountSelectFields      = `id, user_id, name, currency, account_type AS account_type, initial_balance, current_balance, linked_goal_id, custom_type_id, is_main, is_archived, show_status, created_at, updated_at`
	transactionSelectFields  = `id, user_id, type, status, account_id, from_account_id, to_account_id, reference_type, reference_id, amount, currency, base_currency, rate_used_to_base, converted_amount_to_base, to_amount, to_currency, effective_rate_from_to, fee_amount, fee_category_id, category_id, category, subcategory_id, name, description, date, time, linked_goal_id, budget_id, linked_debt_id, habit_id, counterparty_id, recurring_id, attachments, tags, is_balance_adjustment, skip_budget_matching, show_status, related_budget_id, related_debt_id, planned_amount, paid_amount, original_currency, original_amount, conversion_rate, occurred_at, metadata, created_at, updated_at`
	budgetSelectFields       = `id, user_id, name, budget_type, category_ids, linked_goal_id, account_id, transaction_type, currency, limit_amount, period_type, start_date, end_date, spent_amount, remaining_amount, percent_used, is_overspent, rollover_mode, notify_on_exceed, contribution_total, current_balance, is_archived, show_status, created_at, updated_at`
	debtSelectFields         = `id, user_id, name, balance, direction, counterparty_id, counterparty_name, description, principal_amount, principal_currency, principal_original_amount, principal_original_currency, base_currency, rate_on_start, principal_base_value, repayment_currency, repayment_amount, repayment_rate_on_start, is_fixed_repayment_amount, start_date, due_date, interest_mode, interest_rate_annual, schedule_hint, linked_goal_id, linked_budget_id, funding_account_id, funding_transaction_id, lent_from_account_id, return_to_account_id, received_to_account_id, pay_from_account_id, custom_rate_used, exchange_rate_current, reminder_enabled, reminder_time, status, settled_at, final_rate_used, final_profit_loss, final_profit_loss_currency, total_paid_in_repayment_currency, remaining_amount, total_paid, percent_paid, show_status, created_at, updated_at`
	debtPaymentSelectFields  = `dp.id, dp.debt_id, dp.amount, dp.currency, dp.base_currency, dp.rate_used_to_base, dp.converted_amount_to_base, dp.rate_used_to_debt, dp.converted_amount_to_debt, dp.payment_date, dp.account_id, dp.note, dp.related_transaction_id, dp.applied_rate, dp.created_at AS created_at, dp.updated_at AS updated_at, dp.deleted_at`
	counterpartySelectFields = `id, user_id, display_name, phone_number, comment, search_keywords, show_status, created_at, updated_at, deleted_at`
	fxRateSelectFields       = `id, rate_date, from_currency, to_currency, rate, rate_mid, rate_bid, rate_ask, nominal, spread_percent, source, created_at, updated_at`
	categorySelectFields     = `id, type, name_i18n, icon_name, color, is_default, sort_order, is_active, created_at, updated_at`
)

type PostgresRepository struct {
	db *sqlx.DB
}

type categoryRow struct {
	ID        string         `db:"id"`
	Type      string         `db:"type"`
	NameI18n  []byte         `db:"name_i18n"`
	IconName  string         `db:"icon_name"`
	Color     sql.NullString `db:"color"`
	IsDefault bool           `db:"is_default"`
	SortOrder int            `db:"sort_order"`
	IsActive  bool           `db:"is_active"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
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
		log.Printf("[ListAccounts] DB query error for user=%s: %v", userID, err)
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		var row accountRow
		if err := rows.StructScan(&row); err != nil {
			log.Printf("[ListAccounts] Row scan error: %v", err)
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
			log.Printf("[GetAccountByID] Account not found: id=%s, user=%s", id, userID)
			return nil, appErrors.AccountNotFound
		}
		log.Printf("[GetAccountByID] DB error for id=%s: %v", id, err)
		return nil, appErrors.DatabaseError
	}
	return mapRowToAccount(row), nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *Account) (*Transaction, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	if account.ID == "" {
		account.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	account.UserID = userID
	account.CreatedAt = now
	account.UpdatedAt = now
	if account.CurrentBalance == 0 {
		account.CurrentBalance = account.InitialBalance
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[CreateAccount] Failed to begin transaction: %v", err)
		return nil, appErrors.DatabaseError
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (id, user_id, name, currency, account_type, initial_balance, current_balance, linked_goal_id, custom_type_id, is_main, is_archived, show_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, account.ID, userID, account.Name, account.Currency, account.AccountType, account.InitialBalance, account.CurrentBalance, account.LinkedGoalID, account.CustomTypeID, account.IsMain, account.IsArchived, account.ShowStatus, account.CreatedAt, account.UpdatedAt); err != nil {
		log.Printf("[CreateAccount] INSERT error for account=%s: %v", account.Name, err)
		_ = tx.Rollback()
		return nil, appErrors.DatabaseError
	}

	var openingTxn *Transaction
	if account.InitialBalance != 0 {
		openingTxn = buildOpeningTransaction(account, userID)
		normalizeTransaction(openingTxn)
		if err := r.insertTransaction(ctx, tx, userID, openingTxn); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, appErrors.DatabaseError
	}

	return openingTxn, nil
}

func (r *PostgresRepository) UpdateAccount(ctx context.Context, account *Account) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	account.UpdatedAt = utils.NowUTC()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[UpdateAccount] Failed to begin transaction for id=%s: %v", account.ID, err)
		return appErrors.DatabaseError
	}

	current, err := fetchAccountForUpdate(ctx, tx, userID, account.ID)
	if err != nil {
		log.Printf("[UpdateAccount] Failed to fetch account for update id=%s: %v", account.ID, err)
		_ = tx.Rollback()
		return err
	}
	if account.Currency != "" && current.Currency != account.Currency && current.CurrentBalance != 0 {
		log.Printf("[UpdateAccount] Cannot change currency for account with balance: id=%s, currentBalance=%.2f", account.ID, current.CurrentBalance)
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	desiredBalance := account.CurrentBalance
	newBalance := desiredBalance
	delta := desiredBalance - current.CurrentBalance

	result, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET name = $1,
			currency = $2,
			account_type = $3,
			initial_balance = $4,
			current_balance = $5,
			linked_goal_id = $6,
			custom_type_id = $7,
			is_main = $8,
			is_archived = $9,
			show_status = $10,
			updated_at = $11
		WHERE id = $12 AND user_id = $13 AND deleted_at IS NULL
	`, account.Name, account.Currency, account.AccountType, account.InitialBalance, newBalance,
		account.LinkedGoalID, account.CustomTypeID, account.IsMain, account.IsArchived, account.ShowStatus,
		account.UpdatedAt, account.ID, userID)

	if err != nil {
		log.Printf("[UpdateAccount] UPDATE error for id=%s: %v", account.ID, err)
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("[UpdateAccount] RowsAffected error for id=%s: %v", account.ID, err)
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if rows == 0 {
		log.Printf("[UpdateAccount] No rows affected for id=%s, user=%s", account.ID, userID)
		_ = tx.Rollback()
		return appErrors.AccountNotFound
	}

	if delta != 0 {
		referenceType := "account"
		referenceID := account.ID
		adjustment := &Transaction{
			ID:                  uuid.NewString(),
			UserID:              userID,
			Type:                TransactionTypeSystemAdjustment,
			AccountID:           &account.ID,
			ReferenceType:       &referenceType,
			ReferenceID:         &referenceID,
			Amount:              delta,
			Currency:            account.Currency,
			BaseCurrency:        account.Currency,
			Date:                time.Now().UTC().Format("2006-01-02"),
			IsBalanceAdjustment: true,
			ShowStatus:          "active",
			Attachments:         []string{},
			Tags:                []string{},
		}
		normalizeTransaction(adjustment)
		if err := r.insertTransaction(ctx, tx, userID, adjustment); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) DeleteAccount(ctx context.Context, id string) (*Transaction, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}
	log.Printf("[DeleteAccount] Starting delete for id=%s, user=%s", id, userID)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[DeleteAccount] Failed to begin transaction for id=%s: %v", id, err)
		return nil, appErrors.DatabaseError
	}

	account, err := fetchAccountForUpdate(ctx, tx, userID, id)
	if err != nil {
		log.Printf("[DeleteAccount] Failed to fetch account for id=%s: %v", id, err)
		_ = tx.Rollback()
		return nil, err
	}

	var withdrawal *Transaction
	if account.CurrentBalance != 0 {
		log.Printf("[DeleteAccount] Creating withdrawal transaction for balance=%.2f", account.CurrentBalance)
		referenceType := "account"
		referenceID := account.ID
		withdrawal = &Transaction{
			ID:                  uuid.NewString(),
			UserID:              userID,
			Type:                TransactionTypeAccountDeleteWithdrawal,
			AccountID:           &account.ID,
			ReferenceType:       &referenceType,
			ReferenceID:         &referenceID,
			Amount:              account.CurrentBalance,
			Currency:            account.Currency,
			BaseCurrency:        account.Currency,
			Date:                time.Now().UTC().Format("2006-01-02"),
			IsBalanceAdjustment: true,
			ShowStatus:          "active",
			Status:              TransactionStatusCompleted,
			Attachments:         []string{},
			Tags:                []string{},
		}
		normalizeTransaction(withdrawal)
		if err := r.insertTransaction(ctx, tx, userID, withdrawal); err != nil {
			log.Printf("[DeleteAccount] Failed to insert withdrawal transaction for id=%s: %v", id, err)
			_ = tx.Rollback()
			return nil, err
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, 0); err != nil {
			log.Printf("[DeleteAccount] Failed to update balance to 0 for id=%s: %v", id, err)
			_ = tx.Rollback()
			return nil, err
		}
	}

	now := utils.NowUTC()
	result, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
	`, now, now, id, userID)

	if err != nil {
		log.Printf("[DeleteAccount] UPDATE (soft-delete) error for id=%s: %v", id, err)
		_ = tx.Rollback()
		return nil, appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("[DeleteAccount] RowsAffected error for id=%s: %v", id, err)
		_ = tx.Rollback()
		return nil, appErrors.DatabaseError
	}

	if rows == 0 {
		log.Printf("[DeleteAccount] No rows affected (not found) for id=%s, user=%s", id, userID)
		_ = tx.Rollback()
		return nil, appErrors.AccountNotFound
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[DeleteAccount] Commit error for id=%s: %v", id, err)
		return nil, appErrors.DatabaseError
	}

	log.Printf("[DeleteAccount] Successfully deleted account id=%s", id)
	return withdrawal, nil
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
		log.Printf("[ListTransactions] DB query error for user=%s: %v", userID, err)
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var row transactionRow
		if err := rows.StructScan(&row); err != nil {
			log.Printf("[ListTransactions] Row scan error: %v", err)
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
			log.Printf("[GetTransactionByID] Transaction not found: id=%s", id)
			return nil, appErrors.TransactionNotFound
		}
		log.Printf("[GetTransactionByID] DB error for id=%s: %v", id, err)
		return nil, appErrors.DatabaseError
	}
	return mapRowToTransaction(row), nil
}

func (r *PostgresRepository) CreateTransaction(ctx context.Context, txn *Transaction) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if txn.Amount == 0 {
		log.Printf("[CreateTransaction] Invalid amount=0 for type=%s", txn.Type)
		return appErrors.InvalidFinanceData
	}
	txn.Type = strings.ToLower(txn.Type)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[CreateTransaction] Failed to begin transaction: %v", err)
		return appErrors.DatabaseError
	}

	switch strings.ToLower(txn.Type) {
	case TransactionTypeIncome, TransactionTypeExpense:
		if txn.AccountID == nil || *txn.AccountID == "" {
			log.Printf("[CreateTransaction] Missing accountId for type=%s", txn.Type)
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		account, err := fetchAccountForUpdate(ctx, tx, userID, *txn.AccountID)
		if err != nil {
			log.Printf("[CreateTransaction] Failed to fetch account=%s: %v", *txn.AccountID, err)
			_ = tx.Rollback()
			return err
		}
		if txn.Currency == "" {
			txn.Currency = account.Currency
		}
		if txn.Currency != account.Currency {
			log.Printf("[CreateTransaction] Currency mismatch: txn=%s, account=%s", txn.Currency, account.Currency)
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		if txn.Type == TransactionTypeExpense && account.CurrentBalance < txn.Amount {
			log.Printf("[CreateTransaction] Insufficient funds: required=%.2f, available=%.2f", txn.Amount, account.CurrentBalance)
			_ = tx.Rollback()
			return appErrors.InsufficientFunds
		}
		normalizeTransaction(txn)
		if err := r.insertTransaction(ctx, tx, userID, txn); err != nil {
			_ = tx.Rollback()
			return err
		}
		newBalance := account.CurrentBalance
		if txn.Type == TransactionTypeIncome {
			newBalance += txn.Amount
		} else {
			newBalance -= txn.Amount
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, newBalance); err != nil {
			_ = tx.Rollback()
			return err
		}
	case TransactionTypeDebtAdjustment, TransactionTypeBudgetAddValue, TransactionTypeDebtAddValue, TransactionTypeDebtFullPayment, TransactionTypeAccountDeleteWithdrawal:
		if txn.AccountID == nil || *txn.AccountID == "" {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		account, err := fetchAccountForUpdate(ctx, tx, userID, *txn.AccountID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if txn.Currency == "" {
			txn.Currency = account.Currency
		}
		if txn.Currency != account.Currency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		delta := txn.Amount
		if delta < 0 && account.CurrentBalance < -delta {
			_ = tx.Rollback()
			return appErrors.InsufficientFunds
		}
		normalizeTransaction(txn)
		if err := r.insertTransaction(ctx, tx, userID, txn); err != nil {
			_ = tx.Rollback()
			return err
		}
		newBalance := account.CurrentBalance + delta
		if err := updateAccountBalance(ctx, tx, userID, account.ID, newBalance); err != nil {
			_ = tx.Rollback()
			return err
		}
	case TransactionTypeTransfer:
		if txn.FromAccountID == nil || txn.ToAccountID == nil || *txn.FromAccountID == "" || *txn.ToAccountID == "" {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		fromAccount, toAccount, err := fetchTransferAccountsForUpdate(ctx, tx, userID, *txn.FromAccountID, *txn.ToAccountID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if fromAccount.Currency != toAccount.Currency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		if txn.Currency == "" {
			txn.Currency = fromAccount.Currency
		}
		if txn.Currency != fromAccount.Currency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		if fromAccount.CurrentBalance < txn.Amount {
			_ = tx.Rollback()
			return appErrors.InsufficientFunds
		}
		if txn.ToAmount == 0 {
			txn.ToAmount = txn.Amount
		}
		normalizeTransaction(txn)
		referenceType := "transfer"
		referenceID := uuid.NewString()

		transferOut := *txn
		transferOut.ID = ""
		transferOut.Type = TransactionTypeTransferOut
		transferOut.AccountID = txn.FromAccountID
		transferOut.ReferenceType = &referenceType
		transferOut.ReferenceID = &referenceID
		normalizeTransaction(&transferOut)
		if err := r.insertTransaction(ctx, tx, userID, &transferOut); err != nil {
			_ = tx.Rollback()
			return err
		}

		transferIn := *txn
		transferIn.ID = ""
		transferIn.Type = TransactionTypeTransferIn
		transferIn.AccountID = txn.ToAccountID
		transferIn.Amount = txn.ToAmount
		transferIn.Currency = toAccount.Currency
		transferIn.ReferenceType = &referenceType
		transferIn.ReferenceID = &referenceID
		normalizeTransaction(&transferIn)
		if err := r.insertTransaction(ctx, tx, userID, &transferIn); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := updateAccountBalance(ctx, tx, userID, fromAccount.ID, fromAccount.CurrentBalance-txn.Amount); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := updateAccountBalance(ctx, tx, userID, toAccount.ID, toAccount.CurrentBalance+txn.ToAmount); err != nil {
			_ = tx.Rollback()
			return err
		}

		txn.ID = transferOut.ID
		txn.Type = transferOut.Type
		txn.ReferenceType = transferOut.ReferenceType
		txn.ReferenceID = transferOut.ReferenceID
	default:
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	if err := tx.Commit(); err != nil {
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

	attachments, err := json.Marshal(txn.Attachments)
	if err != nil {
		return appErrors.InvalidFinanceData
	}
	tags, err := json.Marshal(txn.Tags)
	if err != nil {
		return appErrors.InvalidFinanceData
	}
	metadata, err := json.Marshal(txn.Metadata)
	if err != nil {
		return appErrors.InvalidFinanceData
	}

	categoryValue := resolveTransactionCategory(txn)

	result, err := r.db.ExecContext(ctx, `
		UPDATE transactions
		SET type = $1,
			status = $2,
			account_id = $3,
			from_account_id = $4,
			to_account_id = $5,
			amount = $6,
			currency = $7,
			base_currency = $8,
			rate_used_to_base = $9,
			converted_amount_to_base = $10,
			to_amount = $11,
			to_currency = $12,
			effective_rate_from_to = $13,
			fee_amount = $14,
			fee_category_id = $15,
			category_id = $16,
			category = $17,
			subcategory_id = $18,
			name = $19,
			description = $20,
			date = $21,
			time = $22,
			linked_goal_id = $23,
			budget_id = $24,
			linked_debt_id = $25,
			habit_id = $26,
			counterparty_id = $27,
			recurring_id = $28,
			attachments = $29,
			tags = $30,
			is_balance_adjustment = $31,
			skip_budget_matching = $32,
			show_status = $33,
			related_budget_id = $34,
			related_debt_id = $35,
			planned_amount = $36,
			paid_amount = $37,
			original_currency = $38,
			original_amount = $39,
			conversion_rate = $40,
			occurred_at = $41,
			metadata = $42,
			updated_at = $43
		WHERE id = $44 AND user_id = $45 AND deleted_at IS NULL
	`, txn.Type, txn.Status, txn.AccountID, txn.FromAccountID, txn.ToAccountID,
		txn.Amount, txn.Currency, txn.BaseCurrency, txn.RateUsedToBase, txn.ConvertedAmountToBase,
		txn.ToAmount, txn.ToCurrency, txn.EffectiveRateFromTo, txn.FeeAmount, txn.FeeCategoryID,
		txn.CategoryID, categoryValue, txn.SubcategoryID, txn.Name, txn.Description, txn.Date, txn.Time,
		txn.GoalID, txn.BudgetID, txn.DebtID, txn.HabitID, txn.CounterpartyID, txn.RecurringID,
		attachments, tags, txn.IsBalanceAdjustment, txn.SkipBudgetMatching, txn.ShowStatus,
		txn.RelatedBudgetID, txn.RelatedDebtID, txn.PlannedAmount, txn.PaidAmount,
		txn.OriginalCurrency, txn.OriginalAmount, txn.ConversionRate, txn.OccurredAt, metadata, txn.UpdatedAt, txn.ID, userID)

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

func resolveTransactionCategory(txn *Transaction) string {
	if txn == nil {
		return "Other"
	}
	if txn.CategoryID != nil {
		trimmed := strings.TrimSpace(*txn.CategoryID)
		if trimmed != "" {
			return trimmed
		}
	}
	return "Other"
}

type accountBalanceRow struct {
	ID             string  `db:"id"`
	Currency       string  `db:"currency"`
	InitialBalance float64 `db:"initial_balance"`
	CurrentBalance float64 `db:"current_balance"`
}

type debtBalanceRow struct {
	ID                  string         `db:"id"`
	Direction           string         `db:"direction"`
	PrincipalAmount     float64        `db:"principal_amount"`
	PrincipalCurrency   string         `db:"principal_currency"`
	RepaymentCurrency   sql.NullString `db:"repayment_currency"`
	RemainingAmount     float64        `db:"remaining_amount"`
	TotalPaid           float64        `db:"total_paid"`
	TotalPaidInRepaymentCurrency float64 `db:"total_paid_in_repayment_currency"`
	FundingAccountID    sql.NullString `db:"funding_account_id"`
	LentFromAccountID   sql.NullString `db:"lent_from_account_id"`
	ReceivedToAccountID sql.NullString `db:"received_to_account_id"`
}

type debtPaymentBalanceRow struct {
	ID                    string         `db:"id"`
	Amount                float64        `db:"amount"`
	Currency              string         `db:"currency"`
	ConvertedAmountToDebt float64        `db:"converted_amount_to_debt"`
	AccountID             sql.NullString `db:"account_id"`
}

func buildOpeningTransaction(account *Account, userID string) *Transaction {
	now := utils.NowUTC()
	openingDate := time.Now().UTC().Format("2006-01-02")
	referenceType := "account"
	referenceID := account.ID
	return &Transaction{
		ID:                  uuid.NewString(),
		UserID:              userID,
		Type:                TransactionTypeAccountCreateFunding,
		AccountID:           &account.ID,
		ReferenceType:       &referenceType,
		ReferenceID:         &referenceID,
		Amount:              account.InitialBalance,
		Currency:            account.Currency,
		BaseCurrency:        account.Currency,
		Date:                openingDate,
		IsBalanceAdjustment: true,
		ShowStatus:          "active",
		Status:              TransactionStatusCompleted,
		CreatedAt:           now,
		UpdatedAt:           now,
		Attachments:         []string{},
		Tags:                []string{},
	}
}

func (r *PostgresRepository) insertTransaction(ctx context.Context, execer sqlx.ExtContext, userID string, txn *Transaction) error {
	if txn.ID == "" {
		txn.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	txn.UserID = userID
	if txn.CreatedAt == "" {
		txn.CreatedAt = now
	}
	txn.UpdatedAt = now
	if strings.TrimSpace(txn.Date) == "" {
		txn.Date = time.Now().UTC().Format("2006-01-02")
	}

	attachments, err := json.Marshal(txn.Attachments)
	if err != nil {
		return appErrors.InvalidFinanceData
	}
	tags, err := json.Marshal(txn.Tags)
	if err != nil {
		return appErrors.InvalidFinanceData
	}
	metadata, err := json.Marshal(txn.Metadata)
	if err != nil {
		return appErrors.InvalidFinanceData
	}
	categoryValue := resolveTransactionCategory(txn)

	if _, err := execer.ExecContext(ctx, `
		INSERT INTO transactions (
			id, user_id, type, status, account_id, from_account_id, to_account_id,
			reference_type, reference_id,
			amount, currency, base_currency, rate_used_to_base, converted_amount_to_base,
			to_amount, to_currency, effective_rate_from_to, fee_amount, fee_category_id,
			category_id, category, subcategory_id, name, description, date, time,
			linked_goal_id, budget_id, linked_debt_id, habit_id,
			counterparty_id, recurring_id, attachments, tags,
			is_balance_adjustment, skip_budget_matching, show_status,
			related_budget_id, related_debt_id, planned_amount, paid_amount,
			original_currency, original_amount, conversion_rate, occurred_at, metadata, created_at, updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,
			$10,$11,$12,$13,$14,
			$15,$16,$17,$18,$19,
			$20,$21,$22,$23,$24,$25,$26,
			$27,$28,$29,$30,
			$31,$32,$33,$34,
			$35,$36,$37,
			$38,$39,$40,$41,$42,$43,$44,$45,$46,$47
			,$48
		)
	`, txn.ID, userID, txn.Type, txn.Status, txn.AccountID, txn.FromAccountID, txn.ToAccountID,
		txn.ReferenceType, txn.ReferenceID,
		txn.Amount, txn.Currency, txn.BaseCurrency, txn.RateUsedToBase, txn.ConvertedAmountToBase,
		txn.ToAmount, txn.ToCurrency, txn.EffectiveRateFromTo, txn.FeeAmount, txn.FeeCategoryID,
		txn.CategoryID, categoryValue, txn.SubcategoryID, txn.Name, txn.Description, txn.Date, txn.Time,
		txn.GoalID, txn.BudgetID, txn.DebtID, txn.HabitID,
		txn.CounterpartyID, txn.RecurringID, attachments, tags,
		txn.IsBalanceAdjustment, txn.SkipBudgetMatching, txn.ShowStatus,
		txn.RelatedBudgetID, txn.RelatedDebtID, txn.PlannedAmount, txn.PaidAmount,
		txn.OriginalCurrency, txn.OriginalAmount, txn.ConversionRate, txn.OccurredAt, metadata, txn.CreatedAt, txn.UpdatedAt,
	); err != nil {
		log.Printf("[insertTransaction] INSERT error for type=%s, amount=%.2f: %v", txn.Type, txn.Amount, err)
		return appErrors.DatabaseError
	}
	return nil
}

func updateAccountBalance(ctx context.Context, tx *sqlx.Tx, userID, accountID string, balance float64) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET current_balance = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
	`, balance, utils.NowUTC(), accountID, userID); err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func fetchAccountForUpdate(ctx context.Context, tx *sqlx.Tx, userID, accountID string) (*accountBalanceRow, error) {
	var row accountBalanceRow
	if err := tx.GetContext(ctx, &row, `
		SELECT id, currency, initial_balance, current_balance
		FROM accounts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`, accountID, userID); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[fetchAccountForUpdate] Account not found: accountID=%s, userID=%s", accountID, userID)
			hasAccounts, err := userHasAccounts(ctx, tx, userID)
			if err != nil {
				log.Printf("[fetchAccountForUpdate] userHasAccounts error: %v", err)
				return nil, err
			}
			if !hasAccounts {
				log.Printf("[fetchAccountForUpdate] User has no accounts: userID=%s", userID)
				return nil, appErrors.AccountRequired
			}
			return nil, appErrors.AccountNotFound
		}
		log.Printf("[fetchAccountForUpdate] DB error for accountID=%s, userID=%s: %v", accountID, userID, err)
		return nil, appErrors.DatabaseError
	}
	return &row, nil
}

func fetchTransferAccountsForUpdate(ctx context.Context, tx *sqlx.Tx, userID, fromID, toID string) (*accountBalanceRow, *accountBalanceRow, error) {
	rows, err := tx.QueryxContext(ctx, `
		SELECT id, currency, current_balance
		FROM accounts
		WHERE id IN ($1, $2) AND user_id = $3 AND deleted_at IS NULL
		FOR UPDATE
	`, fromID, toID, userID)
	if err != nil {
		return nil, nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var fromAccount *accountBalanceRow
	var toAccount *accountBalanceRow
	for rows.Next() {
		var row accountBalanceRow
		if err := rows.StructScan(&row); err != nil {
			return nil, nil, appErrors.DatabaseError
		}
		if row.ID == fromID {
			fromAccount = &row
		} else if row.ID == toID {
			toAccount = &row
		}
	}
	if fromAccount == nil || toAccount == nil {
		hasAccounts, err := userHasAccounts(ctx, tx, userID)
		if err != nil {
			return nil, nil, err
		}
		if !hasAccounts {
			return nil, nil, appErrors.AccountRequired
		}
		return nil, nil, appErrors.AccountNotFound
	}
	return fromAccount, toAccount, nil
}

func fetchDebtForUpdate(ctx context.Context, tx *sqlx.Tx, userID, debtID string) (*debtBalanceRow, error) {
	var row debtBalanceRow
	if err := tx.GetContext(ctx, &row, `
		SELECT id, direction, principal_amount, principal_currency, repayment_currency,
			remaining_amount, total_paid, total_paid_in_repayment_currency,
			funding_account_id, lent_from_account_id, received_to_account_id
		FROM debts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`, debtID, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.DebtNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return &row, nil
}

func fetchDebtPaymentForUpdate(ctx context.Context, tx *sqlx.Tx, userID, debtID, paymentID string) (*debtPaymentBalanceRow, error) {
	var row debtPaymentBalanceRow
	if err := tx.GetContext(ctx, &row, `
		SELECT dp.id, dp.amount, dp.currency, dp.converted_amount_to_debt, dp.account_id
		FROM debt_payments dp
		JOIN debts d ON dp.debt_id = d.id
		WHERE dp.id = $1 AND dp.debt_id = $2 AND d.user_id = $3 AND dp.deleted_at IS NULL
		FOR UPDATE
	`, paymentID, debtID, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.DebtPaymentNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return &row, nil
}

func resolveDebtFundingAccountID(debt *Debt, existing *debtBalanceRow) *string {
	if debt == nil && existing == nil {
		return nil
	}
	if debt != nil {
		switch debt.Direction {
		case "they_owe_me":
			if debt.LentFromAccountID != nil && *debt.LentFromAccountID != "" {
				return debt.LentFromAccountID
			}
		case "i_owe":
			if debt.ReceivedToAccountID != nil && *debt.ReceivedToAccountID != "" {
				return debt.ReceivedToAccountID
			}
		}
		if debt.FundingAccountID != nil && *debt.FundingAccountID != "" {
			return debt.FundingAccountID
		}
	}
	if existing != nil {
		switch existing.Direction {
		case "they_owe_me":
			if existing.LentFromAccountID.Valid && existing.LentFromAccountID.String != "" {
				return &existing.LentFromAccountID.String
			}
		case "i_owe":
			if existing.ReceivedToAccountID.Valid && existing.ReceivedToAccountID.String != "" {
				return &existing.ReceivedToAccountID.String
			}
		}
		if existing.FundingAccountID.Valid && existing.FundingAccountID.String != "" {
			return &existing.FundingAccountID.String
		}
	}
	return nil
}

func userHasAccounts(ctx context.Context, tx *sqlx.Tx, userID string) (bool, error) {
	var exists bool
	if err := tx.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1 FROM accounts WHERE user_id = $1 AND deleted_at IS NULL
		)
	`, userID); err != nil {
		log.Printf("[userHasAccounts] DB error for userID=%s: %v", userID, err)
		return false, appErrors.DatabaseError
	}
	return exists, nil
}

func (r *PostgresRepository) accountHasTransactions(ctx context.Context, userID, accountID string) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1 FROM transactions
			WHERE user_id = $1
				AND deleted_at IS NULL
				AND (account_id = $2 OR from_account_id = $2 OR to_account_id = $2)
				AND NOT (type = 'system_opening' AND amount = 0)
		)
	`, userID, accountID); err != nil {
		return false, appErrors.DatabaseError
	}
	return exists, nil
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
			log.Printf("[GetBudgetByID] Budget not found: id=%s", id)
			return nil, appErrors.BudgetNotFound
		}
		log.Printf("[GetBudgetByID] DB error for id=%s: %v", id, err)
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

	categoryIDs, err := json.Marshal(budget.CategoryIDs)
	if err != nil {
		return appErrors.InvalidFinanceData
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO budgets (
			id, user_id, name, budget_type, category_ids, linked_goal_id, account_id, transaction_type,
			currency, limit_amount, period_type, start_date, end_date, spent_amount, remaining_amount,
			percent_used, is_overspent, rollover_mode, notify_on_exceed, contribution_total, current_balance,
			is_archived, show_status, created_at, updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,$15,
			$16,$17,$18,$19,$20,$21,
			$22,$23,$24,$25
		)
	`, budget.ID, userID, budget.Name, budget.BudgetType, categoryIDs, budget.LinkedGoalID, budget.AccountID, budget.TransactionType,
		budget.Currency, budget.LimitAmount, budget.PeriodType, budget.StartDate, budget.EndDate, budget.SpentAmount, budget.RemainingAmount,
		budget.PercentUsed, budget.IsOverspent, budget.RolloverMode, budget.NotifyOnExceed, budget.ContributionTotal, budget.CurrentBalance,
		budget.IsArchived, budget.ShowStatus, budget.CreatedAt, budget.UpdatedAt)

	if err != nil {
		log.Printf("[CreateBudget] INSERT error for name=%s: %v", budget.Name, err)
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

	categoryIDs, err := json.Marshal(budget.CategoryIDs)
	if err != nil {
		return appErrors.InvalidFinanceData
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE budgets
		SET name = $1,
			budget_type = $2,
			category_ids = $3,
			linked_goal_id = $4,
			account_id = $5,
			transaction_type = $6,
			currency = $7,
			limit_amount = $8,
			period_type = $9,
			start_date = $10,
			end_date = $11,
			spent_amount = $12,
			remaining_amount = $13,
			percent_used = $14,
			is_overspent = $15,
			rollover_mode = $16,
			notify_on_exceed = $17,
			contribution_total = $18,
			current_balance = $19,
			is_archived = $20,
			show_status = $21,
			updated_at = $22
		WHERE id = $23 AND user_id = $24 AND deleted_at IS NULL
	`, budget.Name, budget.BudgetType, categoryIDs, budget.LinkedGoalID, budget.AccountID, budget.TransactionType, budget.Currency,
		budget.LimitAmount, budget.PeriodType, budget.StartDate, budget.EndDate, budget.SpentAmount, budget.RemainingAmount,
		budget.PercentUsed, budget.IsOverspent, budget.RolloverMode, budget.NotifyOnExceed, budget.ContributionTotal, budget.CurrentBalance,
		budget.IsArchived, budget.ShowStatus, budget.UpdatedAt, budget.ID, userID)

	if err != nil {
		log.Printf("[UpdateBudget] UPDATE error for id=%s: %v", budget.ID, err)
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("[UpdateBudget] RowsAffected error for id=%s: %v", budget.ID, err)
		return appErrors.DatabaseError
	}

	if rows == 0 {
		log.Printf("[UpdateBudget] Budget not found: id=%s", budget.ID)
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
			log.Printf("[GetDebtByID] Debt not found: id=%s", id)
			return nil, appErrors.DebtNotFound
		}
		log.Printf("[GetDebtByID] DB error for id=%s: %v", id, err)
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

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[CreateDebt] Failed to begin transaction: %v", err)
		return appErrors.DatabaseError
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO debts (
			id, user_id, name, balance, direction, counterparty_id, counterparty_name, description,
			principal_amount, principal_currency, principal_original_amount, principal_original_currency,
			base_currency, rate_on_start, principal_base_value, repayment_currency, repayment_amount,
			repayment_rate_on_start, is_fixed_repayment_amount, start_date, due_date, interest_mode,
			interest_rate_annual, schedule_hint, linked_goal_id, linked_budget_id, funding_account_id,
			funding_transaction_id, lent_from_account_id, return_to_account_id, received_to_account_id,
			pay_from_account_id, custom_rate_used, exchange_rate_current, reminder_enabled, reminder_time, status, settled_at,
			final_rate_used, final_profit_loss, final_profit_loss_currency, total_paid_in_repayment_currency,
			remaining_amount, total_paid, percent_paid, show_status, created_at, updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,
			$13,$14,$15,$16,$17,
			$18,$19,$20,$21,$22,
			$23,$24,$25,$26,$27,
			$28,$29,$30,$31,$32,
			$33,$34,$35,$36,$37,
			$38,$39,$40,$41,$42,
			$43,$44,$45,$46,$47,$48
		)
	`, debt.ID, userID, debt.CounterpartyName, debt.PrincipalAmount, debt.Direction, debt.CounterpartyID, debt.CounterpartyName, debt.Description,
		debt.PrincipalAmount, debt.PrincipalCurrency, debt.PrincipalOriginalAmount, debt.PrincipalOriginalCurrency,
		debt.BaseCurrency, debt.RateOnStart, debt.PrincipalBaseValue, debt.RepaymentCurrency, debt.RepaymentAmount,
		debt.RepaymentRateOnStart, debt.IsFixedRepaymentAmount, debt.StartDate, debt.DueDate, debt.InterestMode,
		debt.InterestRateAnnual, debt.ScheduleHint, debt.LinkedGoalID, debt.LinkedBudgetID, debt.FundingAccountID,
		debt.FundingTransactionID, debt.LentFromAccountID, debt.ReturnToAccountID, debt.ReceivedToAccountID,
		debt.PayFromAccountID, debt.CustomRateUsed, debt.ExchangeRateCurrent, debt.ReminderEnabled, debt.ReminderTime, debt.Status, debt.SettledAt,
		debt.FinalRateUsed, debt.FinalProfitLoss, debt.FinalProfitLossCurrency, debt.TotalPaidInRepaymentCurrency,
		debt.RemainingAmount, debt.TotalPaid, debt.PercentPaid, debt.ShowStatus, debt.CreatedAt, debt.UpdatedAt); err != nil {
		log.Printf("[CreateDebt] INSERT error for counterparty=%s: %v", debt.CounterpartyName, err)
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	accountID := resolveDebtFundingAccountID(debt, nil)
	if accountID != nil && *accountID != "" && debt.PrincipalAmount != 0 {
		account, err := fetchAccountForUpdate(ctx, tx, userID, *accountID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if debt.PrincipalCurrency == "" {
			debt.PrincipalCurrency = account.Currency
		}
		if account.Currency != debt.PrincipalCurrency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		delta := debt.PrincipalAmount
		switch debt.Direction {
		case "they_owe_me":
			delta = -debt.PrincipalAmount
		case "i_owe":
			delta = debt.PrincipalAmount
		default:
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		referenceType := "debt"
		referenceID := debt.ID
		createdTxn := &Transaction{
			ID:            uuid.NewString(),
			UserID:        userID,
			Type:          TransactionTypeDebtCreate,
			AccountID:     accountID,
			ReferenceType: &referenceType,
			ReferenceID:   &referenceID,
			Amount:        delta,
			Currency:      account.Currency,
			BaseCurrency:  account.Currency,
			Date:          debt.StartDate,
			DebtID:        &debt.ID,
			RelatedDebtID: &debt.ID,
			ShowStatus:    "active",
			Attachments:   []string{},
			Tags:          []string{},
		}
		normalizeTransaction(createdTxn)
		if err := r.insertTransaction(ctx, tx, userID, createdTxn); err != nil {
			_ = tx.Rollback()
			return err
		}
		debt.FundingTransactionID = &createdTxn.ID
		if _, err := tx.ExecContext(ctx, `
			UPDATE debts
			SET funding_transaction_id = $1, updated_at = $2
			WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
		`, debt.FundingTransactionID, utils.NowUTC(), debt.ID, userID); err != nil {
			_ = tx.Rollback()
			return appErrors.DatabaseError
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, account.CurrentBalance+delta); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
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

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return appErrors.DatabaseError
	}

	current, err := fetchDebtForUpdate(ctx, tx, userID, debt.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE debts
		SET name = $1,
			balance = $2,
			direction = $3,
			counterparty_id = $4,
			counterparty_name = $5,
			description = $6,
			principal_amount = $7,
			principal_currency = $8,
			principal_original_amount = $9,
			principal_original_currency = $10,
			base_currency = $11,
			rate_on_start = $12,
			principal_base_value = $13,
			repayment_currency = $14,
			repayment_amount = $15,
			repayment_rate_on_start = $16,
			is_fixed_repayment_amount = $17,
			start_date = $18,
			due_date = $19,
			interest_mode = $20,
			interest_rate_annual = $21,
			schedule_hint = $22,
			linked_goal_id = $23,
			linked_budget_id = $24,
			funding_account_id = $25,
			funding_transaction_id = $26,
			lent_from_account_id = $27,
			return_to_account_id = $28,
			received_to_account_id = $29,
			pay_from_account_id = $30,
			custom_rate_used = $31,
			exchange_rate_current = $32,
			reminder_enabled = $33,
			reminder_time = $34,
			status = $35,
			settled_at = $36,
			final_rate_used = $37,
			final_profit_loss = $38,
			final_profit_loss_currency = $39,
			total_paid_in_repayment_currency = $40,
			remaining_amount = $41,
			total_paid = $42,
			percent_paid = $43,
			show_status = $44,
			updated_at = $45
		WHERE id = $46 AND user_id = $47 AND deleted_at IS NULL
	`, debt.CounterpartyName, debt.PrincipalAmount, debt.Direction, debt.CounterpartyID, debt.CounterpartyName, debt.Description, debt.PrincipalAmount,
		debt.PrincipalCurrency, debt.PrincipalOriginalAmount, debt.PrincipalOriginalCurrency, debt.BaseCurrency,
		debt.RateOnStart, debt.PrincipalBaseValue, debt.RepaymentCurrency, debt.RepaymentAmount, debt.RepaymentRateOnStart,
		debt.IsFixedRepaymentAmount, debt.StartDate, debt.DueDate, debt.InterestMode, debt.InterestRateAnnual,
		debt.ScheduleHint, debt.LinkedGoalID, debt.LinkedBudgetID, debt.FundingAccountID, debt.FundingTransactionID,
		debt.LentFromAccountID, debt.ReturnToAccountID, debt.ReceivedToAccountID, debt.PayFromAccountID,
		debt.CustomRateUsed, debt.ExchangeRateCurrent, debt.ReminderEnabled, debt.ReminderTime, debt.Status, debt.SettledAt, debt.FinalRateUsed,
		debt.FinalProfitLoss, debt.FinalProfitLossCurrency, debt.TotalPaidInRepaymentCurrency, debt.RemainingAmount,
		debt.TotalPaid, debt.PercentPaid, debt.ShowStatus, debt.UpdatedAt, debt.ID, userID)

	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if rows == 0 {
		_ = tx.Rollback()
		return appErrors.DebtNotFound
	}

	delta := debt.PrincipalAmount - current.PrincipalAmount
	accountID := resolveDebtFundingAccountID(debt, current)
	if delta != 0 && accountID != nil && *accountID != "" {
		account, err := fetchAccountForUpdate(ctx, tx, userID, *accountID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if debt.PrincipalCurrency == "" {
			debt.PrincipalCurrency = current.PrincipalCurrency
		}
		if account.Currency != debt.PrincipalCurrency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		adjustment := delta
		switch debt.Direction {
		case "they_owe_me":
			adjustment = -delta
		case "i_owe":
			adjustment = delta
		default:
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		referenceType := "debt"
		referenceID := debt.ID
		adjustmentTxn := &Transaction{
			ID:            uuid.NewString(),
			UserID:        userID,
			Type:          TransactionTypeDebtAdjustment,
			AccountID:     accountID,
			ReferenceType: &referenceType,
			ReferenceID:   &referenceID,
			Amount:        adjustment,
			Currency:      account.Currency,
			BaseCurrency:  account.Currency,
			Date:          time.Now().UTC().Format("2006-01-02"),
			DebtID:        &debt.ID,
			RelatedDebtID: &debt.ID,
			ShowStatus:    "active",
			Attachments:   []string{},
			Tags:          []string{},
		}
		normalizeTransaction(adjustmentTxn)
		if err := r.insertTransaction(ctx, tx, userID, adjustmentTxn); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, account.CurrentBalance+adjustment); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
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

// ========== DEBT PAYMENTS ==========

func (r *PostgresRepository) ListDebtPayments(ctx context.Context, debtID string) ([]*DebtPayment, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM debt_payments dp
		JOIN debts d ON dp.debt_id = d.id
		WHERE dp.debt_id = $1 AND d.user_id = $2 AND dp.deleted_at IS NULL
		ORDER BY dp.created_at DESC
	`, debtPaymentSelectFields)

	rows, err := r.db.QueryxContext(ctx, query, debtID, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var payments []*DebtPayment
	for rows.Next() {
		var row debtPaymentRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		payments = append(payments, mapRowToDebtPayment(row))
	}

	return payments, nil
}

func (r *PostgresRepository) GetDebtPaymentByID(ctx context.Context, debtID, paymentID string) (*DebtPayment, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM debt_payments dp
		JOIN debts d ON dp.debt_id = d.id
		WHERE dp.id = $1 AND dp.debt_id = $2 AND d.user_id = $3 AND dp.deleted_at IS NULL
	`, debtPaymentSelectFields)

	var row debtPaymentRow
	if err := r.db.GetContext(ctx, &row, query, paymentID, debtID, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.DebtPaymentNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToDebtPayment(row), nil
}

func (r *PostgresRepository) CreateDebtPayment(ctx context.Context, payment *DebtPayment) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if payment.ID == "" {
		payment.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[CreateDebtPayment] Failed to begin transaction: %v", err)
		return appErrors.DatabaseError
	}

	debtRow, err := fetchDebtForUpdate(ctx, tx, userID, payment.DebtID)
	if err != nil {
		log.Printf("[CreateDebtPayment] Failed to fetch debt=%s: %v", payment.DebtID, err)
		_ = tx.Rollback()
		return err
	}

	if payment.AccountID == nil || *payment.AccountID == "" {
		log.Printf("[CreateDebtPayment] Missing accountId for debt=%s", payment.DebtID)
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	account, err := fetchAccountForUpdate(ctx, tx, userID, *payment.AccountID)
	if err != nil {
		log.Printf("[CreateDebtPayment] Failed to fetch account=%s: %v", *payment.AccountID, err)
		_ = tx.Rollback()
		return err
	}
	if payment.Currency == "" {
		payment.Currency = account.Currency
	}
	if payment.Currency != account.Currency {
		log.Printf("[CreateDebtPayment] Currency mismatch: payment=%s, account=%s", payment.Currency, account.Currency)
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	delta := payment.Amount
	switch debtRow.Direction {
	case "they_owe_me":
		delta = payment.Amount
	case "i_owe":
		delta = -payment.Amount
	default:
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	referenceType := "debt"
	referenceID := payment.DebtID
	linkedDebtID := payment.DebtID
	txnType := TransactionTypeDebtPayment
	if strings.TrimSpace(payment.TransactionType) != "" {
		txnType = payment.TransactionType
	}
	conversionRate := 1.0
	if payment.RateUsedToDebt > 0 {
		conversionRate = 1 / payment.RateUsedToDebt
	}
	txn := &Transaction{
		ID:            uuid.NewString(),
		UserID:        userID,
		Type:          txnType,
		AccountID:     payment.AccountID,
		ReferenceType: &referenceType,
		ReferenceID:   &referenceID,
		Amount:        delta,
		Currency:      account.Currency,
		BaseCurrency:  account.Currency,
		Date:          payment.PaymentDate,
		DebtID:        &linkedDebtID,
		RelatedDebtID: &linkedDebtID,
		ShowStatus:    "active",
		Status:        TransactionStatusCompleted,
		Attachments:   []string{},
		Tags:          []string{},
		Description:   payment.Note,
		OriginalCurrency: func() *string {
			if debtRow.PrincipalCurrency != "" {
				value := debtRow.PrincipalCurrency
				return &value
			}
			return nil
		}(),
		OriginalAmount: payment.ConvertedAmountToDebt,
		ConversionRate: conversionRate,
	}
	normalizeTransaction(txn)
	if err := r.insertTransaction(ctx, tx, userID, txn); err != nil {
		_ = tx.Rollback()
		return err
	}
	payment.RelatedTransactionID = &txn.ID

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO debt_payments (
			id, debt_id, amount, currency, base_currency, rate_used_to_base,
			converted_amount_to_base, rate_used_to_debt, converted_amount_to_debt,
			payment_date, account_id, note, related_transaction_id, applied_rate, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, payment.ID, payment.DebtID, payment.Amount, payment.Currency, payment.BaseCurrency, payment.RateUsedToBase,
		payment.ConvertedAmountToBase, payment.RateUsedToDebt, payment.ConvertedAmountToDebt, payment.PaymentDate,
		payment.AccountID, payment.Note, payment.RelatedTransactionID, payment.AppliedRate, payment.CreatedAt, payment.UpdatedAt); err != nil {
		log.Printf("[CreateDebtPayment] INSERT error for debt=%s, amount=%.2f: %v", payment.DebtID, payment.Amount, err)
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if err := updateAccountBalance(ctx, tx, userID, account.ID, account.CurrentBalance+delta); err != nil {
		_ = tx.Rollback()
		return err
	}

	paymentInDebt := payment.ConvertedAmountToDebt
	if paymentInDebt == 0 && payment.RateUsedToDebt > 0 {
		paymentInDebt = payment.Amount * payment.RateUsedToDebt
	}
	remaining := debtRow.RemainingAmount
	if remaining <= 0 {
		remaining = debtRow.PrincipalAmount
	}
	remaining -= paymentInDebt
	if remaining < 0 {
		remaining = 0
	}
	totalPaid := debtRow.TotalPaid + paymentInDebt
	if totalPaid < 0 {
		totalPaid = 0
	}
	if debtRow.PrincipalAmount > 0 && totalPaid > debtRow.PrincipalAmount {
		totalPaid = debtRow.PrincipalAmount
	}
	percentPaid := 0.0
	if debtRow.PrincipalAmount > 0 {
		percentPaid = (totalPaid / debtRow.PrincipalAmount) * 100
		if percentPaid > 100 {
			percentPaid = 100
		}
	}
	totalPaidInRepayment := debtRow.TotalPaidInRepaymentCurrency
	if debtRow.RepaymentCurrency.Valid {
		repaymentCurrency := debtRow.RepaymentCurrency.String
		if strings.EqualFold(repaymentCurrency, payment.Currency) {
			totalPaidInRepayment += payment.Amount
		} else if strings.EqualFold(repaymentCurrency, debtRow.PrincipalCurrency) {
			totalPaidInRepayment += paymentInDebt
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE debts
		SET remaining_amount = $1,
			total_paid = $2,
			percent_paid = $3,
			total_paid_in_repayment_currency = $4,
			updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, remaining, totalPaid, percentPaid, totalPaidInRepayment, utils.NowUTC(), debtRow.ID, userID); err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) UpdateDebtPayment(ctx context.Context, payment *DebtPayment) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	payment.UpdatedAt = utils.NowUTC()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return appErrors.DatabaseError
	}

	debtRow, err := fetchDebtForUpdate(ctx, tx, userID, payment.DebtID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	existingPayment, err := fetchDebtPaymentForUpdate(ctx, tx, userID, payment.DebtID, payment.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	accountID := payment.AccountID
	if accountID == nil || *accountID == "" {
		if existingPayment.AccountID.Valid {
			accountID = &existingPayment.AccountID.String
		}
	}
	if accountID == nil || *accountID == "" {
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	account, err := fetchAccountForUpdate(ctx, tx, userID, *accountID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if payment.Currency == "" {
		payment.Currency = existingPayment.Currency
	}
	if payment.Currency == "" {
		payment.Currency = account.Currency
	}
	if payment.Currency != account.Currency {
		_ = tx.Rollback()
		return appErrors.InvalidFinanceData
	}

	amountDelta := payment.Amount - existingPayment.Amount
	if amountDelta != 0 {
		adjustment := amountDelta
		switch debtRow.Direction {
		case "they_owe_me":
			adjustment = amountDelta
		case "i_owe":
			adjustment = -amountDelta
		default:
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		referenceType := "debt"
		referenceID := payment.DebtID
		linkedDebtID := payment.DebtID
		adjustmentTxn := &Transaction{
			ID:            uuid.NewString(),
			UserID:        userID,
			Type:          TransactionTypeDebtAdjustment,
			AccountID:     accountID,
			ReferenceType: &referenceType,
			ReferenceID:   &referenceID,
			Amount:        adjustment,
			Currency:      account.Currency,
			BaseCurrency:  account.Currency,
			Date:          payment.PaymentDate,
			DebtID:        &linkedDebtID,
			RelatedDebtID: &linkedDebtID,
			ShowStatus:    "active",
			Attachments:   []string{},
			Tags:          []string{},
			Description:   payment.Note,
		}
		normalizeTransaction(adjustmentTxn)
		if err := r.insertTransaction(ctx, tx, userID, adjustmentTxn); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, account.CurrentBalance+adjustment); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE debt_payments
		SET amount = $1,
			currency = $2,
			base_currency = $3,
			rate_used_to_base = $4,
			converted_amount_to_base = $5,
			rate_used_to_debt = $6,
			converted_amount_to_debt = $7,
			payment_date = $8,
			account_id = $9,
			note = $10,
			related_transaction_id = $11,
			updated_at = $12
		WHERE id = $13 AND debt_id = $14 AND deleted_at IS NULL
	`, payment.Amount, payment.Currency, payment.BaseCurrency, payment.RateUsedToBase, payment.ConvertedAmountToBase,
		payment.RateUsedToDebt, payment.ConvertedAmountToDebt, payment.PaymentDate, accountID, payment.Note,
		payment.RelatedTransactionID, payment.UpdatedAt, payment.ID, payment.DebtID)

	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if rows == 0 {
		_ = tx.Rollback()
		return appErrors.DebtPaymentNotFound
	}

	deltaDebt := payment.ConvertedAmountToDebt - existingPayment.ConvertedAmountToDebt
	if deltaDebt != 0 {
		remaining := debtRow.RemainingAmount
		if remaining <= 0 {
			remaining = debtRow.PrincipalAmount
		}
		remaining -= deltaDebt
		if remaining < 0 {
			remaining = 0
		}
		totalPaid := debtRow.TotalPaid + deltaDebt
		if totalPaid < 0 {
			totalPaid = 0
		}
		if debtRow.PrincipalAmount > 0 && totalPaid > debtRow.PrincipalAmount {
			totalPaid = debtRow.PrincipalAmount
		}
		percentPaid := 0.0
		if debtRow.PrincipalAmount > 0 {
			percentPaid = (totalPaid / debtRow.PrincipalAmount) * 100
			if percentPaid > 100 {
				percentPaid = 100
			}
		}
		totalPaidInRepayment := debtRow.TotalPaidInRepaymentCurrency
		if debtRow.RepaymentCurrency.Valid {
			repaymentCurrency := debtRow.RepaymentCurrency.String
			if strings.EqualFold(repaymentCurrency, payment.Currency) {
				totalPaidInRepayment += (payment.Amount - existingPayment.Amount)
			} else if strings.EqualFold(repaymentCurrency, debtRow.PrincipalCurrency) {
				totalPaidInRepayment += deltaDebt
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE debts
			SET remaining_amount = $1,
				total_paid = $2,
				percent_paid = $3,
				total_paid_in_repayment_currency = $4,
				updated_at = $5
			WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
		`, remaining, totalPaid, percentPaid, totalPaidInRepayment, utils.NowUTC(), debtRow.ID, userID); err != nil {
			_ = tx.Rollback()
			return appErrors.DatabaseError
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (r *PostgresRepository) DeleteDebtPayment(ctx context.Context, debtID, paymentID string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return appErrors.DatabaseError
	}

	debtRow, err := fetchDebtForUpdate(ctx, tx, userID, debtID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	existingPayment, err := fetchDebtPaymentForUpdate(ctx, tx, userID, debtID, paymentID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if existingPayment.AccountID.Valid && existingPayment.AccountID.String != "" {
		account, err := fetchAccountForUpdate(ctx, tx, userID, existingPayment.AccountID.String)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if existingPayment.Currency != "" && existingPayment.Currency != account.Currency {
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		adjustment := existingPayment.Amount
		switch debtRow.Direction {
		case "they_owe_me":
			adjustment = -existingPayment.Amount
		case "i_owe":
			adjustment = existingPayment.Amount
		default:
			_ = tx.Rollback()
			return appErrors.InvalidFinanceData
		}
		referenceType := "debt"
		referenceID := debtID
		linkedDebtID := debtID
		adjustmentTxn := &Transaction{
			ID:            uuid.NewString(),
			UserID:        userID,
			Type:          TransactionTypeDebtAdjustment,
			AccountID:     &existingPayment.AccountID.String,
			ReferenceType: &referenceType,
			ReferenceID:   &referenceID,
			Amount:        adjustment,
			Currency:      account.Currency,
			BaseCurrency:  account.Currency,
			Date:          time.Now().UTC().Format("2006-01-02"),
			DebtID:        &linkedDebtID,
			RelatedDebtID: &linkedDebtID,
			ShowStatus:    "active",
			Attachments:   []string{},
			Tags:          []string{},
		}
		normalizeTransaction(adjustmentTxn)
		if err := r.insertTransaction(ctx, tx, userID, adjustmentTxn); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := updateAccountBalance(ctx, tx, userID, account.ID, account.CurrentBalance+adjustment); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE debt_payments
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND debt_id = $4 AND deleted_at IS NULL
	`, now, now, paymentID, debtID)

	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if rows == 0 {
		_ = tx.Rollback()
		return appErrors.DebtPaymentNotFound
	}

	deltaDebt := existingPayment.ConvertedAmountToDebt
	remaining := debtRow.RemainingAmount
	if remaining <= 0 {
		remaining = debtRow.PrincipalAmount
	}
	remaining += deltaDebt
	if remaining > debtRow.PrincipalAmount {
		remaining = debtRow.PrincipalAmount
	}
	totalPaid := debtRow.TotalPaid - deltaDebt
	if totalPaid < 0 {
		totalPaid = 0
	}
	percentPaid := 0.0
	if debtRow.PrincipalAmount > 0 {
		percentPaid = (totalPaid / debtRow.PrincipalAmount) * 100
		if percentPaid > 100 {
			percentPaid = 100
		}
	}
	totalPaidInRepayment := debtRow.TotalPaidInRepaymentCurrency
	if debtRow.RepaymentCurrency.Valid {
		repaymentCurrency := debtRow.RepaymentCurrency.String
		if strings.EqualFold(repaymentCurrency, existingPayment.Currency) {
			totalPaidInRepayment -= existingPayment.Amount
		} else if strings.EqualFold(repaymentCurrency, debtRow.PrincipalCurrency) {
			totalPaidInRepayment -= deltaDebt
		}
		if totalPaidInRepayment < 0 {
			totalPaidInRepayment = 0
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE debts
		SET remaining_amount = $1,
			total_paid = $2,
			percent_paid = $3,
			total_paid_in_repayment_currency = $4,
			updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, remaining, totalPaid, percentPaid, totalPaidInRepayment, utils.NowUTC(), debtRow.ID, userID); err != nil {
		_ = tx.Rollback()
		return appErrors.DatabaseError
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

// ========== COUNTERPARTIES ==========

func (r *PostgresRepository) ListCounterparties(ctx context.Context) ([]*Counterparty, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM counterparties
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, counterpartySelectFields)

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var counterparties []*Counterparty
	for rows.Next() {
		var row counterpartyRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		counterparties = append(counterparties, mapRowToCounterparty(row))
	}

	return counterparties, nil
}

func (r *PostgresRepository) GetCounterpartyByID(ctx context.Context, id string) (*Counterparty, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}

	query := fmt.Sprintf(`
		SELECT %s FROM counterparties
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, counterpartySelectFields)

	var row counterpartyRow
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.CounterpartyNotFound
		}
		return nil, appErrors.DatabaseError
	}

	return mapRowToCounterparty(row), nil
}

func (r *PostgresRepository) CreateCounterparty(ctx context.Context, counterparty *Counterparty) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	if counterparty.ID == "" {
		counterparty.ID = uuid.NewString()
	}

	now := utils.NowUTC()
	counterparty.CreatedAt = now
	counterparty.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO counterparties (id, user_id, display_name, phone_number, comment, search_keywords, show_status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, counterparty.ID, userID, counterparty.DisplayName, counterparty.PhoneNumber, counterparty.Comment, counterparty.SearchKeywords, counterparty.ShowStatus, counterparty.CreatedAt, counterparty.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) UpdateCounterparty(ctx context.Context, counterparty *Counterparty) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	counterparty.UpdatedAt = utils.NowUTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE counterparties
		SET display_name = $1,
			phone_number = $2,
			comment = $3,
			search_keywords = $4,
			show_status = $5,
			updated_at = $6
		WHERE id = $7 AND user_id = $8 AND deleted_at IS NULL
	`, counterparty.DisplayName, counterparty.PhoneNumber, counterparty.Comment, counterparty.SearchKeywords, counterparty.ShowStatus, counterparty.UpdatedAt, counterparty.ID, userID)

	if err != nil {
		return appErrors.DatabaseError
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}

	if rows == 0 {
		return appErrors.CounterpartyNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteCounterparty(ctx context.Context, id string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}

	now := utils.NowUTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE counterparties
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
		return appErrors.CounterpartyNotFound
	}

	return nil
}

// ========== FX RATES ==========

func (r *PostgresRepository) ListFXRates(ctx context.Context) ([]*FXRate, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM fx_rates
		ORDER BY rate_date DESC
	`, fxRateSelectFields)

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var rates []*FXRate
	for rows.Next() {
		var row fxRateRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		rates = append(rates, mapRowToFXRate(row))
	}

	return rates, nil
}

func (r *PostgresRepository) GetFXRateByID(ctx context.Context, id string) (*FXRate, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM fx_rates
		WHERE id = $1
	`, fxRateSelectFields)

	var row fxRateRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.FXRateNotFound
		}
		return nil, appErrors.DatabaseError
	}
	return mapRowToFXRate(row), nil
}

func (r *PostgresRepository) CreateFXRate(ctx context.Context, rate *FXRate) error {
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	rate.CreatedAt = now
	rate.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fx_rates (id, rate_date, from_currency, to_currency, rate, rate_mid, rate_bid, rate_ask, nominal, spread_percent, source, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, rate.ID, rate.Date, rate.FromCurrency, rate.ToCurrency, rate.Rate, rate.RateMid, rate.RateBid, rate.RateAsk, rate.Nominal, rate.SpreadPercent, rate.Source, rate.CreatedAt, rate.UpdatedAt)

	if err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) ListCategories(ctx context.Context, categoryType string, activeOnly bool) ([]*FinanceCategory, error) {
	clauses := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if categoryType != "" {
		clauses = append(clauses, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, categoryType)
		argIndex++
	}
	if activeOnly {
		clauses = append(clauses, "is_active = true")
	}

	query := fmt.Sprintf(`
		SELECT %s FROM finance_categories
		WHERE %s
		ORDER BY sort_order ASC, created_at ASC
	`, categorySelectFields, strings.Join(clauses, " AND "))

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	var categories []*FinanceCategory
	for rows.Next() {
		var row categoryRow
		if err := rows.StructScan(&row); err != nil {
			return nil, appErrors.DatabaseError
		}
		category, err := mapRowToCategory(row)
		if err != nil {
			return nil, appErrors.DatabaseError
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *PostgresRepository) CreateCategory(ctx context.Context, category *FinanceCategory) error {
	if category.ID == "" {
		category.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	category.CreatedAt = now
	category.UpdatedAt = now

	namePayload, err := json.Marshal(category.NameI18n)
	if err != nil {
		return appErrors.InvalidFinanceData
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO finance_categories (id, type, name_i18n, icon_name, color, is_default, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, category.ID, category.Type, namePayload, category.IconName, category.Color, category.IsDefault, category.SortOrder, category.IsActive, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func (r *PostgresRepository) UpdateCategory(ctx context.Context, category *FinanceCategory) error {
	category.UpdatedAt = utils.NowUTC()
	namePayload, err := json.Marshal(category.NameI18n)
	if err != nil {
		return appErrors.InvalidFinanceData
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE finance_categories
		SET type = $1,
			name_i18n = $2,
			icon_name = $3,
			color = $4,
			is_default = $5,
			sort_order = $6,
			is_active = $7,
			updated_at = $8
		WHERE id = $9
	`, category.Type, namePayload, category.IconName, category.Color, category.IsDefault, category.SortOrder, category.IsActive, category.UpdatedAt, category.ID)
	if err != nil {
		return appErrors.DatabaseError
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.DatabaseError
	}
	if rows == 0 {
		return appErrors.InvalidFinanceData
	}
	return nil
}

func (r *PostgresRepository) ListQuickExpenseCategories(ctx context.Context, categoryType string) ([]*QuickExpenseCategory, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, appErrors.InvalidToken
	}
	args := []interface{}{userID}
	query := `
		SELECT category_tag, category_name, category_type
		FROM finance_quick_exp_categories
		WHERE user_id = $1
	`
	if strings.TrimSpace(categoryType) != "" {
		query += " AND category_type = $2"
		args = append(args, categoryType)
	}
	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, appErrors.DatabaseError
	}
	defer rows.Close()

	results := make([]*QuickExpenseCategory, 0)
	for rows.Next() {
		var tag, name, ctype string
		if err := rows.Scan(&tag, &name, &ctype); err != nil {
			return nil, appErrors.DatabaseError
		}
		results = append(results, &QuickExpenseCategory{
			Tag:  tag,
			Name: name,
			Type: ctype,
		})
	}
	return results, nil
}

func (r *PostgresRepository) ReplaceQuickExpenseCategories(ctx context.Context, categoryType string, categories []*QuickExpenseCategory) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return appErrors.InvalidToken
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return appErrors.DatabaseError
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM finance_quick_exp_categories
		WHERE user_id = $1 AND category_type = $2
	`, userID, categoryType); err != nil {
		return appErrors.DatabaseError
	}

	now := utils.NowUTC()
	for _, entry := range categories {
		if entry == nil || strings.TrimSpace(entry.Tag) == "" {
			continue
		}
		ctype := strings.TrimSpace(entry.Type)
		if ctype == "" {
			ctype = categoryType
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO finance_quick_exp_categories
				(id, user_id, category_tag, category_name, category_type, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, uuid.NewString(), userID, strings.TrimSpace(entry.Tag), strings.TrimSpace(entry.Name), ctype, now, now); err != nil {
			return appErrors.DatabaseError
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.DatabaseError
	}
	return nil
}

func mapRowToCategory(row categoryRow) (*FinanceCategory, error) {
	name := map[string]string{}
	if len(row.NameI18n) > 0 {
		if err := json.Unmarshal(row.NameI18n, &name); err != nil {
			return nil, err
		}
	}
	var color *string
	if row.Color.Valid {
		color = &row.Color.String
	}
	return &FinanceCategory{
		ID:        row.ID,
		Type:      row.Type,
		NameI18n:  name,
		IconName:  row.IconName,
		Color:     color,
		IsDefault: row.IsDefault,
		SortOrder: row.SortOrder,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: row.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ========== ROW STRUCTS AND MAPPERS ==========

type accountRow struct {
	ID             string         `db:"id"`
	UserID         string         `db:"user_id"`
	Name           string         `db:"name"`
	Currency       string         `db:"currency"`
	AccountType    string         `db:"account_type"`
	InitialBalance float64        `db:"initial_balance"`
	CurrentBalance float64        `db:"current_balance"`
	LinkedGoalID   sql.NullString `db:"linked_goal_id"`
	CustomTypeID   sql.NullString `db:"custom_type_id"`
	IsMain         bool           `db:"is_main"`
	IsArchived     bool           `db:"is_archived"`
	ShowStatus     string         `db:"show_status"`
	CreatedAt      string         `db:"created_at"`
	UpdatedAt      string         `db:"updated_at"`
}

func mapRowToAccount(row accountRow) *Account {
	var linkedGoalID *string
	if row.LinkedGoalID.Valid {
		linkedGoalID = &row.LinkedGoalID.String
	}
	var customTypeID *string
	if row.CustomTypeID.Valid {
		customTypeID = &row.CustomTypeID.String
	}
	return &Account{
		ID:             row.ID,
		UserID:         row.UserID,
		Name:           row.Name,
		AccountType:    row.AccountType,
		Currency:       row.Currency,
		InitialBalance: row.InitialBalance,
		CurrentBalance: row.CurrentBalance,
		LinkedGoalID:   linkedGoalID,
		CustomTypeID:   customTypeID,
		IsMain:         row.IsMain,
		IsArchived:     row.IsArchived,
		ShowStatus:     row.ShowStatus,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

type transactionRow struct {
	ID                    string         `db:"id"`
	UserID                string         `db:"user_id"`
	Type                  string         `db:"type"`
	Status                string         `db:"status"`
	AccountID             sql.NullString `db:"account_id"`
	FromAccountID         sql.NullString `db:"from_account_id"`
	ToAccountID           sql.NullString `db:"to_account_id"`
	ReferenceType         sql.NullString `db:"reference_type"`
	ReferenceID           sql.NullString `db:"reference_id"`
	Amount                float64        `db:"amount"`
	Currency              string         `db:"currency"`
	BaseCurrency          sql.NullString `db:"base_currency"`
	RateUsedToBase        float64        `db:"rate_used_to_base"`
	ConvertedAmountToBase float64        `db:"converted_amount_to_base"`
	ToAmount              float64        `db:"to_amount"`
	ToCurrency            sql.NullString `db:"to_currency"`
	EffectiveRateFromTo   float64        `db:"effective_rate_from_to"`
	FeeAmount             float64        `db:"fee_amount"`
	FeeCategoryID         sql.NullString `db:"fee_category_id"`
	CategoryID            sql.NullString `db:"category_id"`
	Category              sql.NullString `db:"category"`
	SubcategoryID         sql.NullString `db:"subcategory_id"`
	Name                  sql.NullString `db:"name"`
	Description           sql.NullString `db:"description"`
	Date                  sql.NullTime   `db:"date"`
	Time                  sql.NullString `db:"time"`
	GoalID                sql.NullString `db:"linked_goal_id"`
	BudgetID              sql.NullString `db:"budget_id"`
	DebtID                sql.NullString `db:"linked_debt_id"`
	HabitID               sql.NullString `db:"habit_id"`
	CounterpartyID        sql.NullString `db:"counterparty_id"`
	RecurringID           sql.NullString `db:"recurring_id"`
	Attachments           []byte         `db:"attachments"`
	Tags                  []byte         `db:"tags"`
	IsBalanceAdjustment   bool           `db:"is_balance_adjustment"`
	SkipBudgetMatching    bool           `db:"skip_budget_matching"`
	ShowStatus            string         `db:"show_status"`
	RelatedBudgetID       sql.NullString `db:"related_budget_id"`
	RelatedDebtID         sql.NullString `db:"related_debt_id"`
	PlannedAmount         float64        `db:"planned_amount"`
	PaidAmount            float64        `db:"paid_amount"`
	OriginalCurrency      sql.NullString `db:"original_currency"`
	OriginalAmount        float64        `db:"original_amount"`
	ConversionRate        float64        `db:"conversion_rate"`
	OccurredAt            sql.NullTime   `db:"occurred_at"`
	Metadata              []byte         `db:"metadata"`
	CreatedAt             string         `db:"created_at"`
	UpdatedAt             string         `db:"updated_at"`
}

func mapRowToTransaction(row transactionRow) *Transaction {
	var accountID *string
	if row.AccountID.Valid {
		accountID = &row.AccountID.String
	}
	var fromAccountID *string
	if row.FromAccountID.Valid {
		fromAccountID = &row.FromAccountID.String
	}
	var toAccountID *string
	if row.ToAccountID.Valid {
		toAccountID = &row.ToAccountID.String
	}
	var referenceType *string
	if row.ReferenceType.Valid {
		referenceType = &row.ReferenceType.String
	}
	var referenceID *string
	if row.ReferenceID.Valid {
		referenceID = &row.ReferenceID.String
	}
	baseCurrency := row.Currency
	if row.BaseCurrency.Valid && row.BaseCurrency.String != "" {
		baseCurrency = row.BaseCurrency.String
	}
	var toCurrency *string
	if row.ToCurrency.Valid {
		toCurrency = &row.ToCurrency.String
	}
	var feeCategoryID *string
	if row.FeeCategoryID.Valid {
		feeCategoryID = &row.FeeCategoryID.String
	}
	var categoryID *string
	if row.CategoryID.Valid {
		categoryID = &row.CategoryID.String
	} else if row.Category.Valid {
		categoryID = &row.Category.String
	}
	var subcategoryID *string
	if row.SubcategoryID.Valid {
		subcategoryID = &row.SubcategoryID.String
	}
	var name *string
	if row.Name.Valid {
		name = &row.Name.String
	}
	var description *string
	if row.Description.Valid {
		description = &row.Description.String
	}
	date := ""
	if row.Date.Valid {
		date = row.Date.Time.Format("2006-01-02")
	}
	var timeValue *string
	if row.Time.Valid {
		timeValue = &row.Time.String
	}
	var goalID *string
	if row.GoalID.Valid {
		goalID = &row.GoalID.String
	}
	var budgetID *string
	if row.BudgetID.Valid {
		budgetID = &row.BudgetID.String
	}
	var debtID *string
	if row.DebtID.Valid {
		debtID = &row.DebtID.String
	}
	var habitID *string
	if row.HabitID.Valid {
		habitID = &row.HabitID.String
	}
	var counterpartyID *string
	if row.CounterpartyID.Valid {
		counterpartyID = &row.CounterpartyID.String
	}
	var recurringID *string
	if row.RecurringID.Valid {
		recurringID = &row.RecurringID.String
	}
	attachments := []string{}
	if len(row.Attachments) > 0 {
		_ = json.Unmarshal(row.Attachments, &attachments)
	}
	tags := []string{}
	if len(row.Tags) > 0 {
		_ = json.Unmarshal(row.Tags, &tags)
	}
	var relatedBudgetID *string
	if row.RelatedBudgetID.Valid {
		relatedBudgetID = &row.RelatedBudgetID.String
	}
	var relatedDebtID *string
	if row.RelatedDebtID.Valid {
		relatedDebtID = &row.RelatedDebtID.String
	}
	var originalCurrency *string
	if row.OriginalCurrency.Valid {
		originalCurrency = &row.OriginalCurrency.String
	}
	var occurredAt string
	if row.OccurredAt.Valid {
		occurredAt = row.OccurredAt.Time.UTC().Format(time.RFC3339)
	}
	metadata := map[string]interface{}{}
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}
	return &Transaction{
		ID:                    row.ID,
		UserID:                row.UserID,
		Type:                  row.Type,
		Status:                row.Status,
		AccountID:             accountID,
		FromAccountID:         fromAccountID,
		ToAccountID:           toAccountID,
		ReferenceType:         referenceType,
		ReferenceID:           referenceID,
		Amount:                row.Amount,
		Currency:              row.Currency,
		BaseCurrency:          baseCurrency,
		RateUsedToBase:        row.RateUsedToBase,
		ConvertedAmountToBase: row.ConvertedAmountToBase,
		ToAmount:              row.ToAmount,
		ToCurrency:            toCurrency,
		EffectiveRateFromTo:   row.EffectiveRateFromTo,
		FeeAmount:             row.FeeAmount,
		FeeCategoryID:         feeCategoryID,
		CategoryID:            categoryID,
		SubcategoryID:         subcategoryID,
		Name:                  name,
		Description:           description,
		Date:                  date,
		Time:                  timeValue,
		GoalID:                goalID,
		BudgetID:              budgetID,
		DebtID:                debtID,
		HabitID:               habitID,
		CounterpartyID:        counterpartyID,
		RecurringID:           recurringID,
		Attachments:           attachments,
		Tags:                  tags,
		IsBalanceAdjustment:   row.IsBalanceAdjustment,
		SkipBudgetMatching:    row.SkipBudgetMatching,
		ShowStatus:            row.ShowStatus,
		RelatedBudgetID:       relatedBudgetID,
		RelatedDebtID:         relatedDebtID,
		PlannedAmount:         row.PlannedAmount,
		PaidAmount:            row.PaidAmount,
		OriginalCurrency:      originalCurrency,
		OriginalAmount:        row.OriginalAmount,
		ConversionRate:        row.ConversionRate,
		OccurredAt:            occurredAt,
		Metadata:              metadata,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

type budgetRow struct {
	ID                string         `db:"id"`
	UserID            string         `db:"user_id"`
	Name              string         `db:"name"`
	BudgetType        string         `db:"budget_type"`
	CategoryIDs       []byte         `db:"category_ids"`
	LinkedGoalID      sql.NullString `db:"linked_goal_id"`
	AccountID         sql.NullString `db:"account_id"`
	TransactionType   sql.NullString `db:"transaction_type"`
	Currency          string         `db:"currency"`
	LimitAmount       float64        `db:"limit_amount"`
	PeriodType        string         `db:"period_type"`
	StartDate         sql.NullTime   `db:"start_date"`
	EndDate           sql.NullTime   `db:"end_date"`
	SpentAmount       float64        `db:"spent_amount"`
	RemainingAmount   float64        `db:"remaining_amount"`
	PercentUsed       float64        `db:"percent_used"`
	IsOverspent       bool           `db:"is_overspent"`
	RolloverMode      string         `db:"rollover_mode"`
	NotifyOnExceed    bool           `db:"notify_on_exceed"`
	ContributionTotal float64        `db:"contribution_total"`
	CurrentBalance    float64        `db:"current_balance"`
	IsArchived        bool           `db:"is_archived"`
	ShowStatus        string         `db:"show_status"`
	CreatedAt         string         `db:"created_at"`
	UpdatedAt         string         `db:"updated_at"`
}

func mapRowToBudget(row budgetRow) *Budget {
	categoryIDs := []string{}
	if len(row.CategoryIDs) > 0 {
		_ = json.Unmarshal(row.CategoryIDs, &categoryIDs)
	}
	var linkedGoalID *string
	if row.LinkedGoalID.Valid {
		linkedGoalID = &row.LinkedGoalID.String
	}
	var accountID *string
	if row.AccountID.Valid {
		accountID = &row.AccountID.String
	}
	var transactionType *string
	if row.TransactionType.Valid {
		transactionType = &row.TransactionType.String
	}
	var startDate *string
	if row.StartDate.Valid {
		formatted := row.StartDate.Time.Format("2006-01-02")
		startDate = &formatted
	}
	var endDate *string
	if row.EndDate.Valid {
		formatted := row.EndDate.Time.Format("2006-01-02")
		endDate = &formatted
	}
	return &Budget{
		ID:                row.ID,
		UserID:            row.UserID,
		Name:              row.Name,
		BudgetType:        row.BudgetType,
		CategoryIDs:       categoryIDs,
		LinkedGoalID:      linkedGoalID,
		AccountID:         accountID,
		TransactionType:   transactionType,
		Currency:          row.Currency,
		LimitAmount:       row.LimitAmount,
		PeriodType:        row.PeriodType,
		StartDate:         startDate,
		EndDate:           endDate,
		SpentAmount:       row.SpentAmount,
		RemainingAmount:   row.RemainingAmount,
		PercentUsed:       row.PercentUsed,
		IsOverspent:       row.IsOverspent,
		RolloverMode:      row.RolloverMode,
		NotifyOnExceed:    row.NotifyOnExceed,
		ContributionTotal: row.ContributionTotal,
		CurrentBalance:    row.CurrentBalance,
		IsArchived:        row.IsArchived,
		ShowStatus:        row.ShowStatus,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

type debtRow struct {
	ID                           string          `db:"id"`
	UserID                       string          `db:"user_id"`
	Name                         sql.NullString  `db:"name"`
	Balance                      sql.NullFloat64 `db:"balance"`
	Direction                    string          `db:"direction"`
	CounterpartyID               sql.NullString  `db:"counterparty_id"`
	CounterpartyName             string          `db:"counterparty_name"`
	Description                  sql.NullString  `db:"description"`
	PrincipalAmount              float64         `db:"principal_amount"`
	PrincipalCurrency            string          `db:"principal_currency"`
	PrincipalOriginalAmount      float64         `db:"principal_original_amount"`
	PrincipalOriginalCurrency    sql.NullString  `db:"principal_original_currency"`
	BaseCurrency                 string          `db:"base_currency"`
	RateOnStart                  float64         `db:"rate_on_start"`
	PrincipalBaseValue           float64         `db:"principal_base_value"`
	RepaymentCurrency            sql.NullString  `db:"repayment_currency"`
	RepaymentAmount              float64         `db:"repayment_amount"`
	RepaymentRateOnStart         float64         `db:"repayment_rate_on_start"`
	IsFixedRepaymentAmount       bool            `db:"is_fixed_repayment_amount"`
	StartDate                    sql.NullTime    `db:"start_date"`
	DueDate                      sql.NullTime    `db:"due_date"`
	InterestMode                 sql.NullString  `db:"interest_mode"`
	InterestRateAnnual           float64         `db:"interest_rate_annual"`
	ScheduleHint                 sql.NullString  `db:"schedule_hint"`
	LinkedGoalID                 sql.NullString  `db:"linked_goal_id"`
	LinkedBudgetID               sql.NullString  `db:"linked_budget_id"`
	FundingAccountID             sql.NullString  `db:"funding_account_id"`
	FundingTransactionID         sql.NullString  `db:"funding_transaction_id"`
	LentFromAccountID            sql.NullString  `db:"lent_from_account_id"`
	ReturnToAccountID            sql.NullString  `db:"return_to_account_id"`
	ReceivedToAccountID          sql.NullString  `db:"received_to_account_id"`
	PayFromAccountID             sql.NullString  `db:"pay_from_account_id"`
	CustomRateUsed               float64         `db:"custom_rate_used"`
	ExchangeRateCurrent          float64         `db:"exchange_rate_current"`
	ReminderEnabled              bool            `db:"reminder_enabled"`
	ReminderTime                 sql.NullString  `db:"reminder_time"`
	Status                       string          `db:"status"`
	SettledAt                    sql.NullTime    `db:"settled_at"`
	FinalRateUsed                float64         `db:"final_rate_used"`
	FinalProfitLoss              float64         `db:"final_profit_loss"`
	FinalProfitLossCurrency      sql.NullString  `db:"final_profit_loss_currency"`
	TotalPaidInRepaymentCurrency float64         `db:"total_paid_in_repayment_currency"`
	RemainingAmount              float64         `db:"remaining_amount"`
	TotalPaid                    float64         `db:"total_paid"`
	PercentPaid                  float64         `db:"percent_paid"`
	ShowStatus                   string          `db:"show_status"`
	CreatedAt                    string          `db:"created_at"`
	UpdatedAt                    string          `db:"updated_at"`
}

func mapRowToDebt(row debtRow) *Debt {
	var counterpartyID *string
	if row.CounterpartyID.Valid {
		counterpartyID = &row.CounterpartyID.String
	}
	var description *string
	if row.Description.Valid {
		description = &row.Description.String
	}
	var principalOriginalCurrency *string
	if row.PrincipalOriginalCurrency.Valid {
		principalOriginalCurrency = &row.PrincipalOriginalCurrency.String
	}
	var repaymentCurrency *string
	if row.RepaymentCurrency.Valid {
		repaymentCurrency = &row.RepaymentCurrency.String
	}
	var startDate string
	if row.StartDate.Valid {
		startDate = row.StartDate.Time.Format("2006-01-02")
	}
	var dueDate *string
	if row.DueDate.Valid {
		formatted := row.DueDate.Time.Format("2006-01-02")
		dueDate = &formatted
	}
	var interestMode *string
	if row.InterestMode.Valid {
		interestMode = &row.InterestMode.String
	}
	var scheduleHint *string
	if row.ScheduleHint.Valid {
		scheduleHint = &row.ScheduleHint.String
	}
	var linkedGoalID *string
	if row.LinkedGoalID.Valid {
		linkedGoalID = &row.LinkedGoalID.String
	}
	var linkedBudgetID *string
	if row.LinkedBudgetID.Valid {
		linkedBudgetID = &row.LinkedBudgetID.String
	}
	var fundingAccountID *string
	if row.FundingAccountID.Valid {
		fundingAccountID = &row.FundingAccountID.String
	}
	var fundingTransactionID *string
	if row.FundingTransactionID.Valid {
		fundingTransactionID = &row.FundingTransactionID.String
	}
	var lentFromAccountID *string
	if row.LentFromAccountID.Valid {
		lentFromAccountID = &row.LentFromAccountID.String
	}
	var returnToAccountID *string
	if row.ReturnToAccountID.Valid {
		returnToAccountID = &row.ReturnToAccountID.String
	}
	var receivedToAccountID *string
	if row.ReceivedToAccountID.Valid {
		receivedToAccountID = &row.ReceivedToAccountID.String
	}
	var payFromAccountID *string
	if row.PayFromAccountID.Valid {
		payFromAccountID = &row.PayFromAccountID.String
	}
	var reminderTime *string
	if row.ReminderTime.Valid {
		reminderTime = &row.ReminderTime.String
	}
	var settledAt *string
	if row.SettledAt.Valid {
		formatted := row.SettledAt.Time.Format("2006-01-02T15:04:05Z07:00")
		settledAt = &formatted
	}
	var finalProfitLossCurrency *string
	if row.FinalProfitLossCurrency.Valid {
		finalProfitLossCurrency = &row.FinalProfitLossCurrency.String
	}
	counterpartyName := row.CounterpartyName
	if counterpartyName == "" && row.Name.Valid {
		counterpartyName = row.Name.String
	}
	return &Debt{
		ID:                           row.ID,
		UserID:                       row.UserID,
		Direction:                    row.Direction,
		CounterpartyID:               counterpartyID,
		CounterpartyName:             counterpartyName,
		Description:                  description,
		PrincipalAmount:              row.PrincipalAmount,
		PrincipalCurrency:            row.PrincipalCurrency,
		PrincipalOriginalAmount:      row.PrincipalOriginalAmount,
		PrincipalOriginalCurrency:    principalOriginalCurrency,
		BaseCurrency:                 row.BaseCurrency,
		RateOnStart:                  row.RateOnStart,
		PrincipalBaseValue:           row.PrincipalBaseValue,
		RepaymentCurrency:            repaymentCurrency,
		RepaymentAmount:              row.RepaymentAmount,
		RepaymentRateOnStart:         row.RepaymentRateOnStart,
		IsFixedRepaymentAmount:       row.IsFixedRepaymentAmount,
		StartDate:                    startDate,
		DueDate:                      dueDate,
		InterestMode:                 interestMode,
		InterestRateAnnual:           row.InterestRateAnnual,
		ScheduleHint:                 scheduleHint,
		LinkedGoalID:                 linkedGoalID,
		LinkedBudgetID:               linkedBudgetID,
		FundingAccountID:             fundingAccountID,
		FundingTransactionID:         fundingTransactionID,
		LentFromAccountID:            lentFromAccountID,
		ReturnToAccountID:            returnToAccountID,
		ReceivedToAccountID:          receivedToAccountID,
		PayFromAccountID:             payFromAccountID,
		CustomRateUsed:               row.CustomRateUsed,
		ExchangeRateCurrent:          row.ExchangeRateCurrent,
		ReminderEnabled:              row.ReminderEnabled,
		ReminderTime:                 reminderTime,
		Status:                       row.Status,
		SettledAt:                    settledAt,
		FinalRateUsed:                row.FinalRateUsed,
		FinalProfitLoss:              row.FinalProfitLoss,
		FinalProfitLossCurrency:      finalProfitLossCurrency,
		TotalPaidInRepaymentCurrency: row.TotalPaidInRepaymentCurrency,
		RemainingAmount:              row.RemainingAmount,
		TotalPaid:                    row.TotalPaid,
		PercentPaid:                  row.PercentPaid,
		ShowStatus:                   row.ShowStatus,
		CreatedAt:                    row.CreatedAt,
		UpdatedAt:                    row.UpdatedAt,
	}
}

type debtPaymentRow struct {
	ID                    string         `db:"id"`
	DebtID                string         `db:"debt_id"`
	Amount                float64        `db:"amount"`
	Currency              string         `db:"currency"`
	BaseCurrency          string         `db:"base_currency"`
	RateUsedToBase        float64        `db:"rate_used_to_base"`
	ConvertedAmountToBase float64        `db:"converted_amount_to_base"`
	RateUsedToDebt        float64        `db:"rate_used_to_debt"`
	ConvertedAmountToDebt float64        `db:"converted_amount_to_debt"`
	PaymentDate           sql.NullTime   `db:"payment_date"`
	AccountID             sql.NullString `db:"account_id"`
	Note                  sql.NullString `db:"note"`
	RelatedTransactionID  sql.NullString `db:"related_transaction_id"`
	AppliedRate           float64        `db:"applied_rate"`
	CreatedAt             string         `db:"created_at"`
	UpdatedAt             string         `db:"updated_at"`
	DeletedAt             sql.NullString `db:"deleted_at"`
}

func mapRowToDebtPayment(row debtPaymentRow) *DebtPayment {
	var accountID *string
	if row.AccountID.Valid {
		accountID = &row.AccountID.String
	}
	var note *string
	if row.Note.Valid {
		note = &row.Note.String
	}
	var relatedTransactionID *string
	if row.RelatedTransactionID.Valid {
		relatedTransactionID = &row.RelatedTransactionID.String
	}
	paymentDate := ""
	if row.PaymentDate.Valid {
		paymentDate = row.PaymentDate.Time.Format("2006-01-02")
	}
	return &DebtPayment{
		ID:                    row.ID,
		DebtID:                row.DebtID,
		Amount:                row.Amount,
		Currency:              row.Currency,
		BaseCurrency:          row.BaseCurrency,
		RateUsedToBase:        row.RateUsedToBase,
		ConvertedAmountToBase: row.ConvertedAmountToBase,
		RateUsedToDebt:        row.RateUsedToDebt,
		ConvertedAmountToDebt: row.ConvertedAmountToDebt,
		PaymentDate:           paymentDate,
		AccountID:             accountID,
		Note:                  note,
		RelatedTransactionID:  relatedTransactionID,
		AppliedRate:           row.AppliedRate,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

type counterpartyRow struct {
	ID             string         `db:"id"`
	UserID         string         `db:"user_id"`
	DisplayName    string         `db:"display_name"`
	PhoneNumber    sql.NullString `db:"phone_number"`
	Comment        sql.NullString `db:"comment"`
	SearchKeywords sql.NullString `db:"search_keywords"`
	ShowStatus     string         `db:"show_status"`
	CreatedAt      string         `db:"created_at"`
	UpdatedAt      string         `db:"updated_at"`
	DeletedAt      sql.NullString `db:"deleted_at"`
}

func mapRowToCounterparty(row counterpartyRow) *Counterparty {
	var phoneNumber *string
	if row.PhoneNumber.Valid {
		phoneNumber = &row.PhoneNumber.String
	}
	var comment *string
	if row.Comment.Valid {
		comment = &row.Comment.String
	}
	var searchKeywords *string
	if row.SearchKeywords.Valid {
		searchKeywords = &row.SearchKeywords.String
	}
	return &Counterparty{
		ID:             row.ID,
		UserID:         row.UserID,
		DisplayName:    row.DisplayName,
		PhoneNumber:    phoneNumber,
		Comment:        comment,
		SearchKeywords: searchKeywords,
		ShowStatus:     row.ShowStatus,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

type fxRateRow struct {
	ID            string       `db:"id"`
	RateDate      sql.NullTime `db:"rate_date"`
	FromCurrency  string       `db:"from_currency"`
	ToCurrency    string       `db:"to_currency"`
	Rate          float64      `db:"rate"`
	RateMid       float64      `db:"rate_mid"`
	RateBid       float64      `db:"rate_bid"`
	RateAsk       float64      `db:"rate_ask"`
	Nominal       float64      `db:"nominal"`
	SpreadPercent float64      `db:"spread_percent"`
	Source        string       `db:"source"`
	CreatedAt     string       `db:"created_at"`
	UpdatedAt     string       `db:"updated_at"`
}

func mapRowToFXRate(row fxRateRow) *FXRate {
	date := ""
	if row.RateDate.Valid {
		date = row.RateDate.Time.Format("2006-01-02")
	}
	return &FXRate{
		ID:            row.ID,
		Date:          date,
		FromCurrency:  row.FromCurrency,
		ToCurrency:    row.ToCurrency,
		Rate:          row.Rate,
		RateMid:       row.RateMid,
		RateBid:       row.RateBid,
		RateAsk:       row.RateAsk,
		Nominal:       row.Nominal,
		SpreadPercent: row.SpreadPercent,
		Source:        row.Source,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
