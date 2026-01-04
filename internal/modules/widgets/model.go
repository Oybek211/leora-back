package widgets

// Widget represents a home screen card.
type Widget struct {
	ID     string                 `json:"id"`
	Title  string                 `json:"title"`
	Config map[string]interface{} `json:"config"`
	CreatedAt string              `json:"createdAt,omitempty"`
	UpdatedAt string              `json:"updatedAt,omitempty"`
	DeletedAt string              `json:"deletedAt,omitempty"`
}
