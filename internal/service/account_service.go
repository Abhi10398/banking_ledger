package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject/internal/models/db_model"
	"awesomeProject/internal/models/request_model"
	"awesomeProject/internal/models/service_model"
	repoInterfaces "awesomeProject/internal/repository/interfaces"
	"awesomeProject/logger"
	"awesomeProject/service_errors"
)

var (
	accountService     *AccountService
	accountServiceOnce sync.Once
)

type AccountService struct {
	accountRepo  repoInterfaces.AccountRepositoryInterface
	transferRepo repoInterfaces.TransferRepositoryInterface
	auditRepo    repoInterfaces.AuditRepositoryInterface
}

func NewAccountService(
	accountRepo repoInterfaces.AccountRepositoryInterface,
	transferRepo repoInterfaces.TransferRepositoryInterface,
	auditRepo repoInterfaces.AuditRepositoryInterface,
) *AccountService {
	accountServiceOnce.Do(func() {
		accountService = &AccountService{
			accountRepo:  accountRepo,
			transferRepo: transferRepo,
			auditRepo:    auditRepo,
		}
	})
	return accountService
}

func (s *AccountService) CreateAccount(ctx context.Context, req *request_model.CreateAccountRequest) (*service_model.Account, error) {
	account := &db_model.Account{
		Name:     req.Name,
		Currency: req.Currency,
		Balance:  0,
	}
	created, err := s.accountRepo.CreateAccount(ctx, account)
	if err != nil {
		logger.Get(ctx).Errorf("CreateAccount failed: %v", err)
		return nil, service_errors.ServiceError("failed to create account")
	}
	return new(service_model.Account).FromDBModel(created), nil
}

func (s *AccountService) ListAccounts(ctx context.Context, limit, offset int) ([]*service_model.Account, error) {
	accounts, err := s.accountRepo.ListAccounts(ctx, limit, offset)
	if err != nil {
		logger.Get(ctx).Errorf("ListAccounts failed: %v", err)
		return nil, service_errors.ServiceError("failed to list accounts")
	}
	result := make([]*service_model.Account, 0, len(accounts))
	for _, a := range accounts {
		result = append(result, new(service_model.Account).FromDBModel(a))
	}
	return result, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id uuid.UUID) (*service_model.Account, error) {
	account, err := s.accountRepo.GetAccountByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("account not found")
		}
		logger.Get(ctx).Errorf("GetAccount failed: %v", err)
		return nil, service_errors.ServiceError("failed to fetch account")
	}
	return new(service_model.Account).FromDBModel(account), nil
}

func (s *AccountService) Deposit(ctx context.Context, id uuid.UUID, req *request_model.DepositRequest) (*service_model.Account, error) {
	tx := s.accountRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

	account, err := s.accountRepo.GetAccountByIDForUpdateInTransaction(ctx, tx, id)
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpDeposit, nil, &id, req.Amount, "account not found")
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("account not found")
		}
		return nil, service_errors.ServiceError("deposit failed")
	}

	newBalance := account.Balance + req.Amount
	transfer := &db_model.Transfer{ToAccountID: &id, Amount: req.Amount, Status: db_model.TransferStatusCompleted}

	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, account, newBalance); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpDeposit, nil, &id, req.Amount, err.Error())
		return nil, service_errors.ServiceError("deposit failed")
	}

	created, err := s.transferRepo.CreateTransferInTransaction(ctx, tx, transfer)
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpDeposit, nil, &id, req.Amount, err.Error())
		return nil, service_errors.ServiceError("deposit failed")
	}

	audit := &db_model.AuditEntry{
		Operation:   db_model.AuditOpDeposit,
		ToAccountID: &id,
		Amount:      req.Amount,
		Outcome:     db_model.AuditOutcomeSuccess,
		TransferID:  &created.ID,
	}
	if err = s.auditRepo.CreateAuditEntryInTransaction(ctx, tx, audit); err != nil {
		tx.Rollback()
		return nil, service_errors.ServiceError("deposit failed: audit error")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, service_errors.ServiceError("deposit commit failed")
	}
	return new(service_model.Account).FromDBModel(account), nil
}

