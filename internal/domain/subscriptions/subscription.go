package subscriptions

// Plan represents a subscription offering.
type Plan struct {
    ID       string            `json:"id"`
    Name     string            `json:"name"`
    Tier     string            `json:"tier"`
    Interval string            `json:"interval"`
    Prices   map[string]float64 `json:"prices"`
    Discount int               `json:"discount"`
    Popular  bool              `json:"isPopular"`
}

// Subscription captures a user's active subscription.
type Subscription struct {
    ID               string `json:"id"`
    Tier             string `json:"tier"`
    Status           string `json:"status"`
    CurrentPeriodEnd string `json:"currentPeriodEnd"`
    CancelAtPeriodEnd bool  `json:"cancelAtPeriodEnd"`
}

// SubscriptionFeature indicates limits for a feature.
type SubscriptionFeature struct {
    Limit int `json:"limit"`
    Used  int `json:"used"`
}
