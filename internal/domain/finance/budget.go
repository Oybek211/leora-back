package finance

import "github.com/leora/leora-server/internal/domain/common"

// Budget represents a spending limit.
type Budget struct {
    common.BaseEntity
    Name        string  `json:"name"`
    Currency    string  `json:"currency"`
    Limit       float64 `json:"limit"`
    Spent       float64 `json:"spent"`
    ResetDay    int     `json:"resetDay"`
    AccountIDs  []string `json:"accountIds"`
}
