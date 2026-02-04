package finance

import (
	"errors"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Handler exposes finance endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Accounts(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	data, err := h.service.Accounts(c.Context())
	if err != nil {
		log.Printf("[Handler.Accounts] Error: %v", err)
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	account, err := h.service.GetAccount(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.AccountNotFound) {
			return response.Failure(c, appErrors.AccountNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, account, nil)
}

func (h *Handler) CreateAccount(c *fiber.Ctx) error {
	type createAccountRequest struct {
		Name           string   `json:"name"`
		Currency       string   `json:"currency"`
		AccountType    string   `json:"accountType"`
		OpeningBalance *float64 `json:"opening_balance"`
		InitialBalance *float64 `json:"initialBalance"`
		LinkedGoalID   *string  `json:"linkedGoalId"`
		CustomTypeID   *string  `json:"customTypeId"`
		IsMain         *bool    `json:"isMain"`
	}
	var payload createAccountRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	account := Account{
		Name:         payload.Name,
		Currency:     payload.Currency,
		AccountType:  payload.AccountType,
		LinkedGoalID: payload.LinkedGoalID,
		CustomTypeID: payload.CustomTypeID,
	}
	if payload.IsMain != nil {
		account.IsMain = *payload.IsMain
	}
	if payload.InitialBalance != nil {
		account.InitialBalance = *payload.InitialBalance
	}
	if payload.OpeningBalance != nil {
		account.InitialBalance = *payload.OpeningBalance
	}
	if err := validateAccount(&account); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, openingTxn, err := h.service.CreateAccount(c.Context(), &account)
	if err != nil {
		log.Printf("[Handler.CreateAccount] Error for name=%s: %v", account.Name, err)
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	summary, _ := h.service.FinanceSummary(c.Context(), "", "", account.Currency, nil)
	return response.SuccessWithStatus(c, fiber.StatusCreated, fiber.Map{
		"account":            created,
		"openingTransaction": openingTxn,
		"transaction":        openingTxn,
		"summary":            summary,
	}, nil)
}

func (h *Handler) UpdateAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Account
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateAccount(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateAccount(c.Context(), id, &payload)
	if err != nil {
		log.Printf("[Handler.UpdateAccount] Error for id=%s: %v", id, err)
		if errors.Is(err, appErrors.AccountNotFound) {
			return response.Failure(c, appErrors.AccountNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.PatchAccount(c.Context(), id, payload)
	if err != nil {
		log.Printf("[Handler.PatchAccount] Error for id=%s: %v", id, err)
		if errors.Is(err, appErrors.AccountNotFound) {
			return response.Failure(c, appErrors.AccountNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	withdrawalTxn, err := h.service.DeleteAccount(c.Context(), id)
	if err != nil {
		log.Printf("[Handler.DeleteAccount] Error for id=%s: %v", id, err)
		if errors.Is(err, appErrors.AccountNotFound) {
			return response.Failure(c, appErrors.AccountNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	summary, _ := h.service.FinanceSummary(c.Context(), "", "", "", nil)
	return response.Success(c, fiber.Map{
		"id":          id,
		"status":      "deleted",
		"transaction": withdrawalTxn,
		"summary":     summary,
	}, nil)
}

func (h *Handler) AccountTransactions(c *fiber.Ctx) error {
	accountID := c.Params("id")
	transactions, err := h.service.AccountTransactions(c.Context(), accountID)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, transactions, nil)
}

func (h *Handler) AccountBalanceHistory(c *fiber.Ctx) error {
	accountID := c.Params("id")
	history, err := h.service.AccountBalanceHistory(c.Context(), accountID)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, history, nil)
}

func (h *Handler) FinanceSummary(c *fiber.Ctx) error {
	dateFrom := c.Query("from")
	dateTo := c.Query("to")
	baseCurrency := c.Query("currency")
	accountIDs := parseAccountIDs(c.Query("accountIds"))
	summary, err := h.service.FinanceSummary(c.Context(), dateFrom, dateTo, baseCurrency, accountIDs)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, summary, nil)
}

func (h *Handler) FinanceBootstrap(c *fiber.Ctx) error {
	dateFrom := c.Query("from")
	dateTo := c.Query("to")
	baseCurrency := c.Query("currency")
	accountIDs := parseAccountIDs(c.Query("accountIds"))
	bootstrap, err := h.service.FinanceBootstrap(c.Context(), dateFrom, dateTo, baseCurrency, accountIDs)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, bootstrap, nil)
}

func (h *Handler) Transactions(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	filter := TransactionFilter{
		AccountID:  c.Query("accountId"),
		Type:       c.Query("type"),
		CategoryID: c.Query("categoryId"),
		DateFrom:   c.Query("dateFrom"),
		DateTo:     c.Query("dateTo"),
		GoalID:     c.Query("goalId"),
		BudgetID:   c.Query("budgetId"),
		DebtID:     c.Query("debtId"),
	}
	data, err := h.service.Transactions(c.Context(), filter)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	txn, err := h.service.GetTransaction(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.TransactionNotFound) {
			return response.Failure(c, appErrors.TransactionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, txn, nil)
}

func (h *Handler) CreateTransaction(c *fiber.Ctx) error {
	var payload Transaction
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateTransaction(&payload); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateTransaction(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) CreateTransfer(c *fiber.Ctx) error {
	var payload Transaction
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	payload.Type = "transfer"
	if err := validateTransaction(&payload); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateTransaction(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) CreateTransactionsBulk(c *fiber.Ctx) error {
	var items []Transaction
	if err := c.BodyParser(&items); err != nil {
		var wrapper struct {
			Items []Transaction `json:"items"`
		}
		if err := c.BodyParser(&wrapper); err != nil {
			return response.Failure(c, appErrors.InvalidFinanceData)
		}
		items = wrapper.Items
	}
	created := make([]*Transaction, 0, len(items))
	for i := range items {
		if err := validateTransaction(&items[i]); err != nil {
			if typed, ok := err.(*appErrors.Error); ok {
				return response.Failure(c, typed)
			}
			return response.Failure(c, appErrors.InvalidFinanceData)
		}
		item := items[i]
		stored, err := h.service.CreateTransaction(c.Context(), &item)
		if err != nil {
			if typed, ok := err.(*appErrors.Error); ok {
				return response.Failure(c, typed)
			}
			return response.Failure(c, appErrors.InternalServerError)
		}
		created = append(created, stored)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Transaction
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateTransaction(&payload); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateTransaction(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.TransactionNotFound) {
			return response.Failure(c, appErrors.TransactionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.PatchTransaction(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.TransactionNotFound) {
			return response.Failure(c, appErrors.TransactionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteTransaction(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.TransactionNotFound) {
			return response.Failure(c, appErrors.TransactionNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) Budgets(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	var isArchived *bool
	if raw := c.Query("isArchived"); raw != "" {
		value := strings.ToLower(raw) == "true"
		isArchived = &value
	}
	filter := BudgetFilter{
		PeriodType:   c.Query("periodType"),
		IsArchived:   isArchived,
		LinkedGoalID: c.Query("linkedGoalId"),
	}
	data, err := h.service.Budgets(c.Context(), filter)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	budget, err := h.service.GetBudget(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.BudgetNotFound) {
			return response.Failure(c, appErrors.BudgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, budget, nil)
}

func (h *Handler) CreateBudget(c *fiber.Ctx) error {
	var payload Budget
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateBudget(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateBudget(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Budget
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateBudget(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateBudget(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.BudgetNotFound) {
			return response.Failure(c, appErrors.BudgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.PatchBudget(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.BudgetNotFound) {
			return response.Failure(c, appErrors.BudgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) AddBudgetValue(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload struct {
		AccountID      string  `json:"account_id"`
		Amount         float64 `json:"amount"`
		AmountCurrency string  `json:"amount_currency"`
		Note           *string `json:"note"`
		Date           *string `json:"date"`
		CategoryID     *string `json:"category_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}

	result, err := h.service.AddBudgetValue(c.Context(), id, BudgetAddValueInput{
		AccountID:      payload.AccountID,
		Amount:         payload.Amount,
		AmountCurrency: payload.AmountCurrency,
		Note:           payload.Note,
		Date:           payload.Date,
		CategoryID:     payload.CategoryID,
	})
	if err != nil {
		log.Printf("[Handler.AddBudgetValue] Error for budget=%s, account=%s, amount=%.2f: %v", id, payload.AccountID, payload.Amount, err)
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	account, _ := h.service.GetAccount(c.Context(), payload.AccountID)
	summary, _ := h.service.FinanceSummary(c.Context(), "", "", result.Budget.Currency, nil)

	return response.Success(c, fiber.Map{
		"budget":      result.Budget,
		"transaction": result.Transaction,
		"account":     account,
		"summary":     summary,
	}, nil)
}

func (h *Handler) DeleteBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteBudget(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.BudgetNotFound) {
			return response.Failure(c, appErrors.BudgetNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) BudgetTransactions(c *fiber.Ctx) error {
	id := c.Params("id")
	transactions, err := h.service.BudgetTransactions(c.Context(), id)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, transactions, nil)
}

func (h *Handler) BudgetSpending(c *fiber.Ctx) error {
	id := c.Params("id")
	items, err := h.service.BudgetSpending(c.Context(), id)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, items, nil)
}

func (h *Handler) RecalculateBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	budget, err := h.service.RecalculateBudget(c.Context(), id)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, budget, nil)
}

func (h *Handler) Debts(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	filter := DebtFilter{
		Direction:    c.Query("direction"),
		Status:       c.Query("status"),
		LinkedGoalID: c.Query("linkedGoalId"),
	}
	// Use the new method that embeds counterparties
	data, err := h.service.DebtsWithCounterparties(c.Context(), filter)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	// Use the new method that embeds counterparty
	debt, err := h.service.GetDebtWithCounterparty(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.DebtNotFound) {
			return response.Failure(c, appErrors.DebtNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, debt, nil)
}

func (h *Handler) CreateDebt(c *fiber.Ctx) error {
	var payload CreateDebtRequest
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}

	if err := validateDebtRequest(&payload); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InvalidFinanceData)
	}

	// Handle inline counterparty creation
	if payload.InlineCounterparty != nil {
		// Create the counterparty first
		counterparty := &Counterparty{
			DisplayName: strings.TrimSpace(payload.InlineCounterparty.DisplayName),
			PhoneNumber: payload.InlineCounterparty.PhoneNumber,
			Comment:     payload.InlineCounterparty.Comment,
		}
		if err := validateCounterparty(counterparty); err != nil {
			if typed, ok := err.(*appErrors.Error); ok {
				return response.Failure(c, typed)
			}
			return response.Failure(c, appErrors.CounterpartyNameTooShort)
		}
		created, err := h.service.CreateCounterparty(c.Context(), counterparty)
		if err != nil {
			if typed, ok := err.(*appErrors.Error); ok {
				return response.Failure(c, typed)
			}
			return response.Failure(c, appErrors.InternalServerError)
		}
		// Link the new counterparty to the debt
		payload.CounterpartyID = &created.ID
		payload.CounterpartyName = created.DisplayName
	}

	created, err := h.service.CreateDebtWithCounterparty(c.Context(), &payload.Debt)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.SuccessWithStatus(c, fiber.StatusCreated, created, nil)
}

func (h *Handler) UpdateDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Debt
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateDebt(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateDebt(c.Context(), id, &payload)
	if err != nil {
		if errors.Is(err, appErrors.DebtNotFound) {
			return response.Failure(c, appErrors.DebtNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) PatchDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.PatchDebt(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.DebtNotFound) {
			return response.Failure(c, appErrors.DebtNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteDebt(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.DebtNotFound) {
			return response.Failure(c, appErrors.DebtNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) RepayDebt(c *fiber.Ctx) error {
	debtID := c.Params("id")
	var payload struct {
		AccountID      string  `json:"account_id"`
		Amount         float64 `json:"amount"`
		AmountCurrency string  `json:"amount_currency"`
		Note           *string `json:"note"`
		Date           *string `json:"date"`
		AppliedRate    float64 `json:"applied_rate"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}

	result, err := h.service.RepayDebt(c.Context(), debtID, DebtValueInput{
		AccountID:      payload.AccountID,
		Amount:         payload.Amount,
		AmountCurrency: payload.AmountCurrency,
		Note:           payload.Note,
		Date:           payload.Date,
		AppliedRate:    payload.AppliedRate,
	})
	if err != nil {
		log.Printf("[Handler.RepayDebt] Error for debt=%s, account=%s, amount=%.2f: %v", debtID, payload.AccountID, payload.Amount, err)
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	account, _ := h.service.GetAccount(c.Context(), payload.AccountID)
	summary, _ := h.service.FinanceSummary(c.Context(), "", "", result.Debt.PrincipalCurrency, nil)

	return response.Success(c, fiber.Map{
		"debt":        result.Debt,
		"payment":     result.Payment,
		"transaction": result.Transaction,
		"account":     account,
		"summary":     summary,
	}, nil)
}

func (h *Handler) AddDebtValue(c *fiber.Ctx) error {
	debtID := c.Params("id")
	var payload struct {
		AccountID      string  `json:"account_id"`
		Amount         float64 `json:"amount"`
		AmountCurrency string  `json:"amount_currency"`
		Note           *string `json:"note"`
		Date           *string `json:"date"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}

	result, err := h.service.AddDebtValue(c.Context(), debtID, DebtValueInput{
		AccountID:      payload.AccountID,
		Amount:         payload.Amount,
		AmountCurrency: payload.AmountCurrency,
		Note:           payload.Note,
		Date:           payload.Date,
	})
	if err != nil {
		log.Printf("[Handler.AddDebtValue] Error for debt=%s, account=%s, amount=%.2f: %v", debtID, payload.AccountID, payload.Amount, err)
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	account, _ := h.service.GetAccount(c.Context(), payload.AccountID)
	summary, _ := h.service.FinanceSummary(c.Context(), "", "", result.Debt.PrincipalCurrency, nil)

	return response.Success(c, fiber.Map{
		"debt":        result.Debt,
		"transaction": result.Transaction,
		"account":     account,
		"summary":     summary,
	}, nil)
}

func (h *Handler) ListDebtPayments(c *fiber.Ctx) error {
	debtID := c.Params("id")
	payments, err := h.service.DebtPayments(c.Context(), debtID)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, payments, nil)
}

func (h *Handler) CreateDebtPayment(c *fiber.Ctx) error {
	debtID := c.Params("id")
	var payload struct {
		DebtPayment
		CreateTransaction bool `json:"createTransaction"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateDebtPayment(&payload.DebtPayment); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	debt, err := h.service.GetDebt(c.Context(), debtID)
	if err != nil {
		return response.Failure(c, appErrors.DebtNotFound)
	}
	payload.DebtID = debtID
	payment, err := h.service.CreateDebtPayment(c.Context(), debt, &payload.DebtPayment)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}

	return response.Success(c, payment, nil)
}

func (h *Handler) UpdateDebtPayment(c *fiber.Ctx) error {
	debtID := c.Params("id")
	paymentID := c.Params("paymentId")
	var payload DebtPayment
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateDebtPayment(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	payload.ID = paymentID
	payload.DebtID = debtID
	debt, err := h.service.GetDebt(c.Context(), debtID)
	if err != nil {
		return response.Failure(c, appErrors.DebtNotFound)
	}
	updated, err := h.service.UpdateDebtPayment(c.Context(), debt, &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteDebtPayment(c *fiber.Ctx) error {
	debtID := c.Params("id")
	paymentID := c.Params("paymentId")
	if err := h.service.DeleteDebtPayment(c.Context(), debtID, paymentID); err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": paymentID, "status": "deleted"}, nil)
}

func (h *Handler) SettleDebt(c *fiber.Ctx) error {
	debtID := c.Params("id")
	updated, err := h.service.SettleDebt(c.Context(), debtID)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) ExtendDebt(c *fiber.Ctx) error {
	debtID := c.Params("id")
	var payload struct {
		DueDate string `json:"dueDate"`
	}
	if err := c.BodyParser(&payload); err != nil || payload.DueDate == "" {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.ExtendDebt(c.Context(), debtID, payload.DueDate)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) Counterparties(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	filter := CounterpartyFilter{Search: c.Query("search")}
	data, err := h.service.Counterparties(c.Context(), filter)
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetCounterparty(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.service.GetCounterparty(c.Context(), id)
	if err != nil {
		if errors.Is(err, appErrors.CounterpartyNotFound) {
			return response.Failure(c, appErrors.CounterpartyNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, item, nil)
}

func (h *Handler) CreateCounterparty(c *fiber.Ctx) error {
	var payload Counterparty
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if err := validateCounterparty(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateCounterparty(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) PatchCounterparty(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.PatchCounterparty(c.Context(), id, payload)
	if err != nil {
		if errors.Is(err, appErrors.CounterpartyNotFound) {
			return response.Failure(c, appErrors.CounterpartyNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) DeleteCounterparty(c *fiber.Ctx) error {
	id := c.Params("id")

	// Check if counterparty has any linked debts
	debts, err := h.service.Debts(c.Context(), DebtFilter{})
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	for _, debt := range debts {
		if debt.CounterpartyID != nil && *debt.CounterpartyID == id {
			return response.Failure(c, appErrors.CounterpartyHasDebts)
		}
	}

	if err := h.service.DeleteCounterparty(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.CounterpartyNotFound) {
			return response.Failure(c, appErrors.CounterpartyNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) CounterpartyDebts(c *fiber.Ctx) error {
	id := c.Params("id")
	debts, err := h.service.Debts(c.Context(), DebtFilter{})
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	filtered := make([]*Debt, 0)
	for _, debt := range debts {
		if debt.CounterpartyID != nil && *debt.CounterpartyID == id {
			filtered = append(filtered, debt)
		}
	}
	return response.Success(c, filtered, nil)
}

func (h *Handler) CounterpartyTransactions(c *fiber.Ctx) error {
	id := c.Params("id")
	transactions, err := h.service.Transactions(c.Context(), TransactionFilter{})
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	filtered := make([]*Transaction, 0)
	for _, txn := range transactions {
		if txn.CounterpartyID != nil && *txn.CounterpartyID == id {
			filtered = append(filtered, txn)
		}
	}
	return response.Success(c, filtered, nil)
}

func (h *Handler) GetFXRates(c *fiber.Ctx) error {
	from := strings.ToUpper(c.Query("from"))
	to := strings.ToUpper(c.Query("to"))
	date := c.Query("date")
	rates, err := h.service.FXRates(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	if from != "" || to != "" || date != "" {
		for _, rate := range rates {
			if from != "" && rate.FromCurrency != from {
				continue
			}
			if to != "" && rate.ToCurrency != to {
				continue
			}
			if date != "" && rate.Date != date {
				continue
			}
			return response.Success(c, rate, nil)
		}
		return response.Failure(c, appErrors.FXRateNotFound)
	}
	return response.Success(c, rates, nil)
}

func (h *Handler) CreateFXRate(c *fiber.Ctx) error {
	var payload FXRate
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	if payload.FromCurrency == "" || payload.ToCurrency == "" || payload.Date == "" {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateFXRate(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) SupportedCurrencies(c *fiber.Ctx) error {
	defaultCurrencies := []string{"USD", "EUR", "UZS", "RUB"}
	catalog := map[string]SupportedCurrency{
		"USD": {Code: "USD", Symbol: "$", Name: "US Dollar", Decimals: 2},
		"EUR": {Code: "EUR", Symbol: "€", Name: "Euro", Decimals: 2},
		"GBP": {Code: "GBP", Symbol: "£", Name: "British Pound", Decimals: 2},
		"UZS": {Code: "UZS", Symbol: "сум", Name: "Uzbek Som", Decimals: 0},
		"RUB": {Code: "RUB", Symbol: "₽", Name: "Russian Ruble", Decimals: 2},
		"JPY": {Code: "JPY", Symbol: "¥", Name: "Japanese Yen", Decimals: 0},
		"CNY": {Code: "CNY", Symbol: "¥", Name: "Chinese Yuan", Decimals: 2},
		"CHF": {Code: "CHF", Symbol: "Fr", Name: "Swiss Franc", Decimals: 2},
		"CAD": {Code: "CAD", Symbol: "C$", Name: "Canadian Dollar", Decimals: 2},
		"AUD": {Code: "AUD", Symbol: "A$", Name: "Australian Dollar", Decimals: 2},
		"TRY": {Code: "TRY", Symbol: "₺", Name: "Turkish Lira", Decimals: 2},
		"SAR": {Code: "SAR", Symbol: "﷼", Name: "Saudi Riyal", Decimals: 2},
		"AED": {Code: "AED", Symbol: "د.إ", Name: "UAE Dirham", Decimals: 2},
		"USDT": {Code: "USDT", Symbol: "₮", Name: "Tether", Decimals: 2},
	}
	rates, _ := h.service.FXRates(c.Context())
	seen := make(map[string]bool)
	for _, currency := range defaultCurrencies {
		seen[currency] = true
	}
	for _, rate := range rates {
		seen[rate.FromCurrency] = true
		seen[rate.ToCurrency] = true
	}
	list := make([]SupportedCurrency, 0, len(seen))
	for currency := range seen {
		code := strings.ToUpper(currency)
		if item, ok := catalog[code]; ok {
			list = append(list, item)
			continue
		}
		decimals := 2
		if code == "UZS" {
			decimals = 0
		}
		list = append(list, SupportedCurrency{
			Code:     code,
			Symbol:   code,
			Name:     code,
			Decimals: decimals,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Code < list[j].Code
	})
	return response.Success(c, list, nil)
}

func (h *Handler) Categories(c *fiber.Ctx) error {
	categoryType := c.Query("type")
	activeOnly := c.Query("active") != "false"
	categories, err := h.service.Categories(c.Context(), categoryType, activeOnly)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, categories, nil)
}

func (h *Handler) QuickExpenseCategories(c *fiber.Ctx) error {
	categoryType := c.Query("type")
	categories, err := h.service.QuickExpenseCategories(c.Context(), categoryType)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, categories, nil)
}

func (h *Handler) UpdateQuickExpenseCategories(c *fiber.Ctx) error {
	type payload struct {
		Type       string                 `json:"type"`
		Categories []QuickExpenseCategory `json:"categories"`
	}
	var req payload
	if err := c.BodyParser(&req); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	// Convert []QuickExpenseCategory to []*QuickExpenseCategory
	categoryPtrs := make([]*QuickExpenseCategory, len(req.Categories))
	for i := range req.Categories {
		categoryPtrs[i] = &req.Categories[i]
	}
	updated, err := h.service.UpdateQuickExpenseCategories(c.Context(), req.Type, categoryPtrs)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	var payload FinanceCategory
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateCategory(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.SuccessWithStatus(c, fiber.StatusCreated, created, nil)
}

func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload FinanceCategory
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateCategory(c.Context(), id, &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, updated, nil)
}

func validateAccount(account *Account) error {
	if strings.TrimSpace(account.Name) == "" {
		return errors.New("name required")
	}
	if strings.TrimSpace(account.Currency) == "" {
		return errors.New("currency required")
	}
	if strings.TrimSpace(account.AccountType) == "" {
		return errors.New("accountType required")
	}
	return nil
}

func parseAccountIDs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	return ids
}

func validateTransaction(txn *Transaction) error {
	if strings.TrimSpace(txn.Type) == "" {
		return errors.New("type required")
	}
	switch txn.Type {
	case "income", "expense":
		if txn.AccountID == nil || *txn.AccountID == "" {
			return errors.New("accountId required")
		}
	case "transfer":
		if txn.FromAccountID == nil || txn.ToAccountID == nil || *txn.FromAccountID == "" || *txn.ToAccountID == "" {
			return errors.New("fromAccountId and toAccountId required")
		}
	case "debt_adjustment":
		if txn.AccountID == nil || *txn.AccountID == "" {
			return errors.New("accountId required")
		}
	default:
		return errors.New("invalid type")
	}
	if strings.TrimSpace(txn.Currency) == "" {
		return appErrors.InvalidCurrency
	}
	if strings.TrimSpace(txn.Date) == "" {
		return errors.New("date required")
	}
	datePart := strings.Split(strings.TrimSpace(txn.Date), "T")[0]
	datePart = strings.Split(datePart, " ")[0]
	if _, err := time.Parse("2006-01-02", datePart); err != nil {
		return appErrors.WithDetails(appErrors.InvalidTransactionDate, map[string]interface{}{"field": "date"})
	}
	if txn.Amount == 0 {
		return appErrors.InvalidAmount
	}
	return nil
}

func validateBudget(budget *Budget) error {
	if strings.TrimSpace(budget.Name) == "" {
		return errors.New("name required")
	}
	if strings.TrimSpace(budget.Currency) == "" {
		return errors.New("currency required")
	}
	return nil
}

func validateDebt(debt *Debt) error {
	// Validate direction
	direction := strings.TrimSpace(debt.Direction)
	if direction != "i_owe" && direction != "they_owe_me" {
		return appErrors.WithDetails(appErrors.InvalidDebtDirection, map[string]interface{}{"field": "direction"})
	}

	// Validate counterparty - must have either ID or name
	hasCounterpartyID := debt.CounterpartyID != nil && strings.TrimSpace(*debt.CounterpartyID) != ""
	hasCounterpartyName := strings.TrimSpace(debt.CounterpartyName) != ""
	if !hasCounterpartyID && !hasCounterpartyName {
		return appErrors.WithDetails(appErrors.CounterpartyRequired, map[string]interface{}{"field": "counterparty"})
	}

	// Validate counterparty name length if provided
	if hasCounterpartyName && len(strings.TrimSpace(debt.CounterpartyName)) < 2 {
		return appErrors.WithDetails(appErrors.CounterpartyNameTooShort, map[string]interface{}{"field": "counterpartyName"})
	}

	if strings.TrimSpace(debt.PrincipalCurrency) == "" {
		return appErrors.WithDetails(appErrors.PrincipalCurrencyRequired, map[string]interface{}{"field": "principalCurrency"})
	}
	if strings.TrimSpace(debt.StartDate) == "" {
		return appErrors.WithDetails(appErrors.DebtStartDateRequired, map[string]interface{}{"field": "startDate"})
	}
	if _, err := time.Parse("2006-01-02", debt.StartDate); err != nil {
		return appErrors.WithDetails(appErrors.InvalidDebtStartDate, map[string]interface{}{"field": "startDate"})
	}
	if debt.PrincipalAmount <= 0 {
		return appErrors.WithDetails(appErrors.InvalidDebtAmount, map[string]interface{}{"field": "principalAmount"})
	}

	// Validate due date is after start date if provided
	if debt.DueDate != nil && strings.TrimSpace(*debt.DueDate) != "" {
		if _, err := time.Parse("2006-01-02", *debt.DueDate); err != nil {
			return appErrors.WithDetails(appErrors.InvalidDebtDueDate, map[string]interface{}{"field": "dueDate"})
		}
		if strings.TrimSpace(debt.StartDate) > strings.TrimSpace(*debt.DueDate) {
			return appErrors.WithDetails(appErrors.InvalidDueDateRange, map[string]interface{}{"field": "dueDate"})
		}
	}

	return nil
}

// validateDebtRequest validates the enhanced debt request with inline counterparty support.
func validateDebtRequest(req *CreateDebtRequest) error {
	// If inline counterparty is provided, validate it
	if req.InlineCounterparty != nil {
		name := strings.TrimSpace(req.InlineCounterparty.DisplayName)
		if name == "" {
			return appErrors.WithDetails(appErrors.CounterpartyRequired, map[string]interface{}{"field": "counterpartyName"})
		}
		if len(name) < 2 {
			return appErrors.WithDetails(appErrors.CounterpartyNameTooShort, map[string]interface{}{"field": "counterpartyName"})
		}
	}

	// Now validate the debt itself, but allow counterparty to come from inline
	if req.InlineCounterparty != nil {
		// Temporarily set counterparty name for validation
		req.CounterpartyName = req.InlineCounterparty.DisplayName
	}

	return validateDebt(&req.Debt)
}

func validateCounterparty(counterparty *Counterparty) error {
	if strings.TrimSpace(counterparty.DisplayName) == "" {
		return errors.New("displayName required")
	}
	return nil
}

func validateDebtPayment(payment *DebtPayment) error {
	if strings.TrimSpace(payment.Currency) == "" {
		return errors.New("currency required")
	}
	if strings.TrimSpace(payment.PaymentDate) == "" {
		return errors.New("paymentDate required")
	}
	if payment.Amount == 0 {
		return errors.New("amount required")
	}
	return nil
}
