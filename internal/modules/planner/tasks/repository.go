package tasks

import "context"

// ListOptions holds filtering and sorting options for task queries
type ListOptions struct {
	Status       string
	ShowStatus   string
	Priority     string
	GoalID       string
	HabitID      string
	DueDate      string
	DueDateFrom  string
	DueDateTo    string
	Search       string
	SortBy       string
	SortOrder    string
}

// Repository defines the interface for task persistence
type Repository interface {
	List(ctx context.Context, opts ListOptions) ([]*Task, error)
	GetByID(ctx context.Context, id string) (*Task, error)
	Create(ctx context.Context, task *Task) error
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	UpdateChecklistItem(ctx context.Context, taskID, itemID string, completed bool) error
}
