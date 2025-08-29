// backend/internal/api/insurance_handler.go

package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jjckrbbt/catalyst/backend/internal/apps/insurance"
	"github.com/labstack/echo/v4"
)

// InsuranceHandler is the handler for our new insurance application module.
type InsuranceHandler struct {
	queries *insurance.Queries
	logger  *slog.Logger
}

// NewInsuranceHandler creates a new instance of the InsuranceHandler.
func NewInsuranceHandler(q *insurance.Queries, logger *slog.Logger) *InsuranceHandler {
	return &InsuranceHandler{
		queries: q,
		logger:  logger.With("component", "insurance_handler"),
	}
}

// HandleListClaims retrieves a paginated and filtered list of insurance claims.
func (h *InsuranceHandler) HandleListClaims(c echo.Context) error {
	ctx := c.Request().Context()

	// --- Parse Pagination and Filtering Parameters ---
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 32)
	if limit <= 0 {
		limit = 50 // Default limit
	}

	page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 32)
	if page <= 0 {
		page = 1 // Default page
	}

	offset := (page - 1) * limit

	// --- Build the Params Struct for sqlc ---
	params := insurance.ListClaimsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		AdjusterAssigned: pgtype.Text{String: c.QueryParam("adjuster_assigned"), Valid: c.QueryParam("adjuster_assigned") != ""},
		Status:           pgtype.Text{String: c.QueryParam("status"), Valid: c.QueryParam("status") != ""},
		PolicyNumber:     pgtype.Text{String: c.QueryParam("policy_number"), Valid: c.QueryParam("policy_number") != ""},
	}

	// --- Execute the Query ---
	claims, err := h.queries.ListClaims(ctx, params)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list insurance claims", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve claims")
	}

	h.logger.InfoContext(ctx, "Successfully retrieved claims list", "count", len(claims))
	return c.JSON(http.StatusOK, claims)
}

// HandleListPolicyholders retrieves a paginated and filtered list of policyholders.
func (h *InsuranceHandler) HandleListPolicyholders(c echo.Context) error {
    ctx := c.Request().Context()

    limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 32)
    if limit <= 0 {
        limit = 50
    }
    page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 32)
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * limit

    params := insurance.ListPolicyholdersParams{
        Limit:          int32(limit),
        Offset:         int32(offset),
        State:          pgtype.Text{String: c.QueryParam("state"), Valid: c.QueryParam("state") != ""},
        CustomerLevel:  pgtype.Text{String: c.QueryParam("customer_level"), Valid: c.QueryParam("customer_level") != ""},
    }

    policyholders, err := h.queries.ListPolicyholders(ctx, params)
    if err != nil {
        h.logger.ErrorContext(ctx, "Failed to list policyholders", "error", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve policyholders")
    }

    h.logger.InfoContext(ctx, "Successfully retrieved policyholders list", "count", len(policyholders))
    return c.JSON(http.StatusOK, policyholders)
}
