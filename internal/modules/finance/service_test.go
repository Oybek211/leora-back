package finance

import (
	"context"
	"testing"
)

func TestRepayDebtUpdatesTotalsSameCurrency(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	repo := NewInMemoryRepository()
	service := NewService(repo, nil)

	account := &Account{
		Name:           "Cash",
		AccountType:    "cash",
		Currency:       "USD",
		InitialBalance: 1000,
		CurrentBalance: 1000,
		ShowStatus:     "active",
	}
	createdAccount, _, err := service.CreateAccount(ctx, account)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	debt := &Debt{
		Name:              "Loan",
		Direction:         "i_owe",
		PrincipalAmount:   100,
		PrincipalCurrency: "USD",
		BaseCurrency:      "USD",
		ShowStatus:        "active",
	}
	if _, err := service.CreateDebt(ctx, debt); err != nil {
		t.Fatalf("create debt: %v", err)
	}

	result, err := service.RepayDebt(ctx, debt.ID, DebtValueInput{
		AccountID:      createdAccount.ID,
		Amount:         30,
		AmountCurrency: "USD",
	})
	if err != nil {
		t.Fatalf("repay debt: %v", err)
	}
	if result.Debt.RemainingAmount != 70 {
		t.Fatalf("remaining amount mismatch: got %.2f, want 70.00", result.Debt.RemainingAmount)
	}
	if result.Debt.TotalPaid != 30 {
		t.Fatalf("total paid mismatch: got %.2f, want 30.00", result.Debt.TotalPaid)
	}

	updatedAccount, err := service.GetAccount(ctx, createdAccount.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updatedAccount.CurrentBalance != 970 {
		t.Fatalf("account balance mismatch: got %.2f, want 970.00", updatedAccount.CurrentBalance)
	}
}

func TestRepayDebtConvertsDebtCurrencyToAccountCurrency(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-2")
	repo := NewInMemoryRepository()
	service := NewService(repo, nil)

	account := &Account{
		Name:           "UZS",
		AccountType:    "cash",
		Currency:       "UZS",
		InitialBalance: 1_000_000,
		CurrentBalance: 1_000_000,
		ShowStatus:     "active",
	}
	createdAccount, _, err := service.CreateAccount(ctx, account)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	debt := &Debt{
		Name:              "USD Debt",
		Direction:         "i_owe",
		PrincipalAmount:   100,
		PrincipalCurrency: "USD",
		BaseCurrency:      "USD",
		ShowStatus:        "active",
	}
	if _, err := service.CreateDebt(ctx, debt); err != nil {
		t.Fatalf("create debt: %v", err)
	}

	_, err = service.CreateFXRate(ctx, &FXRate{
		FromCurrency: "USD",
		ToCurrency:   "UZS",
		Rate:         12_000,
		Date:         "2026-01-01",
	})
	if err != nil {
		t.Fatalf("create fx rate: %v", err)
	}

	_, err = service.RepayDebt(ctx, debt.ID, DebtValueInput{
		AccountID:      createdAccount.ID,
		Amount:         10,
		AmountCurrency: "USD",
		Date:           stringPtr("2026-01-01"),
	})
	if err != nil {
		t.Fatalf("repay debt: %v", err)
	}

	updatedAccount, err := service.GetAccount(ctx, createdAccount.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if updatedAccount.CurrentBalance != 880_000 {
		t.Fatalf("account balance mismatch: got %.2f, want 880000.00", updatedAccount.CurrentBalance)
	}
}

func TestFinanceSummaryUpdatesAfterTransaction(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-3")
	repo := NewInMemoryRepository()
	service := NewService(repo, nil)

	account := &Account{
		Name:           "Cash",
		AccountType:    "cash",
		Currency:       "USD",
		InitialBalance: 0,
		CurrentBalance: 0,
		ShowStatus:     "active",
	}
	createdAccount, _, err := service.CreateAccount(ctx, account)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = service.CreateTransaction(ctx, &Transaction{
		Type:      TransactionTypeIncome,
		AccountID: &createdAccount.ID,
		Amount:    120,
		Currency:  "USD",
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	summary, err := service.FinanceSummary(ctx, "", "", "USD", nil)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Totals.Income != 120 {
		t.Fatalf("summary income mismatch: got %.2f, want 120.00", summary.Totals.Income)
	}
	if summary.Totals.Balance != 120 {
		t.Fatalf("summary balance mismatch: got %.2f, want 120.00", summary.Totals.Balance)
	}
}

func stringPtr(value string) *string {
	return &value
}
