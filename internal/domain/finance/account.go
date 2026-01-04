package finance

import "github.com/leora/leora-server/internal/domain/common"

// AccountType enumerates different account categories.
type AccountType string

const (
    AccountTypeCash       AccountType = "cash"
    AccountTypeCard       AccountType = "card"
    AccountTypeSavings    AccountType = "savings"
    AccountTypeInvestment AccountType = "investment"
    AccountTypeCredit     AccountType = "credit"
    AccountTypeDebt       AccountType = "debt"
    AccountTypeOther      AccountType = "other"
)

// Account represents a financial bucket.
type Account struct {
    common.BaseEntity
    Name           string       `json:"name"`
    Type           AccountType  `json:"accountType"`
    Currency       string       `json:"currency"`
    InitialBalance float64      `json:"initialBalance"`
    CurrentBalance float64      `json:"currentBalance"`
    LinkedGoalID   string       `json:"linkedGoalId,omitempty"`
    CustomTypeID   string       `json:"customTypeId,omitempty"`
    Icon           string       `json:"icon,omitempty"`
    Color          string       `json:"color,omitempty"`
}
