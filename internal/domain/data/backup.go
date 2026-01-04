package data

// Backup represents a stored data snapshot.
type Backup struct {
    ID              string   `json:"id"`
    UserID          string   `json:"userId"`
    Type            string   `json:"type"`
    Status          string   `json:"status"`
    Storage         string   `json:"storage"`
    FileURL         string   `json:"fileUrl,omitempty"`
    EntitiesIncluded []string `json:"entitiesIncluded"`
    EntityCounts    map[string]int `json:"entityCounts,omitempty"`
    CreatedAt       string   `json:"createdAt"`
}

// Export represents a data export request.
type Export struct {
    ID        string   `json:"id"`
    UserID    string   `json:"userId"`
    Format    string   `json:"format"`
    Scope     string   `json:"scope"`
    Status    string   `json:"status"`
    FileURL   string   `json:"fileUrl,omitempty"`
    CreatedAt string   `json:"createdAt"`
}
