package focus

import "github.com/leora/leora-server/internal/domain/common"

// FocusStatus enumerates focus session states.
type FocusStatus string

const (
    FocusStatusInProgress FocusStatus = "in_progress"
    FocusStatusCompleted  FocusStatus = "completed"
    FocusStatusCanceled   FocusStatus = "canceled"
    FocusStatusPaused     FocusStatus = "paused"
)

// FocusSession models a single Pomodoro-style session.
type FocusSession struct {
    common.BaseEntity
    TaskID            string      `json:"taskId,omitempty"`
    GoalID            string      `json:"goalId,omitempty"`
    PlannedMinutes    int         `json:"plannedMinutes"`
    ActualMinutes     int         `json:"actualMinutes"`
    Status            FocusStatus `json:"status"`
    StartedAt         string      `json:"startedAt"`
    EndedAt           string      `json:"endedAt,omitempty"`
    InterruptionsCount int        `json:"interruptionsCount"`
    Notes             string      `json:"notes,omitempty"`
}
