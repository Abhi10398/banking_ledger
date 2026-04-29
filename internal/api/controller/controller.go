package controller

import (
	"encoding/json"
	_ "net/http/pprof"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"awesomeProject/config"
	"awesomeProject/internal/api"
	"awesomeProject/internal/api/controller/middleware"
	"awesomeProject/internal/constants"
	"awesomeProject/logger"
)

type Controller struct {
	App *fiber.App
}

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		BodyLimit:   2 * constants.MB,
	})
	return app
}

func NewController(
	app *fiber.App,
	accountAPI *api.AccountAPI,
	transferAPI *api.TransferAPI,
	config *config.Configuration,
) *Controller {
	configureMiddleWares(app, config)

	// Banking API — all endpoints live under /api.
	app.Mount(constants.APIBasePath, accountAPI.App)
	app.Mount(constants.APIBasePath, transferAPI.App)

	// Serve the dashboard at the root.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	return &Controller{App: app}
}

func configureMiddleWares(app *fiber.App, config *config.Configuration) {
	app.Get("/ping", middleware.PingHealthCheck())
	app.Use(cors.New())
	app.Use(middleware.CorrelationIDMiddleware())
	app.Use(middleware.GetPanicRecoveryMiddleware())
	app.Use(middleware.GetLogMiddleWare())
}

func (c *Controller) Close() {
	logger.Log.Println("Server Closed!")
}
