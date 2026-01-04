package finance

import (
	"errors"

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
	var payload Account
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateAccount(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Account
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	updated, err := h.service.UpdateAccount(c.Context(), id, &payload)
	if err != nil {
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
	if err := h.service.DeleteAccount(c.Context(), id); err != nil {
		if errors.Is(err, appErrors.AccountNotFound) {
			return response.Failure(c, appErrors.AccountNotFound)
		}
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, fiber.Map{"id": id, "status": "deleted"}, nil)
}

func (h *Handler) Transactions(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	data, err := h.service.Transactions(c.Context())
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
	created, err := h.service.CreateTransaction(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Transaction
	if err := c.BodyParser(&payload); err != nil {
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
	data, err := h.service.Budgets(c.Context())
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

func (h *Handler) Debts(c *fiber.Ctx) error {
	page, limit, err := utils.ParsePaginationParams(c.Query("page"), c.Query("limit"))
	if err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	data, err := h.service.Debts(c.Context())
	if err != nil {
		return response.Failure(c, appErrors.InternalServerError)
	}
	start, end := utils.SliceBounds(len(data), page, limit)
	paged := data[start:end]
	return response.Success(c, paged, &response.Meta{Page: page, Limit: limit, Total: len(data), TotalPages: utils.TotalPages(len(data), limit)})
}

func (h *Handler) GetDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	debt, err := h.service.GetDebt(c.Context(), id)
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
	var payload Debt
	if err := c.BodyParser(&payload); err != nil {
		return response.Failure(c, appErrors.InvalidFinanceData)
	}
	created, err := h.service.CreateDebt(c.Context(), &payload)
	if err != nil {
		if typed, ok := err.(*appErrors.Error); ok {
			return response.Failure(c, typed)
		}
		return response.Failure(c, appErrors.InternalServerError)
	}
	return response.Success(c, created, nil)
}

func (h *Handler) UpdateDebt(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload Debt
	if err := c.BodyParser(&payload); err != nil {
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
