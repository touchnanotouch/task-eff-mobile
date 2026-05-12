package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Price       int       `json:"price" db:"price"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	StartDate   string    `json:"start_date" db:"start_date"`
	EndDate     *string   `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required,min=1,max=255"`
	Price       int     `json:"price" binding:"required,min=0"`
	UserID      string  `json:"user_id" binding:"required,uuid"`
	StartDate   string  `json:"start_date" binding:"required,len=7"`
	EndDate     *string `json:"end_date,omitempty" binding:"omitempty,len=7"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" binding:"omitempty,min=1,max=255"`
	Price       *int    `json:"price,omitempty" binding:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty" binding:"omitempty,len=7"`
	EndDate     *string `json:"end_date,omitempty" binding:"omitempty,len=7"`
}

type UserSubscriptionsResponse struct {
	UserID        uuid.UUID      `json:"user_id"`
	TotalCost     int            `json:"total_cost"`
	Subscriptions []Subscription `json:"subscriptions"`
}

type AggregateResponse struct {
	TotalCost     int            `json:"total_cost"`
	PeriodStart   string         `json:"period_start"`
	PeriodEnd     string         `json:"period_end"`
	Subscriptions []Subscription `json:"subscriptions"`
}
