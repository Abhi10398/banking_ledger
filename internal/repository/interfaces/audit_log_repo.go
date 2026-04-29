package interfaces

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
)

type AuditRepositoryInterface interface {
	// CreateAuditEntryInTransaction writes an audit record inside an existing transaction.
	CreateAuditEntryInTransaction(ctx context.Context, tx *gorm.DB, entry *db_model.AuditEntry) error
	ListAuditEntries(ctx context.Context, limit, offset int) ([]*db_model.AuditEntry, error)
	// CreateAuditEntryStandalone writes a failure audit outside any transaction
	// (after a rollback). Uses context.Background internally so a cancelled
	// request context does not silently drop the write.
	CreateAuditEntryStandalone(entry *db_model.AuditEntry) error
	GetAuditLogByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*db_model.AuditEntry, error)
}
