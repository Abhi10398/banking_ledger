package middleware

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"

	"awesomeProject/internal/constants"
)

const RequestBodyPermissibleLength = 3000

func GetLogMiddleWare() fiber.Handler {
	return LogMiddleware
}

// LogMiddleware godoc
func LogMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	zapLogger, _ := zap.NewProduction()

	defer zapLogger.Sync()
	var requestBody, responseBody []byte
	if _, err := c.MultipartForm(); err != nil {
		requestBody = c.Body()
	}
	if c.Response().StatusCode() >= http.StatusBadRequest {
		responseBody = c.Response().Body()
	}

	truncatedRequestBody := truncateBytes(requestBody, RequestBodyPermissibleLength)

	zapLogger.Info("Request",
		zap.Int("status_code", c.Response().StatusCode()),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Any("query_params", getQueryParamsToLog(c)),
		zap.String("latency", time.Since(start).String()),
		zap.Time("time", time.Now()),
		zap.String("host", c.Hostname()),
		zap.Any("req_headers", getHeadersToLog(c)),
		zap.ByteString("request_body", truncatedRequestBody),
		zap.ByteString("response_body", responseBody),
		zap.Any("correlationId", c.UserContext().Value(constants.CorrelationId)),
	)
	return err
}

func getQueryParamsToLog(c *fiber.Ctx) map[string]string {
	allQueryParams := make(map[string]string, c.Context().QueryArgs().Len())
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		allQueryParams[string(key)] = string(value)
	})
	return allQueryParams
}

func getHeadersToLog(c *fiber.Ctx) map[string][]string {
	allHeaders := c.GetReqHeaders()
	headersToLog := map[string]string{}
	headersNotToLog := map[string]string{
		"Authorization": "Authorization",
	}
	headers := map[string][]string{}

	if len(headersToLog) == 0 {
		headers = allHeaders
	} else {
		for _, value := range headersToLog {
			headers[value] = allHeaders[value]
		}
	}

	if len(headersNotToLog) != 0 {
		for _, value := range headersNotToLog {
			delete(headers, value)
		}
	}
	return headers
}

func truncateBytes(data []byte, length int) []byte {
	if len(data) > length {
		return append(data[:length], []byte("...")...)
	}
	return data
}
