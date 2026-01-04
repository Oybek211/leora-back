package finance

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	accounts := router.Group("/accounts")
	accounts.Get("", handler.Accounts)
	accounts.Post("", handler.CreateAccount)
	accounts.Get("/:id", handler.GetAccount)
	accounts.Put("/:id", handler.UpdateAccount)
	accounts.Patch("/:id", handler.PatchAccount)
	accounts.Delete("/:id", handler.DeleteAccount)

	transactions := router.Group("/transactions")
	transactions.Get("", handler.Transactions)
	transactions.Post("", handler.CreateTransaction)
	transactions.Get("/:id", handler.GetTransaction)
	transactions.Put("/:id", handler.UpdateTransaction)
	transactions.Patch("/:id", handler.PatchTransaction)
	transactions.Delete("/:id", handler.DeleteTransaction)

	budgets := router.Group("/budgets")
	budgets.Get("", handler.Budgets)
	budgets.Post("", handler.CreateBudget)
	budgets.Get("/:id", handler.GetBudget)
	budgets.Put("/:id", handler.UpdateBudget)
	budgets.Patch("/:id", handler.PatchBudget)
	budgets.Delete("/:id", handler.DeleteBudget)

	debts := router.Group("/debts")
	debts.Get("", handler.Debts)
	debts.Post("", handler.CreateDebt)
	debts.Get("/:id", handler.GetDebt)
	debts.Put("/:id", handler.UpdateDebt)
	debts.Patch("/:id", handler.PatchDebt)
	debts.Delete("/:id", handler.DeleteDebt)
}
