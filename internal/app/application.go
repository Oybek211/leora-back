package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/config"
	"github.com/leora/leora-server/internal/db"
	"github.com/leora/leora-server/internal/common/localization"
	redisclient "github.com/leora/leora-server/internal/redis"
	"github.com/leora/leora-server/internal/services"
	"github.com/redis/go-redis/v9"
)

// Application wires configuration, services, database, and routes.
type Application struct {
	Config   *config.Config
	FiberApp *fiber.App
	Services *services.Services
	DB       *sqlx.DB
	Cache    *redis.Client
}

// NewApplication builds the application with configuration and handlers.
func NewApplication() (*Application, error) {
	cfg, err := config.Load("configs")
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	fiberApp := fiber.New(fiber.Config{
		Prefork:      false,
		ServerHeader: "Leora",
	})

	dbConn, err := db.NewPostgres(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := db.RunMigrations(dbConn, "migrations"); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	localization.SetDB(dbConn)

	cache, err := redisclient.New(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	svc := services.NewServices()

	application := &Application{
		Config:   cfg,
		FiberApp: fiberApp,
		Services: svc,
		DB:       dbConn,
		Cache:    cache,
	}

	application.setupMiddleware()
	application.registerRoutes()

	return application, nil
}

// Start runs the Fiber server loop.
func (a *Application) Start() error {
	return a.FiberApp.Listen(fmt.Sprintf(":%d", a.Config.App.Port))
}
