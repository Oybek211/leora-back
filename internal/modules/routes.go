package modules

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/leora/leora-server/internal/config"
	adminModule "github.com/leora/leora-server/internal/modules/admin"
	authModule "github.com/leora/leora-server/internal/modules/auth"
	dashboardModule "github.com/leora/leora-server/internal/modules/dashboard"
	financeModule "github.com/leora/leora-server/internal/modules/finance"
	homeModule "github.com/leora/leora-server/internal/modules/home"
	insightsModule "github.com/leora/leora-server/internal/modules/insights"
	"github.com/leora/leora-server/internal/modules/notifications"
	metaModule "github.com/leora/leora-server/internal/modules/meta"
	"github.com/leora/leora-server/internal/modules/planner/focus"
	goalsModule "github.com/leora/leora-server/internal/modules/planner/goals"
	habitsModule "github.com/leora/leora-server/internal/modules/planner/habits"
	tasksModule "github.com/leora/leora-server/internal/modules/planner/tasks"
	premiumModule "github.com/leora/leora-server/internal/modules/premium"
	reportsModule "github.com/leora/leora-server/internal/modules/reports"
	searchModule "github.com/leora/leora-server/internal/modules/search"
	"github.com/leora/leora-server/internal/modules/users"
	widgetsModule "github.com/leora/leora-server/internal/modules/widgets"
)

// RegisterRoutes wires all module routes.
func RegisterRoutes(app fiber.Router, cfg *config.Config, db *sqlx.DB, cache *redis.Client) {
	authRepo := authModule.NewPostgresRepository(db)
	tokenStore := authModule.NewInMemoryTokenStore()
	authService := authModule.NewService(authRepo, tokenStore, cfg.App.JWTSecret, cfg.App.JWTAccessTTL, cfg.App.JWTRefreshTTL)
	authService.SetGoogleClientID(cfg.App.GoogleOAuthClient)
	authService.SetAppleBundleID(cfg.App.AppleBundleID)
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
	financeHandler := financeModule.NewHandler(financeModule.NewService(financeRepo, cache))
	financeGroup := protected.Group("")
	financeGroup.Use(authMiddleware.RequirePermission("finance:read"))
	financeModule.RegisterRoutes(financeGroup, financeHandler)

	// Dashboard module
	dashboardHandler := dashboardModule.NewHandler(dashboardModule.NewService(db))
	dashboardModule.RegisterRoutes(protected, dashboardHandler)

	// Home module
	homeHandler := homeModule.NewHandler(homeModule.NewService(db))
	homeModule.RegisterRoutes(protected, homeHandler)

	// Reports module
	reportsHandler := reportsModule.NewHandler(reportsModule.NewService(db))
	reportsModule.RegisterRoutes(protected, reportsHandler)

	// Meta module
	metaHandler := metaModule.NewHandler(metaModule.NewService(db))
	metaModule.RegisterRoutes(protected, metaHandler)

	// Insights module
	insightsHandler := insightsModule.NewHandler()
	insightsModule.RegisterRoutes(protected, insightsHandler)

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
