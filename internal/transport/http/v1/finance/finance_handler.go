package finance

import (
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/leora/leora-server/internal/domain/common"
    domainFinance "github.com/leora/leora-server/internal/domain/finance"
    appErrors "github.com/leora/leora-server/internal/errors"
    "github.com/leora/leora-server/internal/transport/http/v1/response"
)

// Handler groups finance-related endpoints.
type Handler struct{}

// RegisterRoutes registers finance routes.
func RegisterRoutes(router fiber.Router) {
    h := &Handler{}
    accounts := router.Group("/accounts")
    accounts.Get("", h.listAccounts)
    accounts.Get(":id", h.getAccount)
    accounts.Post("", h.createAccount)
    accounts.Patch(":id", h.updateAccount)
    accounts.Delete(":id", h.deleteAccount)
    accounts.Patch(":id/adjust-balance", h.adjustBalance)
    accounts.Get("/summary", h.summary)

    transactions := router.Group("/transactions")
    transactions.Get("", h.listTransactions)
    transactions.Get(":id", h.getTransaction)
    transactions.Post("", h.createTransaction)
    transactions.Patch(":id", h.updateTransaction)
    transactions.Delete(":id", h.deleteTransaction)

    budgets := router.Group("/budgets")
    budgets.Get("", h.listBudgets)
    budgets.Post("", h.createBudget)
    budgets.Patch(":id", h.updateBudget)
    budgets.Delete(":id", h.deleteBudget)

    debts := router.Group("/debts")
    debts.Get("", h.listDebts)
    debts.Post("", h.createDebt)
    debts.Patch(":id", h.updateDebt)
    debts.Delete(":id", h.deleteDebt)

    fx := router.Group("/fx-rates")
    fx.Get("", h.listFxRates)
    fx.Get(":id", h.getFxRate)

    counterparties := router.Group("/counterparties")
    counterparties.Get("", h.listCounterparties)
    counterparties.Get(":id", h.getCounterparty)
}

func (h *Handler) listAccounts(c *fiber.Ctx) error {
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []domainFinance.Account{sampleAccount()}, meta)
}

func (h *Handler) getAccount(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleAccount(), nil)
}

func (h *Handler) createAccount(c *fiber.Ctx) error {
    var payload domainFinance.Account
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    payload.UpdatedAt = payload.CreatedAt
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) updateAccount(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload["id"] = c.Params("id")
    payload["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) deleteAccount(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "showStatus": "deleted"}, nil)
}

func (h *Handler) adjustBalance(c *fiber.Ctx) error {
    var payload struct {
        NewBalance float64 `json:"newBalance"`
        Reason     string  `json:"reason"`
    }
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "newBalance": payload.NewBalance, "reason": payload.Reason}, nil)
}

func (h *Handler) summary(c *fiber.Ctx) error {
    summary := fiber.Map{
        "totalBalance": fiber.Map{"UZS": 15000000, "USD": 500},
        "totalBalanceInBaseCurrency": 21250000,
        "baseCurrency": "UZS",
        "accountsCount": 5,
        "byType": fiber.Map{"card": 2, "cash": 1, "savings": 2},
    }
    return response.JSONSuccess(c, summary, nil)
}

func (h *Handler) listTransactions(c *fiber.Ctx) error {
    meta := &response.Meta{Page: 1, Limit: 20, Total: 1, TotalPages: 1}
    return response.JSONSuccess(c, []domainFinance.Transaction{sampleTransaction()}, meta)
}

func (h *Handler) getTransaction(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleTransaction(), nil)
}

func (h *Handler) createTransaction(c *fiber.Ctx) error {
    var payload domainFinance.Transaction
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    payload.UpdatedAt = payload.CreatedAt
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) updateTransaction(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload["id"] = c.Params("id")
    payload["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) deleteTransaction(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "deleted"}, nil)
}

func (h *Handler) listBudgets(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainFinance.Budget{sampleBudget()}, nil)
}

func (h *Handler) createBudget(c *fiber.Ctx) error {
    var payload domainFinance.Budget
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) updateBudget(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload["id"] = c.Params("id")
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) deleteBudget(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "deleted"}, nil)
}

