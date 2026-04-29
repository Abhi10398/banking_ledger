package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
	"awesomeProject/internal/models/service_model"
	repoInterfaces "awesomeProject/internal/repository/interfaces"
	"awesomeProject/logger"
	"awesomeProject/service_errors"
)

var (
	transferService     *TransferService
	transferServiceOnce sync.Once
)

type TransferService struct {
	accountRepo  repoInterfaces.AccountRepositoryInterface
	transferRepo repoInterfaces.TransferRepositoryInterface
	auditRepo    repoInterfaces.AuditRepositoryInterface
}

func NewTransferService(
	accountRepo repoInterfaces.AccountRepositoryInterface,
	transferRepo repoInterfaces.TransferRepositoryInterface,
	auditRepo repoInterfaces.AuditRepositoryInterface,
) *TransferService {
	transferServiceOnce.Do(func() {
		transferService = &TransferService{
			accountRepo:  accountRepo,
			transferRepo: transferRepo,
			auditRepo:    auditRepo,
		}
	})
	return transferService
}

// ExecuteTransfer moves amount (in smallest currency unit) from fromID to toID.
//
// Concurrency safety:
//   - Opens a READ COMMITTED transaction (PostgreSQL default, explicit for clarity).
//   - Locks both account rows in ONE query ordered by id — every concurrent transfer
//     acquires locks in the same sequence, so deadlocks cannot occur.
//   - Audit log written inside the same transaction: always consistent with balance state.
func (s *TransferService) ExecuteTransfer(ctx context.Context, fromID, toID uuid.UUID, amount int64) (*service_model.TransferResult, error) {
	if fromID == toID {
		return nil, service_errors.BadRequestError("from and to accounts must differ")
	}
	if amount <= 0 {
		return nil, service_errors.BadRequestError("amount must be positive")
	}

	tx := s.accountRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

	accounts, err := s.accountRepo.GetAccountsByIDsForUpdateInTransaction(ctx, tx, []uuid.UUID{fromID, toID})
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpTransfer, &fromID, &toID, amount, "accounts not found", nil)
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("one or both accounts not found")
		}
		return nil, service_errors.ServiceError("transfer failed")
	}

	fromAccount, toAccount := resolveAccounts(accounts, fromID, toID)

	if fromAccount.Balance < amount {
		tx.Rollback()
		reason := fmt.Sprintf("insufficient funds: balance %d, requested %d", fromAccount.Balance, amount)
		s.auditFailure(ctx, db_model.AuditOpTransfer, &fromID, &toID, amount, reason, nil)
		return nil, service_errors.BadRequestError("insufficient funds")
	}

	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, fromAccount, fromAccount.Balance-amount); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpTransfer, &fromID, &toID, amount, err.Error(), nil)
		return nil, service_errors.ServiceError("transfer failed: debit error")
	}
	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, toAccount, toAccount.Balance+amount); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpTransfer, &fromID, &toID, amount, err.Error(), nil)
		return nil, service_errors.ServiceError("transfer failed: credit error")
	}

	// Capture balances now — UpdateBalanceInTransaction keeps the struct in sync.
	newFromBalance := fromAccount.Balance
	newToBalance := toAccount.Balance

	transfer := &db_model.Transfer{
		FromAccountID: &fromID,
		ToAccountID:   &toID,
		Amount:        amount,
		Status:        db_model.TransferStatusCompleted,
	}
	created, err := s.transferRepo.CreateTransferInTransaction(ctx, tx, transfer)
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpTransfer, &fromID, &toID, amount, err.Error(), nil)
		return nil, service_errors.ServiceError("transfer failed: record error")
	}

	audit := &db_model.AuditEntry{
		Operation:     db_model.AuditOpTransfer,
		FromAccountID: &fromID,
		ToAccountID:   &toID,
		Amount:        amount,
		Outcome:       db_model.AuditOutcomeSuccess,
		TransferID:    &created.ID,
	}
	if err = s.auditRepo.CreateAuditEntryInTransaction(ctx, tx, audit); err != nil {
		tx.Rollback()
		return nil, service_errors.ServiceError("transfer failed: audit error")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, service_errors.ServiceError("transfer commit failed")
	}

	return &service_model.TransferResult{
		TransferID:     created.ID,
		NewFromBalance: newFromBalance,
		NewToBalance:   newToBalance,
	}, nil
}

