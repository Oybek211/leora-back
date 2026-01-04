package focus

// Session models a focus session.
type Session struct {
	ID                string  `json:"id" db:"id"`
	TaskID            *string `json:"taskId,omitempty" db:"task_id"`
	GoalID            *string `json:"goalId,omitempty" db:"goal_id"`
	PlannedMinutes    int     `json:"plannedMinutes" db:"planned_minutes"`
	ActualMinutes     int     `json:"actualMinutes" db:"actual_minutes"`
	Status            string  `json:"status" db:"status"`
	StartedAt         *string `json:"startedAt,omitempty" db:"started_at"`
	EndedAt           *string `json:"endedAt,omitempty" db:"ended_at"`
	InterruptionsCount int    `json:"interruptionsCount" db:"interruptions_count"`
	Notes             *string `json:"notes,omitempty" db:"notes"`
	CreatedAt         string  `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt         string  `json:"updatedAt,omitempty" db:"updated_at"`
	DeletedAt         *string `json:"deletedAt,omitempty" db:"deleted_at"`
}

// SessionStats represents focus session statistics
type SessionStats struct {
	TotalSessions      int     `json:"totalSessions"`
	TotalMinutes       int     `json:"totalMinutes"`
	CompletedSessions  int     `json:"completedSessions"`
	AverageMinutes     float64 `json:"averageMinutes"`
	TotalInterruptions int     `json:"totalInterruptions"`
}
