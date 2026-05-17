package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"sub-service/internal/model"
)

type mockStore struct {
	createFunc        func(ctx context.Context, sub *model.Subscription) error
	getByIDFunc       func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	getAggregatedCost func(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int, error)
	getByPeriodFunc   func(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error)
}

func (m *mockStore) Create(ctx context.Context, sub *model.Subscription) error {
	return m.createFunc(ctx, sub)
}

func (m *mockStore) GetAll(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	return nil, nil
}

func (m *mockStore) CountAll(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockStore) Update(ctx context.Context, sub *model.Subscription) error {
	return nil
}

func (m *mockStore) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockStore) GetByPeriod(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error) {
	return m.getByPeriodFunc(ctx, start, end, userID, serviceName)
}

func (m *mockStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Subscription, error) {
	return nil, nil
}

func (m *mockStore) GetAggregatedCost(ctx context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	return m.getAggregatedCost(ctx, start, end, userID, serviceName)
}

func TestCreate_Valid(t *testing.T) {
	store := &mockStore{
		createFunc: func(_ context.Context, sub *model.Subscription) error {
			sub.ID = uuid.New()
			return nil
		},
	}

	svc := NewSubscriptionService(store)

	userID := uuid.New()

	sub, err := svc.Create(context.Background(), userID, "Netflix", 500, "01-2026", nil)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if sub.ServiceName != "Netflix" {
		t.Errorf("ServiceName = %s, want Netflix", sub.ServiceName)
	}

	if sub.Price != 500 {
		t.Errorf("Price = %d, want 500", sub.Price)
	}

	if sub.UserID != userID {
		t.Errorf("UserID mismatch")
	}
}

func TestCreate_InvalidDate(t *testing.T) {
	svc := NewSubscriptionService(&mockStore{})

	_, err := svc.Create(context.Background(), uuid.New(), "Test", 100, "invalid-date", nil)
	if err == nil {
		t.Fatal("Create() expected error for invalid date, got nil")
	}
}

func TestCreate_EndDateBeforeStartDate(t *testing.T) {
	svc := NewSubscriptionService(&mockStore{})

	endDate := "01-2025"

	_, err := svc.Create(context.Background(), uuid.New(), "Test", 100, "06-2025", &endDate)
	if err == nil {
		t.Fatal("Create() expected error for end_date before start_date, got nil")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	store := &mockStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Subscription, error) {
			return nil, nil
		},
	}

	svc := NewSubscriptionService(store)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetByID() expected error for not found, got nil")
	}
}

func TestAggregate_Valid(t *testing.T) {
	store := &mockStore{
		getAggregatedCost: func(_ context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
			return 1200, nil
		},
		getByPeriodFunc: func(_ context.Context, start, end time.Time, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error) {
			return []model.Subscription{}, nil
		},
	}

	svc := NewSubscriptionService(store)

	totalCost, subs, start, end, err := svc.Aggregate(context.Background(), "01-2026", "03-2026", nil, nil)
	if err != nil {
		t.Fatalf("Aggregate() unexpected error: %v", err)
	}

	if totalCost != 1200 {
		t.Errorf("totalCost = %d, want 1200", totalCost)
	}

	if subs == nil {
		t.Error("subs should not be nil")
	}

	if start != "01-2026" {
		t.Errorf("start = %s, want 01-2026", start)
	}

	if end != "03-2026" {
		t.Errorf("end = %s, want 03-2026", end)
	}
}

func TestAggregate_InvalidPeriod(t *testing.T) {
	svc := NewSubscriptionService(&mockStore{})

	_, _, _, _, err := svc.Aggregate(context.Background(), "03-2026", "01-2026", nil, nil)
	if err == nil {
		t.Fatal("Aggregate() expected error for end < start, got nil")
	}
}
