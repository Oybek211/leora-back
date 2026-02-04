package finance

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository defines finance persistence.
type Repository interface {
	ListAccounts(ctx context.Context) ([]*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	CreateAccount(ctx context.Context, account *Account) (*Transaction, error)
	UpdateAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id string) (*Transaction, error)

	ListTransactions(ctx context.Context) ([]*Transaction, error)
	GetTransactionByID(ctx context.Context, id string) (*Transaction, error)
	CreateTransaction(ctx context.Context, txn *Transaction) error
	UpdateTransaction(ctx context.Context, txn *Transaction) error
	DeleteTransaction(ctx context.Context, id string) error

	ListBudgets(ctx context.Context) ([]*Budget, error)
	GetBudgetByID(ctx context.Context, id string) (*Budget, error)
	CreateBudget(ctx context.Context, budget *Budget) error
	UpdateBudget(ctx context.Context, budget *Budget) error
	DeleteBudget(ctx context.Context, id string) error

	ListDebts(ctx context.Context) ([]*Debt, error)
	GetDebtByID(ctx context.Context, id string) (*Debt, error)
	CreateDebt(ctx context.Context, debt *Debt) error
	UpdateDebt(ctx context.Context, debt *Debt) error
	DeleteDebt(ctx context.Context, id string) error

	ListDebtPayments(ctx context.Context, debtID string) ([]*DebtPayment, error)
	GetDebtPaymentByID(ctx context.Context, debtID, paymentID string) (*DebtPayment, error)
	CreateDebtPayment(ctx context.Context, payment *DebtPayment) error
	UpdateDebtPayment(ctx context.Context, payment *DebtPayment) error
	DeleteDebtPayment(ctx context.Context, debtID, paymentID string) error

	ListCounterparties(ctx context.Context) ([]*Counterparty, error)
	GetCounterpartyByID(ctx context.Context, id string) (*Counterparty, error)
	CreateCounterparty(ctx context.Context, counterparty *Counterparty) error
	UpdateCounterparty(ctx context.Context, counterparty *Counterparty) error
	DeleteCounterparty(ctx context.Context, id string) error

	ListFXRates(ctx context.Context) ([]*FXRate, error)
	GetFXRateByID(ctx context.Context, id string) (*FXRate, error)
	CreateFXRate(ctx context.Context, rate *FXRate) error

	ListCategories(ctx context.Context, categoryType string, activeOnly bool) ([]*FinanceCategory, error)
	CreateCategory(ctx context.Context, category *FinanceCategory) error
	UpdateCategory(ctx context.Context, category *FinanceCategory) error

	ListQuickExpenseCategories(ctx context.Context, categoryType string) ([]*QuickExpenseCategory, error)
	ReplaceQuickExpenseCategories(ctx context.Context, categoryType string, categories []*QuickExpenseCategory) error
}

// InMemoryRepository stores finance data in memory.
type InMemoryRepository struct {
	mu             sync.RWMutex
	accounts       map[string]*Account
	transactions   map[string]*Transaction
	budgets        map[string]*Budget
	debts          map[string]*Debt
	debtPayments   map[string]map[string]*DebtPayment
	counterparties map[string]*Counterparty
	fxRates        map[string]*FXRate
	categories     map[string]*FinanceCategory
	quickExp       map[string][]*QuickExpenseCategory
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		accounts:       make(map[string]*Account),
		transactions:   make(map[string]*Transaction),
		budgets:        make(map[string]*Budget),
		debts:          make(map[string]*Debt),
		debtPayments:   make(map[string]map[string]*DebtPayment),
		counterparties: make(map[string]*Counterparty),
		fxRates:        make(map[string]*FXRate),
		categories:     make(map[string]*FinanceCategory),
		quickExp:       make(map[string][]*QuickExpenseCategory),
	}
}

