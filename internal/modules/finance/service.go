package finance

import "context"

// Service orchestrates finance use cases.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Accounts(ctx context.Context) ([]*Account, error) {
	return s.repo.ListAccounts(ctx)
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *Service) CreateAccount(ctx context.Context, account *Account) (*Account, error) {
	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Service) UpdateAccount(ctx context.Context, id string, account *Account) (*Account, error) {
	account.ID = id
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Service) PatchAccount(ctx context.Context, id string, fields map[string]interface{}) (*Account, error) {
	current, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["name"].(string); ok {
		current.Name = v
	}
	if v, ok := fields["currency"].(string); ok {
		current.Currency = v
	}
	if v, ok := fields["accountType"].(string); ok {
		current.Type = v
	}
	if err := s.repo.UpdateAccount(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	return s.repo.DeleteAccount(ctx, id)
}

func (s *Service) Transactions(ctx context.Context) ([]*Transaction, error) {
	return s.repo.ListTransactions(ctx)
}

func (s *Service) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

func (s *Service) CreateTransaction(ctx context.Context, txn *Transaction) (*Transaction, error) {
	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, err
	}
	return txn, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, id string, txn *Transaction) (*Transaction, error) {
	txn.ID = id
	if err := s.repo.UpdateTransaction(ctx, txn); err != nil {
		return nil, err
	}
	return txn, nil
}

func (s *Service) PatchTransaction(ctx context.Context, id string, fields map[string]interface{}) (*Transaction, error) {
	current, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["accountId"].(string); ok {
		current.AccountID = v
	}
	if v, ok := fields["amount"].(float64); ok {
		current.Amount = v
	}
	if v, ok := fields["currency"].(string); ok {
		current.Currency = v
	}
	if v, ok := fields["category"].(string); ok {
		current.Category = v
	}
	if err := s.repo.UpdateTransaction(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteTransaction(ctx context.Context, id string) error {
	return s.repo.DeleteTransaction(ctx, id)
}

func (s *Service) Budgets(ctx context.Context) ([]*Budget, error) {
	return s.repo.ListBudgets(ctx)
}

func (s *Service) GetBudget(ctx context.Context, id string) (*Budget, error) {
	return s.repo.GetBudgetByID(ctx, id)
}

func (s *Service) CreateBudget(ctx context.Context, budget *Budget) (*Budget, error) {
	if err := s.repo.CreateBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *Service) UpdateBudget(ctx context.Context, id string, budget *Budget) (*Budget, error) {
	budget.ID = id
	if err := s.repo.UpdateBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *Service) PatchBudget(ctx context.Context, id string, fields map[string]interface{}) (*Budget, error) {
	current, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["name"].(string); ok {
		current.Name = v
	}
	if v, ok := fields["currency"].(string); ok {
		current.Currency = v
	}
	if v, ok := fields["limit"].(float64); ok {
		current.Limit = v
	}
	if err := s.repo.UpdateBudget(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteBudget(ctx context.Context, id string) error {
	return s.repo.DeleteBudget(ctx, id)
}

func (s *Service) Debts(ctx context.Context) ([]*Debt, error) {
	return s.repo.ListDebts(ctx)
}

func (s *Service) GetDebt(ctx context.Context, id string) (*Debt, error) {
	return s.repo.GetDebtByID(ctx, id)
}

func (s *Service) CreateDebt(ctx context.Context, debt *Debt) (*Debt, error) {
	if err := s.repo.CreateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

func (s *Service) UpdateDebt(ctx context.Context, id string, debt *Debt) (*Debt, error) {
	debt.ID = id
	if err := s.repo.UpdateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

func (s *Service) PatchDebt(ctx context.Context, id string, fields map[string]interface{}) (*Debt, error) {
	current, err := s.repo.GetDebtByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["name"].(string); ok {
		current.Name = v
	}
	if v, ok := fields["balance"].(float64); ok {
		current.Balance = v
	}
	if err := s.repo.UpdateDebt(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteDebt(ctx context.Context, id string) error {
	return s.repo.DeleteDebt(ctx, id)
}
