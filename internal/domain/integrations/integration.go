package integrations

// IntegrationStatus indicates connectivity state.
type IntegrationStatus string

const (
    IntegrationStatusConnected    IntegrationStatus = "connected"
    IntegrationStatusDisconnected IntegrationStatus = "disconnected"
    IntegrationStatusError        IntegrationStatus = "error"
)

// Integration represents an OAuth connection.
type Integration struct {
    ID          string             `json:"id"`
    UserID      string             `json:"userId"`
    Provider    string             `json:"provider"`
    Category    string             `json:"category"`
    Status      IntegrationStatus  `json:"status"`
    AccountName string             `json:"accountName,omitempty"`
    LastSyncAt  string             `json:"lastSyncAt,omitempty"`
}

// SyncLog tracks manual sync attempts.
type SyncLog struct {
    ID             string `json:"id"`
    IntegrationID  string `json:"integrationId"`
    Direction      string `json:"direction"`
    Status         string `json:"status"`
    ItemsSynced    int    `json:"itemsSynced"`
    StartedAt      string `json:"startedAt"`
    CompletedAt    string `json:"completedAt,omitempty"`
}
