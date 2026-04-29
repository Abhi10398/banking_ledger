package interfaces

import (
	"context"

	"github.com/google/uuid"

	"awesomeProject/internal/models/request_model"
	"awesomeProject/internal/models/service_model"
)

type AccountServiceInterface interface {
	CreateAccount(ctx context.Context, req *request_model.CreateAccountRequest) (*service_model.Account, error)
	ListAccounts(ctx context.Context, limit, offset int) ([]*service_model.Account, error)
	GetAccount(ctx context.Context, id uuid.UUID) (*service_model.Account, error)
	Deposit(ctx context.Context, id uuid.UUID, req *request_model.DepositRequest) (*service_model.Account, error)
	Withdraw(ctx context.Context, id uuid.UUID, req *request_model.WithdrawRequest) (*service_model.Account, error)
	GetAuditLog(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*service_model.AuditEntry, error)
	ListAuditEntries(ctx context.Context, limit, offset int) ([]*service_model.AuditEntry, error)
}
