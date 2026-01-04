package sync

// SyncStatus denotes real-time sync progress.
type SyncStatus string

const (
    SyncStatusPending SyncStatus = "pending"
    SyncStatusRunning SyncStatus = "running"
    SyncStatusSuccess SyncStatus = "success"
    SyncStatusFailed  SyncStatus = "failed"
)

// SyncEvent captures a single sync job.
type SyncEvent struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    Status    SyncStatus `json:"status"`
    StartedAt string    `json:"startedAt"`
    CompletedAt string  `json:"completedAt,omitempty"`
}
