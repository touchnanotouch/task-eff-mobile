package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"sub-service/internal/model"
)

type Store interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetAll(ctx context.Context, limit, offset int) ([]model.Subscription, error)
	CountAll(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByPeriod(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Subscription, error)
	GetAggregatedCost(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int, error)
}

type SubscriptionService struct {
	store Store
}

func NewSubscriptionService(store Store) *SubscriptionService {
	return &SubscriptionService{store: store}
}

func (s *SubscriptionService) Create(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	price int,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	start, err := model.ParseMonth(startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	var end *time.Time

	if endDate != nil {
		parsed, err := model.ParseMonth(*endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}

		if !parsed.After(start) {
			return nil, fmt.Errorf("end_date must be after start_date")
		}

		end = &parsed
	}

	sub := &model.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}

	if err := s.store.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionService) GetAll(ctx context.Context, page, limit int) ([]model.Subscription, int, error) {
	offset := (page - 1) * limit

	total, err := s.store.CountAll(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	subs, err := s.store.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subs, total, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		return nil, model.ErrNotFound
	}

	return sub, nil
}

func (s *SubscriptionService) Update(
	ctx context.Context,
	id uuid.UUID,
	serviceName *string,
	price *int,
	startDate,
	endDate *string,
) (*model.Subscription, error) {
	sub, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		return nil, model.ErrNotFound
	}

	if serviceName != nil {
		sub.ServiceName = *serviceName
	}

	if price != nil {
		sub.Price = *price
	}

	if startDate != nil {
		parsed, err := model.ParseMonth(*startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}

		sub.StartDate = parsed
	}

	if endDate != nil {
		if *endDate == "" {
			sub.EndDate = nil
		} else {
			parsed, err := model.ParseMonth(*endDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date: %w", err)
			}

			sub.EndDate = &parsed
		}
	}

	if sub.EndDate != nil && !sub.EndDate.After(sub.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	if err := s.store.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Aggregate(
	ctx context.Context,
	startDate,
	endDate string,
	userID *uuid.UUID,
	serviceName *string,
) (totalCost int, subs []model.Subscription, periodStart, periodEnd string, err error) {
	start, err := model.ParseMonth(startDate)
	if err != nil {
		return 0, nil, "", "", fmt.Errorf("invalid start_date: %w", err)
	}

	end, err := model.ParseMonth(endDate)
	if err != nil {
		return 0, nil, "", "", fmt.Errorf("invalid end_date: %w", err)
	}

	if end.Before(start) {
		return 0, nil, "", "", fmt.Errorf("start_date must be before or equal to end_date")
	}

	totalCost, err = s.store.GetAggregatedCost(ctx, start, end, userID, serviceName)
	if err != nil {
		return 0, nil, "", "", fmt.Errorf("failed to calculate aggregated cost: %w", err)
	}

	subs, err = s.store.GetByPeriod(ctx, start, end, userID, serviceName)
	if err != nil {
		return 0, nil, "", "", fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return totalCost, subs, startDate, endDate, nil
}

func (s *SubscriptionService) GetByUserID(ctx context.Context, userID uuid.UUID) (int, []model.Subscription, error) {
	subs, err := s.store.GetByUserID(ctx, userID)

	if err != nil {
		return 0, nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	totalCost := 0
	for _, sub := range subs {
		totalCost += sub.Price
	}

	return totalCost, subs, nil
}
