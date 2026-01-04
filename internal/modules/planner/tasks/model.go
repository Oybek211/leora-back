package tasks

// Task represents a user's task with all fields from BACKEND_API.md
type Task struct {
	ID                  string              `json:"id"`
	Title               string              `json:"title"`
	Status              string              `json:"status"`
	ShowStatus          string              `json:"showStatus"`
	Priority            string              `json:"priority"`
	GoalID              *string             `json:"goalId,omitempty"`
	HabitID             *string             `json:"habitId,omitempty"`
	FinanceLink         *string             `json:"financeLink,omitempty"`
	ProgressValue       *float64            `json:"progressValue,omitempty"`
	ProgressUnit        *string             `json:"progressUnit,omitempty"`
	DueDate             *string             `json:"dueDate,omitempty"`
	StartDate           *string             `json:"startDate,omitempty"`
	TimeOfDay           *string             `json:"timeOfDay,omitempty"`
	EstimatedMinutes    *int                `json:"estimatedMinutes,omitempty"`
	EnergyLevel         *int                `json:"energyLevel,omitempty"`
	Context             *string             `json:"context,omitempty"`
	Notes               *string             `json:"notes,omitempty"`
	LastFocusSessionID  *string             `json:"lastFocusSessionId,omitempty"`
	FocusTotalMinutes   int                 `json:"focusTotalMinutes"`
	Checklist           []*ChecklistItem    `json:"checklist,omitempty"`
	Dependencies        []*TaskDependency   `json:"dependencies,omitempty"`
	CreatedAt           string              `json:"createdAt,omitempty"`
	UpdatedAt           string              `json:"updatedAt,omitempty"`
	DeletedAt           *string             `json:"deletedAt,omitempty"`
}

// ChecklistItem represents a subtask within a task
type ChecklistItem struct {
	ID        string `json:"id"`
	TaskID    string `json:"taskId"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Order     int    `json:"order"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	ID              string `json:"id"`
	TaskID          string `json:"taskId"`
	DependsOnTaskID string `json:"dependsOnTaskId"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}