func (r *InMemoryRepository) ListAccounts(ctx context.Context) ([]*Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		if account == nil || account.DeletedAt != "" {
			continue
		}
		results = append(results, cloneAccount(account))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	account, ok := r.accounts[id]
	if !ok || account == nil || account.DeletedAt != "" {
		return nil, appErrors.AccountNotFound
	}
	return cloneAccount(account), nil
}

func (r *InMemoryRepository) CreateAccount(ctx context.Context, account *Account) (*Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		account.UserID = userID
	}
	if account.ID == "" {
		account.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	account.CreatedAt = now
	account.UpdatedAt = now
	r.accounts[account.ID] = cloneAccount(account)
	if account.InitialBalance == 0 {
		return nil, nil
	}
	openingDate := time.Now().UTC().Format("2006-01-02")
	txn := &Transaction{
		ID:           uuid.NewString(),
		UserID:       account.UserID,
		Type:         TransactionTypeAccountCreateFunding,
		AccountID:    &account.ID,
		Amount:       account.InitialBalance,
		Currency:     account.Currency,
		BaseCurrency: account.Currency,
		Date:         openingDate,
		ShowStatus:   "active",
		Status:       TransactionStatusCompleted,
		CreatedAt:    now,
		UpdatedAt:    now,
		Attachments:  []string{},
		Tags:         []string{},
	}
	r.transactions[txn.ID] = cloneTransaction(txn)
	return cloneTransaction(txn), nil
}

func (r *InMemoryRepository) UpdateAccount(ctx context.Context, account *Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.accounts[account.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.AccountNotFound
	}
	account.CreatedAt = current.CreatedAt
	account.UpdatedAt = utils.NowUTC()
	r.accounts[account.ID] = cloneAccount(account)
	return nil
}

func (r *InMemoryRepository) DeleteAccount(ctx context.Context, id string) (*Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	account, ok := r.accounts[id]
	if !ok || account == nil || account.DeletedAt != "" {
		return nil, appErrors.AccountNotFound
	}
	var withdrawal *Transaction
	if account.CurrentBalance != 0 {
		withdrawal = &Transaction{
			ID:           uuid.NewString(),
			UserID:       account.UserID,
			Type:         TransactionTypeAccountDeleteWithdrawal,
			AccountID:    &account.ID,
			Amount:       account.CurrentBalance,
			Currency:     account.Currency,
			BaseCurrency: account.Currency,
			Date:         time.Now().UTC().Format("2006-01-02"),
			ShowStatus:   "active",
			Status:       TransactionStatusCompleted,
			Attachments:  []string{},
			Tags:         []string{},
		}
		account.CurrentBalance = 0
		r.transactions[withdrawal.ID] = cloneTransaction(withdrawal)
	}
	account.DeletedAt = utils.NowUTC()
	account.UpdatedAt = account.DeletedAt
	r.accounts[id] = account
	return withdrawal, nil
}

