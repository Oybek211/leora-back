package finance

import "github.com/leora/leora-server/internal/domain/common"

// TransactionType enumerates inflows/outflows.
type TransactionType string

const (
    TransactionTypeIncome  TransactionType = "income"
    TransactionTypeExpense TransactionType = "expense"
    TransactionTypeTransfer TransactionType = "transfer"
)

// Transaction records cash movement tied to accounts.
type Transaction struct {
    common.BaseEntity
    AccountID     string          `json:"accountId"`
    LinkedGoalID  string          `json:"linkedGoalId,omitempty"`
    LinkedDebtID  string          `json:"linkedDebtId,omitempty"`
    CounterpartyID string         `json:"counterpartyId,omitempty"`
    Amount        float64         `json:"amount"`
    Currency      string          `json:"currency"`
    Category      string          `json:"category"`
    Description   string          `json:"description,omitempty"`
    TransactionType TransactionType `json:"transactionType"`
    Date          string          `json:"date"`
}
