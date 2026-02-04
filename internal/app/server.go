package app

import (
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func (a *Application) setupMiddleware() {
	a.FiberApp.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))
	a.FiberApp.Use(logger.New())
	a.FiberApp.Use(recover.New())
}
