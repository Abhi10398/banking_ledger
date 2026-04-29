//go:build wireinject
// +build wireinject

package factory

import (
	"github.com/google/wire"

	"awesomeProject/internal/repository"
	repositoryInterfaces "awesomeProject/internal/repository/interfaces"
	"awesomeProject/internal/service"
)

func InitializeAccountService() *service.AccountService {
	wire.Build(
		service.NewAccountService,
		InitializeAccountRepo,
		InitializeTransferRepo,
		InitializeAuditLogRepo,
		wire.Bind(new(repositoryInterfaces.AccountRepositoryInterface), new(*repository.AccountRepo)),
		wire.Bind(new(repositoryInterfaces.TransferRepositoryInterface), new(*repository.TransferRepo)),
		wire.Bind(new(repositoryInterfaces.AuditRepositoryInterface), new(*repository.AuditLogRepo)),
	)
	return &service.AccountService{}
}

func InitializeTransferService() *service.TransferService {
	wire.Build(
		service.NewTransferService,
		InitializeAccountRepo,
		InitializeTransferRepo,
		InitializeAuditLogRepo,
		wire.Bind(new(repositoryInterfaces.AccountRepositoryInterface), new(*repository.AccountRepo)),
		wire.Bind(new(repositoryInterfaces.TransferRepositoryInterface), new(*repository.TransferRepo)),
		wire.Bind(new(repositoryInterfaces.AuditRepositoryInterface), new(*repository.AuditLogRepo)),
	)
	return &service.TransferService{}
}
