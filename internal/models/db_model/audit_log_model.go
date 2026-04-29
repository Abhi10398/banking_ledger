package db_model

import (
	"github.com/google/uuid"
)

type AuditOperation string
type AuditOutcome string

const (
	AuditOpTransfer   AuditOperation = "TRANSFER"
	AuditOpReversal   AuditOperation = "REVERSAL"
	AuditOpDeposit    AuditOperation = "DEPOSIT"
	AuditOpWithdrawal AuditOperation = "WITHDRAWAL"

	AuditOutcomeSuccess AuditOutcome = "SUCCESS"
	AuditOutcomeFailure AuditOutcome = "FAILURE"
)

type AuditEntry struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Operation     AuditOperation `gorm:"column:operation;not null"                      json:"operation"`
	FromAccountID *uuid.UUID     `gorm:"type:uuid;column:from_account_id"               json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID     `gorm:"type:uuid;column:to_account_id"                 json:"to_account_id,omitempty"`
	Amount        int64          `gorm:"column:amount;not null"                         json:"amount"`
	Outcome       AuditOutcome   `gorm:"column:outcome;not null"                        json:"outcome"`
	FailureReason *string        `gorm:"column:failure_reason"                          json:"failure_reason,omitempty"`
	TransferID    *uuid.UUID     `gorm:"type:uuid;column:transfer_id"                   json:"transfer_id,omitempty"`
	BaseModel
}

func (AuditEntry) TableName() string {
	return "audit_log"
}
