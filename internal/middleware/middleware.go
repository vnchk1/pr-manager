package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func LoggingMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { //nolint:varnamelen
			start := time.Now()
			err := next(c)

			defer func() {
				latency := time.Since(start)
				logger.Info("completed request",
					"method", c.Request().Method,
					"path", c.Request().URL.Path,
					"status", c.Response().Status,
					"latency_ms", latency.Milliseconds(),
					"ip", c.RealIP())
			}()

			return err
		}
	}
}
