package middleware

import (
	"github.com/gofiber/fiber/v2"

	"awesomeProject/internal/constants"
	"awesomeProject/internal/models/response_model"
	"awesomeProject/service_errors"
)

func ValidateJWT(c *fiber.Ctx) error {
	token := c.Get(constants.Authorisation)
	if len(token) == 0 {
		commonErr := service_errors.BadRequestError("Authorization token is not present in request header")
		return response_model.GetResponseV2(c, commonErr.Code, commonErr, nil)
	}
	return c.Next()
}
