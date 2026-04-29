package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"awesomeProject/internal/models/db_model"
)

var (
	transferRepo     *TransferRepo
	transferRepoOnce sync.Once
)

type TransferRepo struct {
	db *gorm.DB
}

func NewTransferRepo(db *gorm.DB) *TransferRepo {
	transferRepoOnce.Do(func() {
		transferRepo = &TransferRepo{db: db}
	})
	return transferRepo
}

func (r *TransferRepo) CreateTransferInTransaction(ctx context.Context, tx *gorm.DB, transfer *db_model.Transfer) (*db_model.Transfer, error) {
	if err := tx.WithContext(ctx).Create(transfer).Error; err != nil {
		return nil, err
	}
	return transfer, nil
}

func (r *TransferRepo) ListTransfers(ctx context.Context, limit, offset int) ([]*db_model.Transfer, error) {
	var transfers []*db_model.Transfer
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&transfers).Error
	return transfers, err
}

func (r *TransferRepo) GetTransferByID(ctx context.Context, id uuid.UUID) (*db_model.Transfer, error) {
	var transfer db_model.Transfer
	if err := r.db.WithContext(ctx).First(&transfer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

// GetTransferByIDForUpdateInTransaction locks the transfer row so concurrent reversal
// requests serialize on this row before reading status.
func (r *TransferRepo) GetTransferByIDForUpdateInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*db_model.Transfer, error) {
	var transfer db_model.Transfer
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&transfer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *TransferRepo) UpdateTransferStatusInTransaction(ctx context.Context, tx *gorm.DB, id uuid.UUID, status db_model.TransferStatus, reversedBy *uuid.UUID) error {
	updates := map[string]interface{}{"status": status}
	if reversedBy != nil {
		updates["reversed_by"] = reversedBy
	}
	return tx.WithContext(ctx).Model(&db_model.Transfer{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TransferRepo) GetTransfersByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*db_model.Transfer, error) {
	var transfers []*db_model.Transfer
	err := r.db.WithContext(ctx).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transfers).Error
	return transfers, err
}
