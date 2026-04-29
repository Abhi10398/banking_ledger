package repository

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"awesomeProject/internal/models/db_model"
)

var (
	accountRepo     *AccountRepo
	accountRepoOnce sync.Once
)

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	accountRepoOnce.Do(func() {
		accountRepo = &AccountRepo{db: db}
	})
	return accountRepo
}

func (r *AccountRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) *gorm.DB {
	return r.db.WithContext(ctx).Begin(opts)
}

func (r *AccountRepo) CreateAccount(ctx context.Context, account *db_model.Account) (*db_model.Account, error) {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepo) ListAccounts(ctx context.Context, limit, offset int) ([]*db_model.Account, error) {
	var accounts []*db_model.Account
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepo) GetAccountByID(ctx context.Context, id uuid.UUID) (*db_model.Account, error) {
	var account db_model.Account
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetAccountByIDForUpdateInTransaction acquires an exclusive row lock on a single account.
// Use for deposit/withdraw where only one account is involved.
func (r *AccountRepo) GetAccountByIDForUpdateInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*db_model.Account, error) {
	var account db_model.Account
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&account, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetAccountsByIDsForUpdateInTransaction acquires exclusive row locks on both accounts in a
// single query ordered by id. Ordering guarantees every concurrent transfer
// acquires locks in the same sequence, eliminating deadlocks.
func (r *AccountRepo) GetAccountsByIDsForUpdateInTransaction(ctx context.Context, tx *gorm.DB, ids []uuid.UUID) ([]*db_model.Account, error) {
	var accounts []*db_model.Account
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", ids).
		Order("id").
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	if len(accounts) != len(ids) {
		return nil, gorm.ErrRecordNotFound
	}
	return accounts, nil
}

func (r *AccountRepo) UpdateBalanceInTransaction(ctx context.Context, tx *gorm.DB, account *db_model.Account, newBalance int64) error {
	if err := tx.WithContext(ctx).Model(account).Update("balance", newBalance).Error; err != nil {
		return err
	}
	account.Balance = newBalance
	return nil
}
