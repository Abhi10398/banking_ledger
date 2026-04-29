package db_model

import (
	"github.com/google/uuid"
)

type TransferStatus string

const (
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusReversed  TransferStatus = "reversed"
	TransferStatusFailed    TransferStatus = "failed"
)

type Transfer struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FromAccountID *uuid.UUID     `gorm:"type:uuid;column:from_account_id"               json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID     `gorm:"type:uuid;column:to_account_id"                 json:"to_account_id,omitempty"`
	Amount        int64          `gorm:"column:amount;not null"                         json:"amount"`
	Status        TransferStatus `gorm:"column:status;not null;default:'completed'"     json:"status"`
	ReversedBy    *uuid.UUID     `gorm:"type:uuid;column:reversed_by"                   json:"reversed_by,omitempty"`
	ReversalOf    *uuid.UUID     `gorm:"type:uuid;column:reversal_of;uniqueIndex"       json:"reversal_of,omitempty"`
	BaseModel
}

func (Transfer) TableName() string {
	return "transfers"
}
