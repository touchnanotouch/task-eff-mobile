package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sub-service/internal/model"
	"sub-service/internal/service"
	"sub-service/pkg/response"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

var dateRegex = regexp.MustCompile(`^(0[1-9]|1[0-2])-[0-9]{4}$`)

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) RegisterRoutes(r *gin.RouterGroup) {
	subs := r.Group("/subscriptions")

	{
		subs.POST("", h.Create)
		subs.GET("", h.GetAll)
		subs.GET("/aggregate", h.Aggregate)
		subs.GET("/:id", h.GetByID)
		subs.PUT("/:id", h.Update)
		subs.DELETE("/:id", h.Delete)
		subs.GET("/user/:user_id", h.GetByUserID)
	}
}

// Create godoc
// @Summary Create a new subscription
// @Description Create a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body handler.CreateRequest true "Subscription data"
// @Success 201 {object} handler.SubscriptionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req CreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if !dateRegex.MatchString(req.StartDate) {
		response.ValidationError(c, "start_date must be in MM-YYYY format")
		return
	}

	if req.EndDate != nil && !dateRegex.MatchString(*req.EndDate) {
		response.ValidationError(c, "end_date must be in MM-YYYY format")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.ValidationError(c, "invalid user_id format")
		return
	}

	sub, err := h.svc.Create(
		c.Request.Context(), userID, req.ServiceName, req.Price, req.StartDate, req.EndDate,
	)

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, toSubscriptionResponse(*sub))
}

// GetAll godoc
// @Summary Get all subscriptions
// @Description Get paginated list of all subscriptions
// @Tags subscriptions
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) GetAll(c *gin.Context) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	subs, total, err := h.svc.GetAll(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "failed to get subscriptions")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"subscriptions": toSubscriptionResponses(subs),
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Get a single subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID"
// @Success 200 {object} handler.SubscriptionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	sub, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			response.NotFound(c, "subscription not found")
			return
		}
		response.InternalError(c, "failed to get subscription")
		return
	}

	response.Success(c, http.StatusOK, toSubscriptionResponse(*sub))
}

// Update godoc
// @Summary Update subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID"
// @Param subscription body handler.UpdateRequest true "Update data"
// @Success 200 {object} handler.SubscriptionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	var req UpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if req.StartDate != nil && !dateRegex.MatchString(*req.StartDate) {
		response.ValidationError(c, "start_date must be in MM-YYYY format")
		return
	}

	if req.EndDate != nil && !dateRegex.MatchString(*req.EndDate) {
		response.ValidationError(c, "end_date must be in MM-YYYY format")
		return
	}

	sub, err := h.svc.Update(
		c.Request.Context(), id, req.ServiceName, req.Price, req.StartDate, req.EndDate,
	)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			response.NotFound(c, "subscription not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, toSubscriptionResponse(*sub))
}

// Delete godoc
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription UUID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			response.NotFound(c, "subscription not found")
			return
		}
		response.InternalError(c, "failed to delete subscription")
		return
	}

	response.NoContent(c)
}

// Aggregate godoc
// @Summary Aggregate subscription costs over a period
// @Description Calculate total cost of all subscriptions active during the specified period, with optional filters
// @Tags subscriptions
// @Produce json
// @Param start_date query string true "Start of period (MM-YYYY)"
// @Param end_date query string true "End of period (MM-YYYY)"
// @Param user_id query string false "Filter by user UUID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {object} handler.AggregateResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/aggregate [get]
func (h *SubscriptionHandler) Aggregate(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		response.ValidationError(c, "start_date and end_date are required")
		return
	}

	if !dateRegex.MatchString(startDate) {
		response.ValidationError(c, "start_date must be in MM-YYYY format")
		return
	}

	if !dateRegex.MatchString(endDate) {
		response.ValidationError(c, "end_date must be in MM-YYYY format")
		return
	}

	if monthKey(startDate) > monthKey(endDate) {
		response.ValidationError(c, "start_date must be before or equal to end_date")
		return
	}

	var userID *uuid.UUID

	if uid := c.Query("user_id"); uid != "" {
		parsed, err := uuid.Parse(uid)
		if err != nil {
			response.ValidationError(c, "invalid user_id format")
			return
		}

		userID = &parsed
	}

	var serviceName *string

	if sn := c.Query("service_name"); sn != "" {
		serviceName = &sn
	}

	totalCost, subs, periodStart, periodEnd, err := h.svc.Aggregate(
		c.Request.Context(), startDate, endDate, userID, serviceName,
	)

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	resp := AggregateResponse{
		TotalCost:     totalCost,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Subscriptions: toSubscriptionResponses(subs),
	}

	response.Success(c, http.StatusOK, resp)
}

// GetByUserID godoc
// @Summary Get subscriptions by user ID
// @Description Get all subscriptions for a specific user with total cost
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User UUID"
// @Success 200 {object} handler.UserSubscriptionsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/user/{user_id} [get]
func (h *SubscriptionHandler) GetByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.ValidationError(c, "invalid user_id format")
		return
	}

	totalCost, subs, err := h.svc.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	resp := UserSubscriptionsResponse{
		UserID:        userID,
		TotalCost:     totalCost,
		Subscriptions: toSubscriptionResponses(subs),
	}

	response.Success(c, http.StatusOK, resp)
}

func monthKey(s string) string {
	return s[3:7] + s[0:2]
}