// ReverseTransfer reverses a completed transfer. Idempotency is enforced at two layers:
//
//  1. Application layer: lock the original transfer row and check status = 'reversed'.
//  2. Database layer: UNIQUE constraint on transfers.reversal_of. A race that slips
//     past the status check gets a unique-violation (PG 23505), returned as 409.
func (s *TransferService) ReverseTransfer(ctx context.Context, originalID uuid.UUID) (*service_model.ReversalResult, error) {
	tx := s.accountRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

	// Lock the original transfer row — concurrent reversal requests serialize here.
	original, err := s.transferRepo.GetTransferByIDForUpdateInTransaction(ctx, tx, originalID)
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("transfer not found")
		}
		return nil, service_errors.ServiceError("reversal failed")
	}

	// Application-layer idempotency guard.
	switch original.Status {
	case db_model.TransferStatusReversed:
		tx.Rollback()
		return nil, service_errors.BadRequestError("transfer already reversed")
	case db_model.TransferStatusFailed:
		tx.Rollback()
		return nil, service_errors.BadRequestError("cannot reverse a failed transfer")
	}

	if original.FromAccountID == nil || original.ToAccountID == nil {
		tx.Rollback()
		return nil, service_errors.BadRequestError("cannot reverse external deposit or withdrawal")
	}

	fromID := *original.FromAccountID
	toID := *original.ToAccountID

	// Lock both account rows in consistent UUID order.
	accounts, err := s.accountRepo.GetAccountsByIDsForUpdateInTransaction(ctx, tx, []uuid.UUID{fromID, toID})
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpReversal, &toID, &fromID, original.Amount, "accounts not found", &originalID)
		return nil, service_errors.RecordNotFoundError("one or both accounts not found")
	}

	fromAccount, toAccount := resolveAccounts(accounts, fromID, toID)

	// Reversal: money moves toAccount → fromAccount. Check destination can fund it.
	if toAccount.Balance < original.Amount {
		tx.Rollback()
		reason := fmt.Sprintf("insufficient funds for reversal: destination balance %d, needed %d", toAccount.Balance, original.Amount)
		s.auditFailure(ctx, db_model.AuditOpReversal, &toID, &fromID, original.Amount, reason, &originalID)
		return nil, service_errors.BadRequestError("insufficient funds to reverse")
	}

	// Insert reversal — direction is flipped. UNIQUE on reversal_of is the DB guard.
	reversal := &db_model.Transfer{
		FromAccountID: &toID,
		ToAccountID:   &fromID,
		Amount:        original.Amount,
		Status:        db_model.TransferStatusCompleted,
		ReversalOf:    &originalID,
	}
	created, err := s.transferRepo.CreateTransferInTransaction(ctx, tx, reversal)
	if err != nil {
		tx.Rollback()
		if isUniqueViolation(err) {
			return nil, service_errors.BadRequestError("transfer already reversed")
		}
		s.auditFailure(ctx, db_model.AuditOpReversal, &toID, &fromID, original.Amount, err.Error(), &originalID)
		return nil, service_errors.ServiceError("reversal failed: record error")
	}

	if err = s.transferRepo.UpdateTransferStatusInTransaction(ctx, tx, originalID, db_model.TransferStatusReversed, &created.ID); err != nil {
		tx.Rollback()
		return nil, service_errors.ServiceError("reversal failed: could not update original transfer")
	}

	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, toAccount, toAccount.Balance-original.Amount); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpReversal, &toID, &fromID, original.Amount, err.Error(), &originalID)
		return nil, service_errors.ServiceError("reversal failed: debit error")
	}
	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, fromAccount, fromAccount.Balance+original.Amount); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpReversal, &toID, &fromID, original.Amount, err.Error(), &originalID)
		return nil, service_errors.ServiceError("reversal failed: credit error")
	}

	audit := &db_model.AuditEntry{
		Operation:     db_model.AuditOpReversal,
		FromAccountID: &toID,
		ToAccountID:   &fromID,
		Amount:        original.Amount,
		Outcome:       db_model.AuditOutcomeSuccess,
		TransferID:    &created.ID,
	}
	if err = s.auditRepo.CreateAuditEntryInTransaction(ctx, tx, audit); err != nil {
		tx.Rollback()
		return nil, service_errors.ServiceError("reversal failed: audit error")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, service_errors.ServiceError("reversal commit failed")
	}

	return &service_model.ReversalResult{
		ReversalID:         created.ID,
		OriginalTransferID: originalID,
	}, nil
}

func (s *TransferService) GetTransfer(ctx context.Context, id uuid.UUID) (*service_model.Transfer, error) {
	t, err := s.transferRepo.GetTransferByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("transfer not found")
		}
		logger.Get(ctx).Errorf("GetTransfer failed: %v", err)
		return nil, service_errors.ServiceError("failed to fetch transfer")
	}
	return new(service_model.Transfer).FromDBModel(t), nil
}

func (s *TransferService) ListTransfers(ctx context.Context, limit, offset int) ([]*service_model.Transfer, error) {
	transfers, err := s.transferRepo.ListTransfers(ctx, limit, offset)
	if err != nil {
		logger.Get(ctx).Errorf("ListTransfers failed: %v", err)
		return nil, service_errors.ServiceError("failed to list transfers")
	}
	result := make([]*service_model.Transfer, 0, len(transfers))
	for _, t := range transfers {
		result = append(result, new(service_model.Transfer).FromDBModel(t))
	}
	return result, nil
}

// resolveAccounts maps the sorted query result back to the logical from/to pair.
func resolveAccounts(sorted []*db_model.Account, fromID, toID uuid.UUID) (from, to *db_model.Account) {
	if sorted[0].ID == fromID {
		return sorted[0], sorted[1]
	}
	return sorted[1], sorted[0]
}

// isUniqueViolation returns true for PostgreSQL unique-constraint violations (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *TransferService) auditFailure(
	ctx context.Context,
	op db_model.AuditOperation,
	fromID, toID *uuid.UUID,
	amount int64,
	reason string,
	transferID *uuid.UUID,
) {
	entry := &db_model.AuditEntry{
		Operation:     op,
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Outcome:       db_model.AuditOutcomeFailure,
		FailureReason: &reason,
		TransferID:    transferID,
	}
	if err := s.auditRepo.CreateAuditEntryStandalone(entry); err != nil {
		logger.Get(ctx).Errorf("auditFailure write failed: %v", err)
	}
}
