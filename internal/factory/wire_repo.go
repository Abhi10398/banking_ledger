//go:build wireinject
// +build wireinject

package factory

import (
	"github.com/google/wire"

	"awesomeProject/internal/appcontext/database/postgres"
	"awesomeProject/internal/repository"
)

func InitializeAccountRepo() *repository.AccountRepo {
	wire.Build(repository.NewAccountRepo, postgres.GetDB)
	return &repository.AccountRepo{}
}

func InitializeTransferRepo() *repository.TransferRepo {
	wire.Build(repository.NewTransferRepo, postgres.GetDB)
	return &repository.TransferRepo{}
}

func InitializeAuditLogRepo() *repository.AuditLogRepo {
	wire.Build(repository.NewAuditLogRepo, postgres.GetDB)
	return &repository.AuditLogRepo{}
}