func (r *InMemoryRepository) ListTransactions(ctx context.Context) ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Transaction, 0, len(r.transactions))
	for _, txn := range r.transactions {
		if txn == nil || txn.DeletedAt != "" {
			continue
		}
		results = append(results, cloneTransaction(txn))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetTransactionByID(ctx context.Context, id string) (*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	txn, ok := r.transactions[id]
	if !ok || txn == nil || txn.DeletedAt != "" {
		return nil, appErrors.TransactionNotFound
	}
	return cloneTransaction(txn), nil
}

func (r *InMemoryRepository) CreateTransaction(ctx context.Context, txn *Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	userID, _ := ctx.Value("user_id").(string)
	if txn.Amount == 0 {
		return appErrors.InvalidFinanceData
	}
	txn.Type = strings.ToLower(txn.Type)
	hasAccounts := false
	for _, account := range r.accounts {
		if account != nil && account.DeletedAt == "" && account.UserID == userID {
			hasAccounts = true
			break
		}
	}
	if txn.ID == "" {
		txn.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	if userID != "" {
		txn.UserID = userID
	}
	txn.CreatedAt = now
	txn.UpdatedAt = now
	if txn.Status == "" {
		txn.Status = TransactionStatusCompleted
	}
	if txn.Date == "" {
		txn.Date = time.Now().UTC().Format("2006-01-02")
	}
	switch txn.Type {
	case "income", "expense":
		if txn.AccountID == nil || *txn.AccountID == "" {
			return appErrors.InvalidFinanceData
		}
		account, ok := r.accounts[*txn.AccountID]
		if !ok || account == nil || account.DeletedAt != "" {
			if !hasAccounts {
				return appErrors.AccountRequired
			}
			return appErrors.AccountNotFound
		}
		if txn.Currency == "" {
			txn.Currency = account.Currency
		}
		if txn.Currency != account.Currency {
			return appErrors.InvalidFinanceData
		}
		if txn.Type == "expense" && account.CurrentBalance < txn.Amount {
			return appErrors.InsufficientFunds
		}
		if txn.Type == "income" {
			account.CurrentBalance += txn.Amount
		} else {
			account.CurrentBalance -= txn.Amount
		}
		r.accounts[account.ID] = account
	case "transfer":
		if txn.FromAccountID == nil || txn.ToAccountID == nil || *txn.FromAccountID == "" || *txn.ToAccountID == "" {
			return appErrors.InvalidFinanceData
		}
		fromAccount, ok := r.accounts[*txn.FromAccountID]
		if !ok || fromAccount == nil || fromAccount.DeletedAt != "" {
			if !hasAccounts {
				return appErrors.AccountRequired
			}
			return appErrors.AccountNotFound
		}
		toAccount, ok := r.accounts[*txn.ToAccountID]
		if !ok || toAccount == nil || toAccount.DeletedAt != "" {
			if !hasAccounts {
				return appErrors.AccountRequired
			}
			return appErrors.AccountNotFound
		}
		if fromAccount.Currency != toAccount.Currency {
			return appErrors.InvalidFinanceData
		}
		if txn.Currency == "" {
			txn.Currency = fromAccount.Currency
		}
		if txn.Currency != fromAccount.Currency {
			return appErrors.InvalidFinanceData
		}
		if fromAccount.CurrentBalance < txn.Amount {
			return appErrors.InsufficientFunds
		}
		if txn.ToAmount == 0 {
			txn.ToAmount = txn.Amount
		}
		fromAccount.CurrentBalance -= txn.Amount
		toAccount.CurrentBalance += txn.ToAmount
		r.accounts[fromAccount.ID] = fromAccount
		r.accounts[toAccount.ID] = toAccount
	case TransactionTypeSystemAdjustment,
		TransactionTypeDebtCreate,
		TransactionTypeDebtPayment,
		TransactionTypeDebtAdjustment,
		TransactionTypeDebtFullPayment,
		TransactionTypeBudgetAddValue,
		TransactionTypeDebtAddValue,
		TransactionTypeAccountDeleteWithdrawal:
		if txn.AccountID == nil || *txn.AccountID == "" {
			return appErrors.InvalidFinanceData
		}
		account, ok := r.accounts[*txn.AccountID]
		if !ok || account == nil || account.DeletedAt != "" {
			if !hasAccounts {
				return appErrors.AccountRequired
			}
			return appErrors.AccountNotFound
		}
		if txn.Currency == "" {
			txn.Currency = account.Currency
		}
		if txn.Currency != account.Currency {
			return appErrors.InvalidFinanceData
		}
		if txn.Amount < 0 && account.CurrentBalance < -txn.Amount {
			return appErrors.InsufficientFunds
		}
		account.CurrentBalance += txn.Amount
		r.accounts[account.ID] = account
	default:
		return appErrors.InvalidFinanceData
	}
	r.transactions[txn.ID] = cloneTransaction(txn)
	return nil
}

func (r *InMemoryRepository) UpdateTransaction(ctx context.Context, txn *Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.transactions[txn.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.TransactionNotFound
	}
	txn.CreatedAt = current.CreatedAt
	txn.UpdatedAt = utils.NowUTC()
	r.transactions[txn.ID] = cloneTransaction(txn)
	return nil
}

func (r *InMemoryRepository) DeleteTransaction(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	txn, ok := r.transactions[id]
	if !ok || txn == nil || txn.DeletedAt != "" {
		return appErrors.TransactionNotFound
	}
	txn.DeletedAt = utils.NowUTC()
	txn.UpdatedAt = txn.DeletedAt
	r.transactions[id] = txn
	return nil
}

func (r *InMemoryRepository) ListBudgets(ctx context.Context) ([]*Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Budget, 0, len(r.budgets))
	for _, budget := range r.budgets {
		if budget == nil || budget.DeletedAt != "" {
			continue
		}
		results = append(results, cloneBudget(budget))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetBudgetByID(ctx context.Context, id string) (*Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	budget, ok := r.budgets[id]
	if !ok || budget == nil || budget.DeletedAt != "" {
		return nil, appErrors.BudgetNotFound
	}
	return cloneBudget(budget), nil
}

func (r *InMemoryRepository) CreateBudget(ctx context.Context, budget *Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if budget.ID == "" {
		budget.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	budget.CreatedAt = now
	budget.UpdatedAt = now
	r.budgets[budget.ID] = cloneBudget(budget)
	return nil
}

func (r *InMemoryRepository) UpdateBudget(ctx context.Context, budget *Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.budgets[budget.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.BudgetNotFound
	}
	budget.CreatedAt = current.CreatedAt
	budget.UpdatedAt = utils.NowUTC()
	r.budgets[budget.ID] = cloneBudget(budget)
	return nil
}

func (r *InMemoryRepository) DeleteBudget(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	budget, ok := r.budgets[id]
	if !ok || budget == nil || budget.DeletedAt != "" {
		return appErrors.BudgetNotFound
	}
	budget.DeletedAt = utils.NowUTC()
	budget.UpdatedAt = budget.DeletedAt
	r.budgets[id] = budget
	return nil
}

func (r *InMemoryRepository) ListDebts(ctx context.Context) ([]*Debt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Debt, 0, len(r.debts))
	for _, debt := range r.debts {
		if debt == nil || debt.DeletedAt != "" {
			continue
		}
		results = append(results, cloneDebt(debt))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetDebtByID(ctx context.Context, id string) (*Debt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	debt, ok := r.debts[id]
	if !ok || debt == nil || debt.DeletedAt != "" {
		return nil, appErrors.DebtNotFound
	}
	return cloneDebt(debt), nil
}

func (r *InMemoryRepository) CreateDebt(ctx context.Context, debt *Debt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if debt.ID == "" {
		debt.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	debt.CreatedAt = now
	debt.UpdatedAt = now
	r.debts[debt.ID] = cloneDebt(debt)
	return nil
}

func (r *InMemoryRepository) UpdateDebt(ctx context.Context, debt *Debt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.debts[debt.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.DebtNotFound
	}
	debt.CreatedAt = current.CreatedAt
	debt.UpdatedAt = utils.NowUTC()
	r.debts[debt.ID] = cloneDebt(debt)
	return nil
}

func (r *InMemoryRepository) DeleteDebt(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	debt, ok := r.debts[id]
	if !ok || debt == nil || debt.DeletedAt != "" {
		return appErrors.DebtNotFound
	}
	debt.DeletedAt = utils.NowUTC()
	debt.UpdatedAt = debt.DeletedAt
	r.debts[id] = debt
	return nil
}

func (r *InMemoryRepository) ListDebtPayments(ctx context.Context, debtID string) ([]*DebtPayment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	debtPayments := r.debtPayments[debtID]
	results := make([]*DebtPayment, 0, len(debtPayments))
	for _, payment := range debtPayments {
		if payment == nil || payment.DeletedAt != "" {
			continue
		}
		results = append(results, cloneDebtPayment(payment))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetDebtPaymentByID(ctx context.Context, debtID, paymentID string) (*DebtPayment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	debtPayments := r.debtPayments[debtID]
	if debtPayments == nil {
		return nil, appErrors.DebtPaymentNotFound
	}
	payment, ok := debtPayments[paymentID]
	if !ok || payment == nil || payment.DeletedAt != "" {
		return nil, appErrors.DebtPaymentNotFound
	}
	return cloneDebtPayment(payment), nil
}

func (r *InMemoryRepository) CreateDebtPayment(ctx context.Context, payment *DebtPayment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	debt, ok := r.debts[payment.DebtID]
	if !ok || debt == nil || debt.DeletedAt != "" {
		return appErrors.DebtNotFound
	}
	if payment.AccountID == nil || *payment.AccountID == "" {
		return appErrors.InvalidFinanceData
	}
	account, ok := r.accounts[*payment.AccountID]
	if !ok || account == nil || account.DeletedAt != "" {
		return appErrors.AccountNotFound
	}
	if payment.ID == "" {
		payment.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	payment.CreatedAt = now
	payment.UpdatedAt = now
	if _, ok := r.debtPayments[payment.DebtID]; !ok {
		r.debtPayments[payment.DebtID] = make(map[string]*DebtPayment)
	}
	if payment.Currency == "" {
		payment.Currency = account.Currency
	}
	if payment.Currency != account.Currency {
		return appErrors.InvalidFinanceData
	}
	delta := payment.Amount
	switch debt.Direction {
	case "they_owe_me":
		delta = payment.Amount
	case "i_owe":
		delta = -payment.Amount
	default:
		return appErrors.InvalidFinanceData
	}
	if delta < 0 && account.CurrentBalance < -delta {
		return appErrors.InsufficientFunds
	}
	account.CurrentBalance += delta
	r.accounts[account.ID] = account

	referenceType := "debt"
	referenceID := debt.ID
	linkedDebtID := debt.ID
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
		UserID:        account.UserID,
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
		OriginalCurrency: func() *string {
			if debt.PrincipalCurrency != "" {
				value := debt.PrincipalCurrency
				return &value
			}
			return nil
		}(),
		OriginalAmount: payment.ConvertedAmountToDebt,
		ConversionRate: conversionRate,
	}
	r.transactions[txn.ID] = cloneTransaction(txn)
	payment.RelatedTransactionID = &txn.ID

	r.debtPayments[payment.DebtID][payment.ID] = cloneDebtPayment(payment)
	// Update debt totals in debt currency.
	paymentInDebt := payment.ConvertedAmountToDebt
	if paymentInDebt == 0 && payment.RateUsedToDebt > 0 {
		paymentInDebt = payment.Amount * payment.RateUsedToDebt
	}
	remaining := debt.RemainingAmount
	if remaining <= 0 {
		remaining = debt.PrincipalAmount
	}
	remaining -= paymentInDebt
	if remaining < 0 {
		remaining = 0
	}
	totalPaid := debt.TotalPaid + paymentInDebt
	if totalPaid < 0 {
		totalPaid = 0
	}
	if debt.PrincipalAmount > 0 && totalPaid > debt.PrincipalAmount {
		totalPaid = debt.PrincipalAmount
	}
	percentPaid := 0.0
	if debt.PrincipalAmount > 0 {
		percentPaid = (totalPaid / debt.PrincipalAmount) * 100
		if percentPaid > 100 {
			percentPaid = 100
		}
	}
	totalPaidInRepayment := debt.TotalPaidInRepaymentCurrency
	if debt.RepaymentCurrency != nil {
		if strings.EqualFold(*debt.RepaymentCurrency, payment.Currency) {
			totalPaidInRepayment += payment.Amount
		} else if strings.EqualFold(*debt.RepaymentCurrency, debt.PrincipalCurrency) {
			totalPaidInRepayment += paymentInDebt
		}
	}
	debt.RemainingAmount = remaining
	debt.TotalPaid = totalPaid
	debt.PercentPaid = percentPaid
	debt.TotalPaidInRepaymentCurrency = totalPaidInRepayment
	debt.UpdatedAt = utils.NowUTC()
	r.debts[debt.ID] = debt
	return nil
}

func (r *InMemoryRepository) UpdateDebtPayment(ctx context.Context, payment *DebtPayment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	debtPayments := r.debtPayments[payment.DebtID]
	if debtPayments == nil {
		return appErrors.DebtPaymentNotFound
	}
	current, ok := debtPayments[payment.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.DebtPaymentNotFound
	}
	debt, ok := r.debts[payment.DebtID]
	if !ok || debt == nil || debt.DeletedAt != "" {
		return appErrors.DebtNotFound
	}
	accountID := payment.AccountID
	if accountID == nil || *accountID == "" {
		accountID = current.AccountID
	}
	var account *Account
	if accountID != nil && *accountID != "" {
		account = r.accounts[*accountID]
	}
	if payment.Currency == "" {
		payment.Currency = current.Currency
	}
	if payment.Currency != "" && account != nil && payment.Currency != account.Currency {
		return appErrors.InvalidFinanceData
	}
	amountDelta := payment.Amount - current.Amount
	if amountDelta != 0 && account != nil {
		adjustment := amountDelta
		switch debt.Direction {
		case "they_owe_me":
			adjustment = amountDelta
		case "i_owe":
			adjustment = -amountDelta
		default:
			return appErrors.InvalidFinanceData
		}
		account.CurrentBalance += adjustment
		r.accounts[account.ID] = account
	}

	deltaDebt := payment.ConvertedAmountToDebt - current.ConvertedAmountToDebt
	if deltaDebt != 0 {
		remaining := debt.RemainingAmount
		if remaining <= 0 {
			remaining = debt.PrincipalAmount
		}
		remaining -= deltaDebt
		if remaining < 0 {
			remaining = 0
		}
		totalPaid := debt.TotalPaid + deltaDebt
		if totalPaid < 0 {
			totalPaid = 0
		}
		if debt.PrincipalAmount > 0 && totalPaid > debt.PrincipalAmount {
			totalPaid = debt.PrincipalAmount
		}
		percentPaid := 0.0
		if debt.PrincipalAmount > 0 {
			percentPaid = (totalPaid / debt.PrincipalAmount) * 100
			if percentPaid > 100 {
				percentPaid = 100
			}
		}
		totalPaidInRepayment := debt.TotalPaidInRepaymentCurrency
		if debt.RepaymentCurrency != nil {
			if strings.EqualFold(*debt.RepaymentCurrency, payment.Currency) {
				totalPaidInRepayment += amountDelta
			} else if strings.EqualFold(*debt.RepaymentCurrency, debt.PrincipalCurrency) {
				totalPaidInRepayment += deltaDebt
			}
		}
		debt.RemainingAmount = remaining
		debt.TotalPaid = totalPaid
		debt.PercentPaid = percentPaid
		debt.TotalPaidInRepaymentCurrency = totalPaidInRepayment
		debt.UpdatedAt = utils.NowUTC()
		r.debts[debt.ID] = debt
	}
	payment.CreatedAt = current.CreatedAt
	payment.UpdatedAt = utils.NowUTC()
	debtPayments[payment.ID] = cloneDebtPayment(payment)
	return nil
}

func (r *InMemoryRepository) DeleteDebtPayment(ctx context.Context, debtID, paymentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	debtPayments := r.debtPayments[debtID]
	if debtPayments == nil {
		return appErrors.DebtPaymentNotFound
	}
	payment, ok := debtPayments[paymentID]
	if !ok || payment == nil || payment.DeletedAt != "" {
		return appErrors.DebtPaymentNotFound
	}
	debt, ok := r.debts[debtID]
	if !ok || debt == nil || debt.DeletedAt != "" {
		return appErrors.DebtNotFound
	}
	if payment.AccountID != nil {
		account, ok := r.accounts[*payment.AccountID]
		if ok && account != nil && account.DeletedAt == "" {
			adjustment := payment.Amount
			switch debt.Direction {
			case "they_owe_me":
				adjustment = -payment.Amount
			case "i_owe":
				adjustment = payment.Amount
			default:
				return appErrors.InvalidFinanceData
			}
			account.CurrentBalance += adjustment
			r.accounts[account.ID] = account
		}
	}
	deltaDebt := payment.ConvertedAmountToDebt
	remaining := debt.RemainingAmount
	if remaining <= 0 {
		remaining = debt.PrincipalAmount
	}
	remaining += deltaDebt
	if remaining > debt.PrincipalAmount {
		remaining = debt.PrincipalAmount
	}
	totalPaid := debt.TotalPaid - deltaDebt
	if totalPaid < 0 {
		totalPaid = 0
	}
	percentPaid := 0.0
	if debt.PrincipalAmount > 0 {
		percentPaid = (totalPaid / debt.PrincipalAmount) * 100
		if percentPaid > 100 {
			percentPaid = 100
		}
	}
	totalPaidInRepayment := debt.TotalPaidInRepaymentCurrency
	if debt.RepaymentCurrency != nil {
		if strings.EqualFold(*debt.RepaymentCurrency, payment.Currency) {
			totalPaidInRepayment -= payment.Amount
		} else if strings.EqualFold(*debt.RepaymentCurrency, debt.PrincipalCurrency) {
			totalPaidInRepayment -= deltaDebt
		}
		if totalPaidInRepayment < 0 {
			totalPaidInRepayment = 0
		}
	}
	debt.RemainingAmount = remaining
	debt.TotalPaid = totalPaid
	debt.PercentPaid = percentPaid
	debt.TotalPaidInRepaymentCurrency = totalPaidInRepayment
	debt.UpdatedAt = utils.NowUTC()
	r.debts[debt.ID] = debt
	payment.DeletedAt = utils.NowUTC()
	payment.UpdatedAt = payment.DeletedAt
	debtPayments[paymentID] = payment
	return nil
}

func (r *InMemoryRepository) ListCounterparties(ctx context.Context) ([]*Counterparty, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Counterparty, 0, len(r.counterparties))
	for _, counterparty := range r.counterparties {
		if counterparty == nil || counterparty.DeletedAt != "" {
			continue
		}
		results = append(results, cloneCounterparty(counterparty))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetCounterpartyByID(ctx context.Context, id string) (*Counterparty, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	counterparty, ok := r.counterparties[id]
	if !ok || counterparty == nil || counterparty.DeletedAt != "" {
		return nil, appErrors.CounterpartyNotFound
	}
	return cloneCounterparty(counterparty), nil
}

func (r *InMemoryRepository) CreateCounterparty(ctx context.Context, counterparty *Counterparty) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if counterparty.ID == "" {
		counterparty.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	counterparty.CreatedAt = now
	counterparty.UpdatedAt = now
	r.counterparties[counterparty.ID] = cloneCounterparty(counterparty)
	return nil
}

func (r *InMemoryRepository) UpdateCounterparty(ctx context.Context, counterparty *Counterparty) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.counterparties[counterparty.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.CounterpartyNotFound
	}
	counterparty.CreatedAt = current.CreatedAt
	counterparty.UpdatedAt = utils.NowUTC()
	r.counterparties[counterparty.ID] = cloneCounterparty(counterparty)
	return nil
}

func (r *InMemoryRepository) DeleteCounterparty(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	counterparty, ok := r.counterparties[id]
	if !ok || counterparty == nil || counterparty.DeletedAt != "" {
		return appErrors.CounterpartyNotFound
	}
	counterparty.DeletedAt = utils.NowUTC()
	counterparty.UpdatedAt = counterparty.DeletedAt
	r.counterparties[id] = counterparty
	return nil
}

func (r *InMemoryRepository) ListFXRates(ctx context.Context) ([]*FXRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*FXRate, 0, len(r.fxRates))
	for _, rate := range r.fxRates {
		if rate == nil {
			continue
		}
		results = append(results, cloneFXRate(rate))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetFXRateByID(ctx context.Context, id string) (*FXRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rate, ok := r.fxRates[id]
	if !ok || rate == nil {
		return nil, appErrors.FXRateNotFound
	}
	return cloneFXRate(rate), nil
}

func (r *InMemoryRepository) CreateFXRate(ctx context.Context, rate *FXRate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	rate.CreatedAt = now
	rate.UpdatedAt = now
	r.fxRates[rate.ID] = cloneFXRate(rate)
	return nil
}

func (r *InMemoryRepository) ListCategories(ctx context.Context, categoryType string, activeOnly bool) ([]*FinanceCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*FinanceCategory, 0, len(r.categories))
	for _, category := range r.categories {
		if category == nil {
			continue
		}
		if categoryType != "" && category.Type != categoryType {
			continue
		}
		if activeOnly && !category.IsActive {
			continue
		}
		copy := *category
		results = append(results, &copy)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].SortOrder == results[j].SortOrder {
			return strings.Compare(results[i].ID, results[j].ID) < 0
		}
		return results[i].SortOrder < results[j].SortOrder
	})
	return results, nil
}

func (r *InMemoryRepository) CreateCategory(ctx context.Context, category *FinanceCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if category.ID == "" {
		category.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	category.CreatedAt = now
	category.UpdatedAt = now
	copy := *category
	r.categories[category.ID] = &copy
	return nil
}

func (r *InMemoryRepository) UpdateCategory(ctx context.Context, category *FinanceCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.categories[category.ID]; !ok {
		return appErrors.InvalidFinanceData
	}
	category.UpdatedAt = utils.NowUTC()
	copy := *category
	r.categories[category.ID] = &copy
	return nil
}

func (r *InMemoryRepository) ListQuickExpenseCategories(ctx context.Context, categoryType string) ([]*QuickExpenseCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userID, _ := ctx.Value("user_id").(string)
	key := quickExpenseKey(userID, categoryType)
	saved := r.quickExp[key]
	results := make([]*QuickExpenseCategory, 0, len(saved))
	for _, entry := range saved {
		if entry == nil {
			continue
		}
		copy := *entry
		results = append(results, &copy)
	}
	return results, nil
}

func (r *InMemoryRepository) ReplaceQuickExpenseCategories(ctx context.Context, categoryType string, categories []*QuickExpenseCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	userID, _ := ctx.Value("user_id").(string)
	key := quickExpenseKey(userID, categoryType)
	next := make([]*QuickExpenseCategory, 0, len(categories))
	for _, entry := range categories {
		if entry == nil {
			continue
		}
		copy := *entry
		next = append(next, &copy)
	}
	r.quickExp[key] = next
	return nil
}

func quickExpenseKey(userID, categoryType string) string {
	return strings.TrimSpace(userID) + ":" + strings.ToLower(strings.TrimSpace(categoryType))
}

func cloneAccount(account *Account) *Account {
	if account == nil {
		return nil
	}
	copy := *account
	return &copy
}

func cloneTransaction(txn *Transaction) *Transaction {
	if txn == nil {
		return nil
	}
	copy := *txn
	return &copy
}

func cloneBudget(budget *Budget) *Budget {
	if budget == nil {
		return nil
	}
	copy := *budget
	return &copy
}

func cloneDebt(debt *Debt) *Debt {
	if debt == nil {
		return nil
	}
	copy := *debt
	return &copy
}

func cloneDebtPayment(payment *DebtPayment) *DebtPayment {
	if payment == nil {
		return nil
	}
	copy := *payment
	return &copy
}

func cloneCounterparty(counterparty *Counterparty) *Counterparty {
	if counterparty == nil {
		return nil
	}
	copy := *counterparty
	return &copy
}

func cloneFXRate(rate *FXRate) *FXRate {
	if rate == nil {
		return nil
	}
	copy := *rate
	return &copy
}
