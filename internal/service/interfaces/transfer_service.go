package interfaces

import (
	"context"

	"github.com/google/uuid"

	"awesomeProject/internal/models/service_model"
)

type TransferServiceInterface interface {
	ExecuteTransfer(ctx context.Context, fromID, toID uuid.UUID, amount int64) (*service_model.TransferResult, error)
	ReverseTransfer(ctx context.Context, originalID uuid.UUID) (*service_model.ReversalResult, error)
	GetTransfer(ctx context.Context, id uuid.UUID) (*service_model.Transfer, error)
	ListTransfers(ctx context.Context, limit, offset int) ([]*service_model.Transfer, error)
}
