package app

import (
	modules "github.com/leora/leora-server/internal/modules"
	achievementsTransport "github.com/leora/leora-server/internal/transport/http/v1/achievements"
	aiTransport "github.com/leora/leora-server/internal/transport/http/v1/ai"
	dataTransport "github.com/leora/leora-server/internal/transport/http/v1/data"
	devicesTransport "github.com/leora/leora-server/internal/transport/http/v1/devices"
	insightsTransport "github.com/leora/leora-server/internal/transport/http/v1/insights"
	integrationsTransport "github.com/leora/leora-server/internal/transport/http/v1/integrations"
	settingsTransport "github.com/leora/leora-server/internal/transport/http/v1/settings"
	syncTransport "github.com/leora/leora-server/internal/transport/http/v1/sync"
)

func (a *Application) registerRoutes() {
	base := a.FiberApp.Group(a.Config.App.BasePath)

	modules.RegisterRoutes(base, a.Config, a.DB, a.Cache)

	insightsTransport.RegisterRoutes(base)
	syncTransport.RegisterRoutes(base)
	settingsTransport.RegisterRoutes(base)
	achievementsTransport.RegisterRoutes(base)
	aiTransport.RegisterRoutes(base)
	dataTransport.RegisterRoutes(base)
	integrationsTransport.RegisterRoutes(base)
	devicesTransport.RegisterRoutes(base)
}
