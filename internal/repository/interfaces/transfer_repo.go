package interfaces

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
)

type TransferRepositoryInterface interface {
	CreateTransferInTransaction(ctx context.Context, tx *gorm.DB, transfer *db_model.Transfer) (*db_model.Transfer, error)
	ListTransfers(ctx context.Context, limit, offset int) ([]*db_model.Transfer, error)
	GetTransferByID(ctx context.Context, id uuid.UUID) (*db_model.Transfer, error)
	// GetTransferByIDForUpdateInTransaction locks the transfer row within the given transaction.
	GetTransferByIDForUpdateInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*db_model.Transfer, error)
	UpdateTransferStatusInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID, status db_model.TransferStatus, reversedBy *uuid.UUID) error
	GetTransfersByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*db_model.Transfer, error)
}
