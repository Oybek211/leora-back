package modules

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/leora/leora-server/internal/config"
	adminModule "github.com/leora/leora-server/internal/modules/admin"
	authModule "github.com/leora/leora-server/internal/modules/auth"
	financeModule "github.com/leora/leora-server/internal/modules/finance"
	"github.com/leora/leora-server/internal/modules/notifications"
	"github.com/leora/leora-server/internal/modules/planner/focus"
	goalsModule "github.com/leora/leora-server/internal/modules/planner/goals"
	habitsModule "github.com/leora/leora-server/internal/modules/planner/habits"
	tasksModule "github.com/leora/leora-server/internal/modules/planner/tasks"
	premiumModule "github.com/leora/leora-server/internal/modules/premium"
	searchModule "github.com/leora/leora-server/internal/modules/search"
	"github.com/leora/leora-server/internal/modules/users"
	widgetsModule "github.com/leora/leora-server/internal/modules/widgets"
)

// RegisterRoutes wires all module routes.
func RegisterRoutes(app fiber.Router, cfg *config.Config, db *sqlx.DB, cache *redis.Client) {
	authRepo := authModule.NewPostgresRepository(db)
	tokenStore := authModule.NewInMemoryTokenStore()
	authService := authModule.NewService(authRepo, tokenStore, cfg.App.JWTSecret, cfg.App.JWTAccessTTL, cfg.App.JWTRefreshTTL)
	authHandler := authModule.NewHandler(authService)
	authMiddleware := authModule.NewMiddleware(authService)
	authModule.RegisterRoutes(app, authHandler, authMiddleware)

	protected := app.Group("")
	protected.Use(authMiddleware.RequireAuth())

	usersHandler := users.NewHandler(users.NewService(authService, authRepo, cache))
	users.RegisterRoutes(protected, usersHandler, authMiddleware)

	// Tasks module - PostgreSQL
	tasksRepo := tasksModule.NewPostgresRepository(db)
	tasksHandler := tasksModule.NewHandler(tasksModule.NewService(tasksRepo))
	tasksModule.RegisterRoutes(protected, tasksHandler)

	// Goals module - PostgreSQL
	goalsRepo := goalsModule.NewPostgresRepository(db)
	goalsHandler := goalsModule.NewHandler(goalsModule.NewService(goalsRepo))
	goalsModule.RegisterRoutes(protected, goalsHandler)

	// Habits module - PostgreSQL
	habitsRepo := habitsModule.NewPostgresRepository(db)
	habitsHandler := habitsModule.NewHandler(habitsModule.NewService(habitsRepo))
	habitsModule.RegisterRoutes(protected, habitsHandler)

	// Focus module - PostgreSQL
	focusRepo := focus.NewPostgresRepository(db)
	focusHandler := focus.NewHandler(focus.NewService(focusRepo))
	focus.RegisterRoutes(protected, focusHandler)

	// Finance module - PostgreSQL
	financeRepo := financeModule.NewPostgresRepository(db)
	financeHandler := financeModule.NewHandler(financeModule.NewService(financeRepo))
	financeGroup := protected.Group("")
	financeGroup.Use(authMiddleware.RequirePermission("finance:read"))
	financeModule.RegisterRoutes(financeGroup, financeHandler)

	// Notifications module - PostgreSQL
	notificationsRepo := notifications.NewPostgresRepository(db)
	notificationsHandler := notifications.NewHandler(notifications.NewService(notificationsRepo))
	notifications.RegisterRoutes(protected, notificationsHandler)

	// Widgets module - PostgreSQL
	widgetsRepo := widgetsModule.NewPostgresRepository(db)
	widgetsHandler := widgetsModule.NewHandler(widgetsModule.NewService(widgetsRepo))
	widgetsModule.RegisterRoutes(protected, widgetsHandler)

	// Search module - InMemory (qidiruv uchun database kerak emas)
	searchHandler := searchModule.NewHandler(searchModule.NewService(searchModule.NewInMemoryRepository()))
	searchModule.RegisterRoutes(protected, searchHandler)

	// Premium module - PostgreSQL
	premiumRepo := premiumModule.NewPostgresRepository(db)
	premiumHandler := premiumModule.NewHandler(premiumModule.NewService(premiumRepo))
	premiumModule.RegisterRoutes(protected, premiumHandler)

	adminHandler := adminModule.NewHandler(adminModule.NewService(authRepo))
	adminModule.RegisterRoutes(protected, adminHandler, authMiddleware)
}
