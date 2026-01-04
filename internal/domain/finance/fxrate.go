package finance

// FXRate stores exchange rate metadata.
type FXRate struct {
    ID        string  `json:"id"`
    Base      string  `json:"base"`
    Target    string  `json:"target"`
    Rate      float64 `json:"rate"`
    UpdatedAt string  `json:"updatedAt"`
}
