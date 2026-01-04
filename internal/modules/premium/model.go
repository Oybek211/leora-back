package premium

// Plan represents subscription offering.
type Plan struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Interval string             `json:"interval"`
	Prices   map[string]float64 `json:"prices"`
	CreatedAt string            `json:"createdAt,omitempty"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
	DeletedAt string            `json:"deletedAt,omitempty"`
}

// Subscription captures a user's tier.
type Subscription struct {
	ID                string `json:"id"`
	Tier              string `json:"tier"`
	Status            string `json:"status"`
	CancelAtPeriodEnd bool   `json:"cancelAtPeriodEnd"`
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
	DeletedAt         string `json:"deletedAt,omitempty"`
}
