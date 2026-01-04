package tasks

import "github.com/leora/leora-server/internal/domain/common"

// TaskStatus enumerates possible progress states of a task.
type TaskStatus string

const (
    TaskStatusInbox      TaskStatus = "inbox"
    TaskStatusPlanned    TaskStatus = "planned"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusCanceled   TaskStatus = "canceled"
    TaskStatusMoved      TaskStatus = "moved"
    TaskStatusOverdue    TaskStatus = "overdue"
)

// TaskPriority captures priority levels.
type TaskPriority string

const (
    TaskPriorityLow    TaskPriority = "low"
    TaskPriorityMedium TaskPriority = "medium"
    TaskPriorityHigh   TaskPriority = "high"
)

// FinanceLink enumerates task-finance relationships.
type FinanceLink string

const (
    FinanceLinkRecordExpenses FinanceLink = "record_expenses"
    FinanceLinkPayDebt        FinanceLink = "pay_debt"
    FinanceLinkReviewBudget   FinanceLink = "review_budget"
    FinanceLinkTransferMoney  FinanceLink = "transfer_money"
    FinanceLinkNone           FinanceLink = "none"
)

// Task describes a planner task with contextual metadata.
type Task struct {
    common.BaseEntity
    Title              string       `json:"title"`
    Status             TaskStatus   `json:"status"`
    Priority           TaskPriority `json:"priority"`
    GoalID             string       `json:"goalId,omitempty"`
    HabitID            string       `json:"habitId,omitempty"`
    FinanceLink        FinanceLink  `json:"financeLink,omitempty"`
    ProgressValue      float64      `json:"progressValue,omitempty"`
    ProgressUnit       string       `json:"progressUnit,omitempty"`
    DueDate            string       `json:"dueDate,omitempty"`
    StartDate          string       `json:"startDate,omitempty"`
    TimeOfDay          string       `json:"timeOfDay,omitempty"`
    EstimatedMinutes   int          `json:"estimatedMinutes,omitempty"`
    EnergyLevel        int          `json:"energyLevel,omitempty"`
    Context            string       `json:"context,omitempty"`
    Notes              string       `json:"notes,omitempty"`
    LastFocusSessionID string       `json:"lastFocusSessionId,omitempty"`
    FocusTotalMinutes  int          `json:"focusTotalMinutes"`
}

// ChecklistItem captures a subtask item tied to a task.
type ChecklistItem struct {
    ID        string `json:"id"`
    TaskID    string `json:"taskId"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
    Order     int    `json:"order"`
}

// TaskDependency describes prerequisites between tasks.
type TaskDependency struct {
    ID              string      `json:"id"`
    TaskID          string      `json:"taskId"`
    DependsOnTaskID string      `json:"dependsOnTaskId"`
    Status          string      `json:"status"`
}
