//go:build wireinject
// +build wireinject

package factory

import (
	"github.com/google/wire"

	"awesomeProject/internal/api"
	"awesomeProject/internal/api/controller"
	"awesomeProject/internal/service"
	serviceInterface "awesomeProject/internal/service/interfaces"
)

func InitializeAccountAPI() *api.AccountAPI {
	wire.Build(
		api.NewAccountAPI,
		controller.NewFiberApp,
		InitializeAccountService,
		wire.Bind(new(serviceInterface.AccountServiceInterface), new(*service.AccountService)),
	)
	return &api.AccountAPI{}
}

func InitializeTransferAPI() *api.TransferAPI {
	wire.Build(
		api.NewTransferAPI,
		controller.NewFiberApp,
		InitializeTransferService,
		InitializeAccountService,
		wire.Bind(new(serviceInterface.TransferServiceInterface), new(*service.TransferService)),
		wire.Bind(new(serviceInterface.AccountServiceInterface), new(*service.AccountService)),
	)
	return &api.TransferAPI{}
}
