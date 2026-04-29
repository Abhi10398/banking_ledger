package api

import (
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awesomeProject/internal/models/request_model"
	serviceInterface "awesomeProject/internal/service/interfaces"
	"awesomeProject/internal/validation"
	"awesomeProject/service_errors"
)

var (
	transferAPI     *TransferAPI
	transferAPIOnce sync.Once
)

type TransferAPI struct {
	App             *fiber.App
	transferService serviceInterface.TransferServiceInterface
	accountService  serviceInterface.AccountServiceInterface
}

func NewTransferAPI(app *fiber.App, transferSvc serviceInterface.TransferServiceInterface, accountSvc serviceInterface.AccountServiceInterface) *TransferAPI {
	transferAPIOnce.Do(func() {
		transferAPI = &TransferAPI{App: app, transferService: transferSvc, accountService: accountSvc}

		app.Post("/transfers", transferAPI.ExecuteTransfer)
		app.Post("/transfers/:id/reverse", transferAPI.ReverseTransfer)
		app.Get("/transfers", transferAPI.ListTransfers)
		app.Get("/audit", transferAPI.ListAuditEntries)
	})
	return transferAPI
}

// ExecuteTransfer handles POST /api/transfers
func (a *TransferAPI) ExecuteTransfer(c *fiber.Ctx) error {
	var req request_model.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid request body"))
	}
	if err := validation.ValidateStruct(&req); err != nil {
		return service_errors.RespondError(c, err)
	}
	fromID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid from_account_id"))
	}
	toID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid to_account_id"))
	}

	result, err := a.transferService.ExecuteTransfer(c.UserContext(), fromID, toID, req.Amount)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(result)
}

// ReverseTransfer handles POST /api/transfers/{id}/reverse
func (a *TransferAPI) ReverseTransfer(c *fiber.Ctx) error {
	originalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid transfer id"))
	}
	result, err := a.transferService.ReverseTransfer(c.UserContext(), originalID)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(result)
}

// ListTransfers handles GET /api/transfers?limit=20&offset=0
func (a *TransferAPI) ListTransfers(c *fiber.Ctx) error {
	limit, offset := pagination(c)
	transfers, err := a.transferService.ListTransfers(c.UserContext(), limit, offset)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(transfers)
}

// ListAuditEntries handles GET /api/audit?limit=20&offset=0
func (a *TransferAPI) ListAuditEntries(c *fiber.Ctx) error {
	limit, offset := pagination(c)
	entries, err := a.accountService.ListAuditEntries(c.UserContext(), limit, offset)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(entries)
}
