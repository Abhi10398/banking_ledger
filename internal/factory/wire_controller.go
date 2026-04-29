//go:build wireinject
// +build wireinject

package factory

import (
	"github.com/google/wire"

	"awesomeProject/config"
	"awesomeProject/internal/api/controller"
)

func InitializeController() *controller.Controller {
	wire.Build(
		controller.NewController,
		controller.NewFiberApp,
		InitializeAccountAPI,
		InitializeTransferAPI,
		config.Load,
	)
	return &controller.Controller{}
}
