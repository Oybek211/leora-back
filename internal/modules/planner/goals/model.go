package goals

// Goal represents high-level goal metadata with all fields from BACKEND_API.md
type Goal struct {
	ID                  string       `json:"id"`
	Title               string       `json:"title"`
	Description         *string      `json:"description,omitempty"`
	GoalType            string       `json:"goalType"`
	Status              string       `json:"status"`
	ShowStatus          string       `json:"showStatus"`
	MetricType          string       `json:"metricType"`
	Direction           string       `json:"direction"`
	Unit                *string      `json:"unit,omitempty"`
	InitialValue        *float64     `json:"initialValue,omitempty"`
	TargetValue         *float64     `json:"targetValue,omitempty"`
	ProgressTargetValue *float64     `json:"progressTargetValue,omitempty"`
	CurrentValue        float64      `json:"currentValue"`
	FinanceMode         *string      `json:"financeMode,omitempty"`
	Currency            *string      `json:"currency,omitempty"`
	LinkedBudgetID      *string      `json:"linkedBudgetId,omitempty"`
	LinkedDebtID        *string      `json:"linkedDebtId,omitempty"`
	StartDate           *string      `json:"startDate,omitempty"`
	TargetDate          *string      `json:"targetDate,omitempty"`
	CompletedDate       *string      `json:"completedDate,omitempty"`
	ProgressPercent     float64      `json:"progressPercent"`
	Milestones          []*Milestone `json:"milestones,omitempty"`
	Stats               *GoalStats   `json:"stats,omitempty"`
	CreatedAt           string       `json:"createdAt,omitempty"`
	UpdatedAt           string       `json:"updatedAt,omitempty"`
	DeletedAt           *string      `json:"deletedAt,omitempty"`
}

// Milestone represents a goal milestone
type Milestone struct {
	ID            string  `json:"id"`
	GoalID        string  `json:"goalId"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	TargetPercent float64 `json:"targetPercent"`
	DueDate       *string `json:"dueDate,omitempty"`
	CompletedAt   *string `json:"completedAt,omitempty"`
	Order         int     `json:"order"`
	CreatedAt     string  `json:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
}

// CheckIn represents a goal progress check-in
type CheckIn struct {
	ID         string  `json:"id"`
	GoalID     string  `json:"goalId"`
	Value      float64 `json:"value"`
	Note       *string `json:"note,omitempty"`
	SourceType string  `json:"sourceType"`
	SourceID   *string `json:"sourceId,omitempty"`
	DateKey    *string `json:"dateKey,omitempty"`
	CreatedAt  string  `json:"createdAt,omitempty"`
}

// GoalStats represents goal statistics
type GoalStats struct {
	TotalTasks     int `json:"totalTasks" db:"total_tasks"`
	CompletedTasks int `json:"completedTasks" db:"completed_tasks"`
	TotalHabits    int `json:"totalHabits" db:"total_habits"`
	FocusMinutes   int `json:"focusMinutes" db:"focus_minutes"`
}
