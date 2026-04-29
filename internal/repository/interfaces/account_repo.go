package interfaces

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
)

type AccountRepositoryInterface interface {
	// BeginTx starts a transaction with the given isolation options.
	BeginTx(ctx context.Context, opts *sql.TxOptions) *gorm.DB

	CreateAccount(ctx context.Context, account *db_model.Account) (*db_model.Account, error)
	ListAccounts(ctx context.Context, limit, offset int) ([]*db_model.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*db_model.Account, error)
	// GetAccountByIDForUpdateInTransaction locks a single row within the given transaction.
	GetAccountByIDForUpdateInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*db_model.Account, error)
	// GetAccountsByIDsForUpdateInTransaction locks both rows in a single query ordered by id,
	// preventing deadlocks between concurrent transfers.
	GetAccountsByIDsForUpdateInTransaction(ctx context.Context, tx *gorm.DB, ids []uuid.UUID) ([]*db_model.Account, error)
	UpdateBalanceInTransaction(ctx context.Context, tx *gorm.DB, account *db_model.Account, newBalance int64) error
}
