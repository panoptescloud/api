package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func (srv *Server) AccessLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			opID := c.Response().Header().Get("X-Operation-Id")
			apiKeyID := c.Response().Header().Get("X-Api-Key-Id")
			userID := c.Response().Header().Get("X-User-Id")
			authMethod := c.Response().Header().Get("X-Auth-Method")
			requestID := c.Response().Header().Get("X-Request-Id")

			
			if err != nil {
				c.Error(err)
			}

			res := c.Response()
			fields := []any{
				slog.String("request_id", requestID),
				slog.String("operation_id", opID),
				slog.Int("status", res.Status),
				slog.String("method", c.Request().Method),
				slog.String("uri", c.Request().RequestURI),
				slog.String("remote_ip", c.RealIP()),
				slog.Duration("latency", time.Since(start)),
				slog.String("api_key_id", apiKeyID),
				slog.String("user_id", userID),
				slog.String("auth_method", authMethod),

			}

			srv.accessLogger.Log(context.TODO(), slog.LevelInfo, "", fields...)
			return nil
		}
	}
}
