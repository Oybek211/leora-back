package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/leora/leora-server/internal/common/response"
	"github.com/leora/leora-server/internal/config"
	"github.com/leora/leora-server/internal/db"
	appErrors "github.com/leora/leora-server/internal/errors"
	"github.com/leora/leora-server/internal/modules"
	redisclient "github.com/leora/leora-server/internal/redis"
)

func main() {
	cfg, err := config.Load("configs")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	dbConn, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(dbConn, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	cache, err := redisclient.New(cfg.Redis)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	app := fiber.New(fiber.Config{
		Prefork:      false,
		ServerHeader: cfg.App.Name,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Failure(c, appErrors.InternalServerError)
		},
	})
	app.Use(logger.New())
	app.Use(recover.New())

	moduleRouter := app.Group(cfg.App.BasePath)
	modules.RegisterRoutes(moduleRouter, cfg, dbConn, cache)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
