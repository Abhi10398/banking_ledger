package api

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awesomeProject/internal/models/request_model"
	serviceInterface "awesomeProject/internal/service/interfaces"
	"awesomeProject/internal/validation"
	"awesomeProject/service_errors"
)

var (
	accountAPI     *AccountAPI
	accountAPIOnce sync.Once
)

type AccountAPI struct {
	App            *fiber.App
	accountService serviceInterface.AccountServiceInterface
}

func NewAccountAPI(app *fiber.App, svc serviceInterface.AccountServiceInterface) *AccountAPI {
	accountAPIOnce.Do(func() {
		accountAPI = &AccountAPI{App: app, accountService: svc}

		app.Get("/accounts", accountAPI.ListAccounts)
		app.Post("/accounts", accountAPI.CreateAccount)
		app.Get("/accounts/:id", accountAPI.GetAccount)
		app.Post("/accounts/:id/deposit", accountAPI.Deposit)
		app.Post("/accounts/:id/withdraw", accountAPI.Withdraw)
		app.Get("/accounts/:id/audit", accountAPI.GetAuditLog)
	})
	return accountAPI
}

func (a *AccountAPI) ListAccounts(c *fiber.Ctx) error {
	limit, offset := pagination(c)
	accounts, err := a.accountService.ListAccounts(c.UserContext(), limit, offset)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(accounts)
}

func (a *AccountAPI) CreateAccount(c *fiber.Ctx) error {
	var req request_model.CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid request body"))
	}
	if err := validation.ValidateStruct(&req); err != nil {
		return service_errors.RespondError(c, err)
	}
	account, err := a.accountService.CreateAccount(c.UserContext(), &req)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusCreated).JSON(account)
}

func (a *AccountAPI) GetAccount(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid account id"))
	}
	account, err := a.accountService.GetAccount(c.UserContext(), id)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(account)
}

func (a *AccountAPI) Deposit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid account id"))
	}
	var req request_model.DepositRequest
	if err = c.BodyParser(&req); err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid request body"))
	}
	if err = validation.ValidateStruct(&req); err != nil {
		return service_errors.RespondError(c, err)
	}
	account, err := a.accountService.Deposit(c.UserContext(), id, &req)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(account)
}

func (a *AccountAPI) Withdraw(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid account id"))
	}
	var req request_model.WithdrawRequest
	if err = c.BodyParser(&req); err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid request body"))
	}
	if err = validation.ValidateStruct(&req); err != nil {
		return service_errors.RespondError(c, err)
	}
	account, err := a.accountService.Withdraw(c.UserContext(), id, &req)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(account)
}

func (a *AccountAPI) GetAuditLog(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return service_errors.RespondError(c, service_errors.BadRequestError("invalid account id"))
	}
	limit, offset := pagination(c)
	entries, err := a.accountService.GetAuditLog(c.UserContext(), id, limit, offset)
	if err != nil {
		return service_errors.RespondError(c, err)
	}
	return c.Status(http.StatusOK).JSON(entries)
}

// pagination reads limit/offset query params with safe defaults and caps.
func pagination(c *fiber.Ctx) (limit, offset int) {
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	offset, _ = strconv.Atoi(c.Query("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return
}
