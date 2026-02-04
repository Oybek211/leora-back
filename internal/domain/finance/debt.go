package finance

import "github.com/leora/leora-server/internal/domain/common"

// Debt captures owed amount to a counterparty.
type Debt struct {
    common.BaseEntity
    Name           string  `json:"name"`
    Balance        float64 `json:"balance"`
    Interest       float64 `json:"interest"`
    Currency       string  `json:"currency"`
    DueDate        string  `json:"dueDate,omitempty"`
    MinimumPayment float64 `json:"minimumPayment"`
    LinkedGoalID   *string `json:"linkedGoalId,omitempty" db:"linked_goal_id"`
    LinkedBudgetID *string `json:"linkedBudgetId,omitempty" db:"linked_budget_id"`
}
