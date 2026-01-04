package finance

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository defines finance persistence.
type Repository interface {
	ListAccounts(ctx context.Context) ([]*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	CreateAccount(ctx context.Context, account *Account) error
	UpdateAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id string) error

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
}

// InMemoryRepository stores finance data in memory.
type InMemoryRepository struct {
	mu           sync.RWMutex
	accounts     map[string]*Account
	transactions map[string]*Transaction
	budgets      map[string]*Budget
	debts        map[string]*Debt
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		accounts:     make(map[string]*Account),
		transactions: make(map[string]*Transaction),
		budgets:      make(map[string]*Budget),
		debts:        make(map[string]*Debt),
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

func (r *InMemoryRepository) CreateAccount(ctx context.Context, account *Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if account.ID == "" {
		account.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	account.CreatedAt = now
	account.UpdatedAt = now
	r.accounts[account.ID] = cloneAccount(account)
	return nil
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

func (r *InMemoryRepository) DeleteAccount(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	account, ok := r.accounts[id]
	if !ok || account == nil || account.DeletedAt != "" {
		return appErrors.AccountNotFound
	}
	account.DeletedAt = utils.NowUTC()
	account.UpdatedAt = account.DeletedAt
	r.accounts[id] = account
	return nil
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
	if txn.ID == "" {
		txn.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	txn.CreatedAt = now
	txn.UpdatedAt = now
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
