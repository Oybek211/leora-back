package finance

// Counterparty describes financial counterparties (banks, people).
type Counterparty struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    LastSynced  string `json:"lastSyncedAt,omitempty"`
}
