package handler

import (
	"time"

	"github.com/google/uuid"

	"sub-service/internal/model"
)

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserSubscriptionsResponse struct {
	UserID        uuid.UUID              `json:"user_id"`
	TotalCost     int                    `json:"total_cost"`
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type AggregateResponse struct {
	TotalCost     int                    `json:"total_cost"`
	PeriodStart   string                 `json:"period_start"`
	PeriodEnd     string                 `json:"period_end"`
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

func toSubscriptionResponse(s model.Subscription) SubscriptionResponse {
	r := SubscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   model.FormatMonth(s.StartDate),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}

	if s.EndDate != nil {
		endDate := model.FormatMonth(*s.EndDate)
		r.EndDate = &endDate
	}

	return r
}

func toSubscriptionResponses(subs []model.Subscription) []SubscriptionResponse {
	resp := make([]SubscriptionResponse, len(subs))
	for i, s := range subs {
		resp[i] = toSubscriptionResponse(s)
	}

	return resp
}
