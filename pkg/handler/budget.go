package handler

import (
	"context"
	"net/http"
	"strconv"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/service"

	"github.com/gin-gonic/gin"
)

type budgetService interface {
	List(ctx context.Context, userID int64) ([]models.Budget, *apperror.Error)
	Create(ctx context.Context, userID int64, req models.CreateBudgetRequest) (*models.Budget, *apperror.Error)
	GetProgress(ctx context.Context, userID, id int64) (*models.BudgetProgress, *apperror.Error)
	Update(ctx context.Context, userID, id int64, req models.UpdateBudgetRequest) (*models.Budget, *apperror.Error)
	Delete(ctx context.Context, userID, id int64) *apperror.Error
}

type BudgetHandler struct {
	svc budgetService
}

func NewBudgetHandler(svc *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{svc: svc}
}

// List godoc
// @Summary List budgets
// @Description List all budgets for the authenticated user.
// @Tags budgets
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Budget
// @Failure 401 {object} apperror.ErrorEnvelope
// @Router /api/v1/budgets [get]
func (h *BudgetHandler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	out, appErr := h.svc.List(c.Request.Context(), userID)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Create godoc
// @Summary Create budget
// @Description Create a new budget for the authenticated user.
// @Tags budgets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateBudgetRequest true "Create budget payload"
// @Success 201 {object} models.Budget
// @Failure 400 {object} apperror.ErrorEnvelope
// @Failure 401 {object} apperror.ErrorEnvelope
// @Router /api/v1/budgets [post]
func (h *BudgetHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	var req models.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.svc.Create(c.Request.Context(), userID, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// GetProgress godoc
// @Summary Get budget progress
// @Description Get spending progress for a specific budget.
// @Tags budgets
// @Security BearerAuth
// @Produce json
// @Param id path int true "Budget ID"
// @Success 200 {object} models.BudgetProgress
// @Failure 401 {object} apperror.ErrorEnvelope
// @Failure 404 {object} apperror.ErrorEnvelope
// @Router /api/v1/budgets/{id}/progress [get]
func (h *BudgetHandler) GetProgress(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid budget id"))
		return
	}
	out, appErr := h.svc.GetProgress(c.Request.Context(), userID, id)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Update godoc
// @Summary Update budget
// @Description Update an existing budget.
// @Tags budgets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Budget ID"
// @Param request body models.UpdateBudgetRequest true "Update budget payload"
// @Success 200 {object} models.Budget
// @Failure 400 {object} apperror.ErrorEnvelope
// @Failure 401 {object} apperror.ErrorEnvelope
// @Failure 404 {object} apperror.ErrorEnvelope
// @Router /api/v1/budgets/{id} [patch]
func (h *BudgetHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid budget id"))
		return
	}
	var req models.UpdateBudgetRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.svc.Update(c.Request.Context(), userID, id, req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Delete godoc
// @Summary Delete budget
// @Description Delete a budget by ID.
// @Tags budgets
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Success 204
// @Failure 401 {object} apperror.ErrorEnvelope
// @Failure 404 {object} apperror.ErrorEnvelope
// @Router /api/v1/budgets/{id} [delete]
func (h *BudgetHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, apperror.Validation("invalid budget id"))
		return
	}
	if appErr := h.svc.Delete(c.Request.Context(), userID, id); appErr != nil {
		writeError(c, appErr)
		return
	}
	c.Status(http.StatusNoContent)
}
