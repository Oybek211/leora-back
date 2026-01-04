package goals

import "github.com/leora/leora-server/internal/domain/common"

// GoalType enumerates categories of goals.
type GoalType string

const (
    GoalTypeFinancial    GoalType = "financial"
    GoalTypeHealth       GoalType = "health"
    GoalTypeEducation    GoalType = "education"
    GoalTypeProductivity GoalType = "productivity"
    GoalTypePersonal     GoalType = "personal"
)

// MetricType describes how a goal is measured.
type MetricType string

const (
    MetricTypeNone     MetricType = "none"
    MetricTypeAmount   MetricType = "amount"
    MetricTypeWeight   MetricType = "weight"
    MetricTypeCount    MetricType = "count"
    MetricTypeDuration MetricType = "duration"
    MetricTypeCustom   MetricType = "custom"
)

// Direction indicates whether goal should increase or decrease.
type Direction string

const (
    DirectionIncrease Direction = "increase"
    DirectionDecrease Direction = "decrease"
    DirectionNeutral  Direction = "neutral"
)

// Goal captures long-term objectives.
type Goal struct {
    common.BaseEntity
    Title                string       `json:"title"`
    Description          string       `json:"description,omitempty"`
    Type                 GoalType     `json:"goalType"`
    Status               string       `json:"status"`
    MetricType           MetricType   `json:"metricType"`
    Direction            Direction    `json:"direction"`
    Unit                 string       `json:"unit,omitempty"`
    InitialValue         float64      `json:"initialValue,omitempty"`
    TargetValue          float64      `json:"targetValue,omitempty"`
    ProgressTargetValue  float64      `json:"progressTargetValue,omitempty"`
    CurrentValue         float64      `json:"currentValue"`
    FinanceMode          string       `json:"financeMode,omitempty"`
    Currency             string       `json:"currency,omitempty"`
    LinkedBudgetID       string       `json:"linkedBudgetId,omitempty"`
    LinkedDebtID         string       `json:"linkedDebtId,omitempty"`
    StartDate            string       `json:"startDate,omitempty"`
    TargetDate           string       `json:"targetDate,omitempty"`
    CompletedDate        string       `json:"completedDate,omitempty"`
    ProgressPercent      float64      `json:"progressPercent"`
}

// GoalStats captures aggregated progress for a goal.
type GoalStats struct {
    FinancialProgressPercent float64 `json:"financialProgressPercent"`
    HabitsProgressPercent    float64 `json:"habitsProgressPercent"`
    TasksProgressPercent     float64 `json:"tasksProgressPercent"`
    FocusMinutesLast30       int     `json:"focusMinutesLast30"`
}

// GoalMilestone describes a checkpoint for a goal.
type GoalMilestone struct {
    ID            string  `json:"id"`
    GoalID        string  `json:"goalId"`
    Title         string  `json:"title"`
    Description   string  `json:"description,omitempty"`
    TargetPercent float64 `json:"targetPercent"`
    DueDate       string  `json:"dueDate,omitempty"`
    CompletedAt   string  `json:"completedAt,omitempty"`
    Order         int     `json:"order"`
}

// GoalCheckIn captures progress added to a goal.
type GoalCheckIn struct {
    ID         string  `json:"id"`
    GoalID     string  `json:"goalId"`
    Value      float64 `json:"value"`
    Note       string  `json:"note,omitempty"`
    SourceType string  `json:"sourceType"`
    SourceID   string  `json:"sourceId,omitempty"`
    DateKey    string  `json:"dateKey,omitempty"`
    CreatedAt  string  `json:"createdAt"`
}