func (s *AccountService) Withdraw(ctx context.Context, id uuid.UUID, req *request_model.WithdrawRequest) (*service_model.Account, error) {
	tx := s.accountRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

	account, err := s.accountRepo.GetAccountByIDForUpdateInTransaction(ctx, tx, id)
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpWithdrawal, &id, nil, req.Amount, "account not found")
		if err == gorm.ErrRecordNotFound {
			return nil, service_errors.RecordNotFoundError("account not found")
		}
		return nil, service_errors.ServiceError("withdraw failed")
	}

	if account.Balance < req.Amount {
		tx.Rollback()
		reason := fmt.Sprintf("insufficient funds: balance %d, requested %d", account.Balance, req.Amount)
		s.auditFailure(ctx, db_model.AuditOpWithdrawal, &id, nil, req.Amount, reason)
		return nil, service_errors.BadRequestError("insufficient funds")
	}

	newBalance := account.Balance - req.Amount
	transfer := &db_model.Transfer{FromAccountID: &id, Amount: req.Amount, Status: db_model.TransferStatusCompleted}

	if err = s.accountRepo.UpdateBalanceInTransaction(ctx, tx, account, newBalance); err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpWithdrawal, &id, nil, req.Amount, err.Error())
		return nil, service_errors.ServiceError("withdraw failed")
	}

	created, err := s.transferRepo.CreateTransferInTransaction(ctx, tx, transfer)
	if err != nil {
		tx.Rollback()
		s.auditFailure(ctx, db_model.AuditOpWithdrawal, &id, nil, req.Amount, err.Error())
		return nil, service_errors.ServiceError("withdraw failed")
	}

	audit := &db_model.AuditEntry{
		Operation:     db_model.AuditOpWithdrawal,
		FromAccountID: &id,
		Amount:        req.Amount,
		Outcome:       db_model.AuditOutcomeSuccess,
		TransferID:    &created.ID,
	}
	if err = s.auditRepo.CreateAuditEntryInTransaction(ctx, tx, audit); err != nil {
		tx.Rollback()
		return nil, service_errors.ServiceError("withdraw failed: audit error")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, service_errors.ServiceError("withdraw commit failed")
	}
	return new(service_model.Account).FromDBModel(account), nil
}

func (s *AccountService) GetAuditLog(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*service_model.AuditEntry, error) {
	entries, err := s.auditRepo.GetAuditLogByAccount(ctx, accountID, limit, offset)
	if err != nil {
		logger.Get(ctx).Errorf("GetAuditLog failed: %v", err)
		return nil, service_errors.ServiceError("failed to fetch audit log")
	}
	return toAuditEntries(entries), nil
}

func (s *AccountService) ListAuditEntries(ctx context.Context, limit, offset int) ([]*service_model.AuditEntry, error) {
	entries, err := s.auditRepo.ListAuditEntries(ctx, limit, offset)
	if err != nil {
		logger.Get(ctx).Errorf("ListAuditEntries failed: %v", err)
		return nil, service_errors.ServiceError("failed to list audit entries")
	}
	return toAuditEntries(entries), nil
}

func toAuditEntries(rows []*db_model.AuditEntry) []*service_model.AuditEntry {
	result := make([]*service_model.AuditEntry, 0, len(rows))
	for _, e := range rows {
		result = append(result, new(service_model.AuditEntry).FromDBModel(e))
	}
	return result
}

func (s *AccountService) auditFailure(ctx context.Context, op db_model.AuditOperation, fromID, toID *uuid.UUID, amount int64, reason string) {
	entry := &db_model.AuditEntry{
		Operation:     op,
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Outcome:       db_model.AuditOutcomeFailure,
		FailureReason: &reason,
	}
	if err := s.auditRepo.CreateAuditEntryStandalone(entry); err != nil {
		logger.Get(ctx).Errorf("auditFailure write failed: %v", err)
	}
}
