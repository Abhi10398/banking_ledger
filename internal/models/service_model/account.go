package service_model

import (
	"time"

	"github.com/google/uuid"

	"awesomeProject/internal/models/db_model"
)

type Account struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Account) FromDBModel(m *db_model.Account) *Account {
	a.ID = m.ID
	a.Name = m.Name
	a.Balance = m.Balance
	a.Currency = m.Currency
	a.CreatedAt = m.CreatedAt
	a.UpdatedAt = m.UpdatedAt
	return a
}

type Transfer struct {
	ID            uuid.UUID               `json:"id"`
	FromAccountID *uuid.UUID              `json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID              `json:"to_account_id,omitempty"`
	Amount        int64                   `json:"amount"`
	Status        db_model.TransferStatus `json:"status"`
	ReversedBy    *uuid.UUID              `json:"reversed_by,omitempty"`
	ReversalOf    *uuid.UUID              `json:"reversal_of,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

func (t *Transfer) FromDBModel(m *db_model.Transfer) *Transfer {
	t.ID = m.ID
	t.FromAccountID = m.FromAccountID
	t.ToAccountID = m.ToAccountID
	t.Amount = m.Amount
	t.Status = m.Status
	t.ReversedBy = m.ReversedBy
	t.ReversalOf = m.ReversalOf
	t.CreatedAt = m.CreatedAt
	t.UpdatedAt = m.UpdatedAt
	return t
}

type AuditEntry struct {
	ID            uuid.UUID               `json:"id"`
	Operation     db_model.AuditOperation `json:"operation"`
	FromAccountID *uuid.UUID              `json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID              `json:"to_account_id,omitempty"`
	Amount        int64                   `json:"amount"`
	Outcome       db_model.AuditOutcome   `json:"outcome"`
	FailureReason *string                 `json:"failure_reason,omitempty"`
	TransferID    *uuid.UUID              `json:"transfer_id,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

func (a *AuditEntry) FromDBModel(m *db_model.AuditEntry) *AuditEntry {
	a.ID = m.ID
	a.Operation = m.Operation
	a.FromAccountID = m.FromAccountID
	a.ToAccountID = m.ToAccountID
	a.Amount = m.Amount
	a.Outcome = m.Outcome
	a.FailureReason = m.FailureReason
	a.TransferID = m.TransferID
	a.CreatedAt = m.CreatedAt
	a.UpdatedAt = m.UpdatedAt
	return a
}
