package finance

// Account represents a financial account.
type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
	Type     string `json:"accountType"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	DeletedAt string `json:"deletedAt,omitempty"`
}

// Transaction models a ledger entry.
type Transaction struct {
	ID        string  `json:"id"`
	AccountID string  `json:"accountId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Category  string  `json:"category"`
	CreatedAt string  `json:"createdAt,omitempty"`
	UpdatedAt string  `json:"updatedAt,omitempty"`
	DeletedAt string  `json:"deletedAt,omitempty"`
}

// Budget tracks spending goals.
type Budget struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Currency string  `json:"currency"`
	Limit    float64 `json:"limit"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	DeletedAt string `json:"deletedAt,omitempty"`
}

// Debt represents an owed balance.
type Debt struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	DeletedAt string `json:"deletedAt,omitempty"`
}
