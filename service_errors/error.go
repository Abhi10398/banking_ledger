package service_errors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var _ error = Error{}

var (
	contextCancelled = Error{Code: 513, Name: "Internal server error", Description: context.Canceled.Error()}

	BadRequest = Invalid("req")
)

type Error struct {
	Code        int    `json:"code"`
	Name        string `json:"error"`
	Description string `json:"-"`
}

func (e Error) Error() string {
	if e.Description == "" {
		return e.Name
	}
	return fmt.Sprintf("%s (%s)", e.Name, e.Description)
}

func InternalServerError() Error {
	return Error{Code: 500, Name: "Internal server error"}
}

func Invalid(key string) Error {
	return Error{Code: 400, Name: fmt.Sprintf("invalid_%s", key)}
}

func New(text string) Error {
	return Error{Code: 500, Name: text}
}

func From(code int, name string) Error {
	return Error{Code: code, Name: name}
}

func FromError(err error) Error {
	var e Error
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, context.Canceled) {
		return contextCancelled
	}
	e = InternalServerError()
	log.Printf(err.Error())
	return e
}

func IsRetryable(err error) bool {
	return FromError(err).Code >= 500
}

func ServiceError(message string) Error {
	return Error{Name: message, Code: http.StatusInternalServerError}
}

func ServiceErrorWithErrorCode(message string, errorCode int) Error {
	return Error{Name: message, Code: errorCode}
}

func BadRequestError(message string) Error {
	return Error{Name: message, Code: http.StatusBadRequest}
}

func RecordNotFoundError(message string) Error {
	return Error{Name: message, Code: http.StatusNotFound}
}

func UnauthorizedError(message string) Error {
	return Error{Name: message, Code: http.StatusUnauthorized}
}

func TooManyRequestsError(message string) Error {
	return Error{Name: message, Code: http.StatusTooManyRequests}
}

// AppError is the canonical error envelope for all banking API responses.
type AppError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e AppError) Error() string { return e.Message }

// Sentinel error codes that API clients key on.
const (
	CodeInsufficientFunds = "insufficient_funds"
	CodeAccountNotFound   = "account_not_found"
	CodeTransferNotFound  = "transfer_not_found"
	CodeAlreadyReversed   = "already_reversed"
	CodeCannotReverse     = "cannot_reverse"
	CodeInvalidAmount     = "invalid_amount"
	CodeInvalidRequest    = "invalid_request"
	CodeInternal          = "internal_error"
)

// RespondError maps a service-layer error to an AppError and writes the JSON response.
func RespondError(c *fiber.Ctx, err error) error {
	ae := toAppError(err)
	return c.Status(ae.Status).JSON(ae)
}

func toAppError(err error) AppError {
	se := FromError(err)
	name := strings.ToLower(se.Name)

	switch {
	case se.Code == http.StatusNotFound && strings.Contains(name, "transfer"):
		return AppError{Code: CodeTransferNotFound, Message: se.Name, Status: http.StatusNotFound}
	case se.Code == http.StatusNotFound:
		return AppError{Code: CodeAccountNotFound, Message: se.Name, Status: http.StatusNotFound}
	case strings.Contains(name, "already reversed"):
		return AppError{Code: CodeAlreadyReversed, Message: se.Name, Status: http.StatusConflict}
	case strings.Contains(name, "cannot reverse"):
		return AppError{Code: CodeCannotReverse, Message: se.Name, Status: http.StatusUnprocessableEntity}
	case strings.Contains(name, "insufficient funds"):
		return AppError{Code: CodeInsufficientFunds, Message: se.Name, Status: http.StatusUnprocessableEntity}
	case strings.Contains(name, "invalid") || se.Code == http.StatusBadRequest:
		return AppError{Code: CodeInvalidRequest, Message: se.Name, Status: http.StatusBadRequest}
	case strings.Contains(name, "amount"):
		return AppError{Code: CodeInvalidAmount, Message: se.Name, Status: http.StatusBadRequest}
	default:
		return AppError{Code: CodeInternal, Message: "an unexpected error occurred", Status: http.StatusInternalServerError}
	}
}