func (h *Handler) listDebts(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainFinance.Debt{sampleDebt()}, nil)
}

func (h *Handler) createDebt(c *fiber.Ctx) error {
    var payload domainFinance.Debt
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload.ID = uuid.NewString()
    payload.CreatedAt = time.Now().UTC().Format(time.RFC3339)
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) updateDebt(c *fiber.Ctx) error {
    var payload map[string]any
    if err := c.BodyParser(&payload); err != nil {
        return response.JSONError(c, appErrors.InvalidFinanceData)
    }
    payload["id"] = c.Params("id")
    return response.JSONSuccess(c, payload, nil)
}

func (h *Handler) deleteDebt(c *fiber.Ctx) error {
    return response.JSONSuccess(c, fiber.Map{"id": c.Params("id"), "status": "deleted"}, nil)
}

func (h *Handler) listFxRates(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainFinance.FXRate{sampleFxRate()}, nil)
}

func (h *Handler) getFxRate(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleFxRate(), nil)
}

func (h *Handler) listCounterparties(c *fiber.Ctx) error {
    return response.JSONSuccess(c, []domainFinance.Counterparty{sampleCounterparty()}, nil)
}

func (h *Handler) getCounterparty(c *fiber.Ctx) error {
    return response.JSONSuccess(c, sampleCounterparty(), nil)
}

func sampleAccount() domainFinance.Account {
    now := time.Now().UTC().Format(time.RFC3339)
    return domainFinance.Account{
        BaseEntity: common.BaseEntity{
            ID:        uuid.NewString(),
            UserID:    uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt: now,
            UpdatedAt: now,
        },
        Name:           "Primary Card",
        Type:           domainFinance.AccountTypeCard,
        Currency:       "UZS",
        InitialBalance: 500000,
        CurrentBalance: 500000,
        Icon:           "credit-card",
        Color:          "#4CAF50",
    }
}

func sampleTransaction() domainFinance.Transaction {
    now := time.Now().UTC().Format(time.RFC3339)
    return domainFinance.Transaction{
        BaseEntity: common.BaseEntity{
            ID:        uuid.NewString(),
            UserID:    uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt: now,
            UpdatedAt: now,
        },
        AccountID:    uuid.NewString(),
        Amount:       120000,
        Currency:     "UZS",
        Category:     "food",
        Description:  "Lunch",
        TransactionType: domainFinance.TransactionTypeExpense,
        Date:         now,
    }
}

func sampleBudget() domainFinance.Budget {
    now := time.Now().UTC().Format(time.RFC3339)
    return domainFinance.Budget{
        BaseEntity: common.BaseEntity{
            ID:         uuid.NewString(),
            UserID:     uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt:  now,
            UpdatedAt:  now,
        },
        Name:       "Groceries",
        Currency:   "UZS",
        Limit:      300000,
        Spent:      150000,
        ResetDay:   1,
        AccountIDs: []string{uuid.NewString()},
    }
}

func sampleDebt() domainFinance.Debt {
    now := time.Now().UTC().Format(time.RFC3339)
    return domainFinance.Debt{
        BaseEntity: common.BaseEntity{
            ID:         uuid.NewString(),
            UserID:     uuid.NewString(),
            ShowStatus: common.ShowStatusActive,
            SyncStatus: common.SyncStatusSynced,
            CreatedAt:  now,
            UpdatedAt:  now,
        },
        Name:           "Credit Card",
        Balance:        1200000,
        Interest:       12,
        Currency:       "UZS",
        DueDate:        now,
        MinimumPayment: 50000,
    }
}

func sampleFxRate() domainFinance.FXRate {
    return domainFinance.FXRate{
        ID:        uuid.NewString(),
        Base:      "USD",
        Target:    "UZS",
        Rate:      11660,
        UpdatedAt: time.Now().UTC().Format(time.RFC3339),
    }
}

func sampleCounterparty() domainFinance.Counterparty {
    return domainFinance.Counterparty{
        ID:       uuid.NewString(),
        Name:     "Kapital Bank",
        Type:     "bank",
        LastSynced: time.Now().UTC().Format(time.RFC3339),
    }
}
