package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sub-service/internal/models"
	"sub-service/internal/repository"
	"sub-service/pkg/response"
)

type SubscriptionHandler struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionHandler(repo *repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
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

func toMonths(mmYYYY string) int {
	month, _ := strconv.Atoi(mmYYYY[:2])
	year, _ := strconv.Atoi(mmYYYY[3:])

	return year*12 + month
}

func maxMonths(a, b string) string {
	if toMonths(a) > toMonths(b) {
		return a
	}

	return b
}

func minMonths(a, b string) string {
	if toMonths(a) < toMonths(b) {
		return a
	}

	return b
}

func overlapMonths(subStart, subEnd, qStart, qEnd string) int {
	oStart := maxMonths(subStart, qStart)
	oEnd := minMonths(subEnd, qEnd)

	sM := toMonths(oStart)
	eM := toMonths(oEnd)

	if eM < sM {
		return 0
	}

	return eM - sM + 1
}

var dateRegex = regexp.MustCompile(`^(0[1-9]|1[0-2])-[0-9]{4}$`)

// Create godoc
// @Summary Create a new subscription
// @Description Create a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req models.CreateSubscriptionRequest
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
		response.ValidationError(c, "invalid user_id format, must be UUID")
		return
	}

	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.repo.Create(c.Request.Context(), sub); err != nil {
		response.InternalError(c, "failed to create subscription")
		return
	}

	response.Created(c, sub)
}

// GetAll godoc
// @Summary Get all subscriptions
// @Description Get list of all subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {object} map[string][]models.Subscription
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) GetAll(c *gin.Context) {
	subs, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get subscriptions")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"subscriptions": subs})
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Get a single subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	sub, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to get subscription")
		return
	}

	if sub == nil {
		response.NotFound(c, "subscription not found")
		return
	}

	response.Success(c, http.StatusOK, sub)
}

// Update godoc
// @Summary Update subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID"
// @Param subscription body models.UpdateSubscriptionRequest true "Update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	var req models.UpdateSubscriptionRequest
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

	sub, err := h.repo.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalError(c, "failed to update subscription")
		return
	}

	if sub == nil {
		response.NotFound(c, "subscription not found")
		return
	}

	response.Success(c, http.StatusOK, sub)
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ValidationError(c, "invalid subscription id")
		return
	}

	err = h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
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
// @Success 200 {object} models.AggregateResponse
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

	if toMonths(startDate) > toMonths(endDate) {
		response.ValidationError(c, "start_date must be before or equal to end_date")
		return
	}

	var userID *uuid.UUID

	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			response.ValidationError(c, "invalid user_id format")
			return
		}

		userID = &uid
	}

	var serviceName *string

	sn := c.Query("service_name")
	if sn != "" {
		serviceName = &sn
	}

	subs, err := h.repo.GetSubscriptionsByPeriod(c.Request.Context(), startDate, endDate, userID, serviceName)
	if err != nil {
		response.InternalError(c, "failed to get subscriptions")
		return
	}

	totalCost := 0
	for _, sub := range subs {
		subEnd := endDate
		if sub.EndDate != nil {
			subEnd = *sub.EndDate
		}

		overlap := overlapMonths(sub.StartDate, subEnd, startDate, endDate)
		totalCost += sub.Price * overlap
	}

	resp := models.AggregateResponse{
		TotalCost:     totalCost,
		PeriodStart:   startDate,
		PeriodEnd:     endDate,
		Subscriptions: subs,
	}

	response.Success(c, http.StatusOK, resp)
}

// GetByUserID godoc
// @Summary Get subscriptions by user ID
// @Description Get all subscriptions for a specific user with total cost aggregation
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User UUID"
// @Success 200 {object} models.UserSubscriptionsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscriptions/user/{user_id} [get]
func (h *SubscriptionHandler) GetByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.ValidationError(c, "invalid user_id format")
		return
	}

	subs, err := h.repo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to get subscriptions")
		return
	}

	totalCost, err := h.repo.GetTotalCostByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to calculate total cost")
		return
	}

	resp := models.UserSubscriptionsResponse{
		UserID:        userID,
		TotalCost:     totalCost,
		Subscriptions: subs,
	}

	response.Success(c, http.StatusOK, resp)
}
