package common

// ShowStatus represents the lifecycle state of an entity.
type ShowStatus string

const (
    ShowStatusActive   ShowStatus = "active"
    ShowStatusArchived ShowStatus = "archived"
    ShowStatusDeleted  ShowStatus = "deleted"
)

// SyncStatus captures synchronization state with remote services.
type SyncStatus string

const (
    SyncStatusLocal   SyncStatus = "local"
    SyncStatusSynced  SyncStatus = "synced"
    SyncStatusPending SyncStatus = "pending"
    SyncStatusConflict SyncStatus = "conflict"
)

// BaseEntity groups common metadata for persisted records.
type BaseEntity struct {
    ID             string     `json:"id"`
    UserID         string     `json:"userId"`
    ShowStatus     ShowStatus `json:"showStatus"`
    SyncStatus     SyncStatus `json:"syncStatus"`
    IdempotencyKey string     `json:"idempotencyKey,omitempty"`
    CreatedAt      string     `json:"createdAt"`
    UpdatedAt      string     `json:"updatedAt"`
}
