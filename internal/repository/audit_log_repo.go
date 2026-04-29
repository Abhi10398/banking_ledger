package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
)

var (
	auditLogRepo     *AuditLogRepo
	auditLogRepoOnce sync.Once
)

type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	auditLogRepoOnce.Do(func() {
		auditLogRepo = &AuditLogRepo{db: db}
	})
	return auditLogRepo
}

func (r *AuditLogRepo) CreateAuditEntryInTransaction(ctx context.Context, tx *gorm.DB, entry *db_model.AuditEntry) error {
	return tx.WithContext(ctx).Create(entry).Error
}

func (r *AuditLogRepo) ListAuditEntries(ctx context.Context, limit, offset int) ([]*db_model.AuditEntry, error) {
	var entries []*db_model.AuditEntry
	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&entries).Error
	return entries, err
}

// CreateAuditEntryStandalone writes a failure audit outside any transaction
// (after a rollback). Uses context.Background internally so a cancelled
// request context does not silently drop the write.
func (r *AuditLogRepo) CreateAuditEntryStandalone(entry *db_model.AuditEntry) error {
	return r.db.WithContext(context.Background()).Create(entry).Error
}

func (r *AuditLogRepo) GetAuditLogByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*db_model.AuditEntry, error) {
	var entries []*db_model.AuditEntry
	err := r.db.WithContext(ctx).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&entries).Error
	return entries, err
}
