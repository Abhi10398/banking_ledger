package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang/glog"

	"awesomeProject/logger"
)

func GetPanicRecoveryMiddleware() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			msg := fmt.Sprintf("Panic occurred in service API: \n%v\n%s", e, string(debug.Stack()))
			logger.Get(c.Context()).Error(errors.New("middleware-panic-recover"), msg)
			glog.Flush()
		},
	})
}
