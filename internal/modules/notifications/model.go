package notifications

// Notification represents a pushable alert.
type Notification struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	DeletedAt string `json:"deletedAt,omitempty"`
}
