package api

import (
	"log/slog"

	"github.com/jjckrbbt/catalyst/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	// Fields for a JWT validator, etc.
}

func NewAuthMiddleware(domain, audience string, q repository.Querier, logger *slog.Logger) (*AuthMiddleware, error) {
	return &AuthMiddleware{}, nil
}

func (m *AuthMiddleware) ValidateRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Real JWT validation logic goes here.
		return next(c)
	}
}
