package ai

// AIUsage tracks a single AI request.
type AIUsage struct {
    ID           string `json:"id"`
    UserID       string `json:"userId"`
    Channel      string `json:"channel"`
    RequestType  string `json:"requestType"`
    TokensUsed   int    `json:"tokensUsed"`
    ResponseTime int    `json:"responseTime"`
    Success      bool   `json:"success"`
    ErrorCode    string `json:"errorCode,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt    string `json:"createdAt"`
}

// AIQuota holds usage limits for a channel.
type AIQuota struct {
    ID         string `json:"id"`
    UserID     string `json:"userId"`
    Channel    string `json:"channel"`
    PeriodStart string `json:"periodStart"`
    PeriodEnd   string `json:"periodEnd"`
    Limit       int    `json:"limit"`
    Used        int    `json:"used"`
    CreatedAt   string `json:"createdAt"`
    UpdatedAt   string `json:"updatedAt"`
}
