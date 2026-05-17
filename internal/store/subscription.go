package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"sub-service/internal/config"
	"sub-service/internal/model"
)

type SubscriptionStore struct {
	db *sqlx.DB
}

const selectCols = `id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

func NewSubscriptionStore(cfg config.DatabaseConfig) (*SubscriptionStore, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &SubscriptionStore{db: db}, nil
}

func (s *SubscriptionStore) Close() error {
	return s.db.Close()
}

func (s *SubscriptionStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SubscriptionStore) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()

	sub.ID = uuid.New()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx, query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID,
		sub.StartDate, sub.EndDate, sub.CreatedAt, sub.UpdatedAt,
	)

	return err
}

func (s *SubscriptionStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription

	query := `SELECT ` + selectCols + ` FROM subscriptions WHERE id = $1`

	err := s.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &sub, nil
}

func (s *SubscriptionStore) GetAll(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	subs := make([]model.Subscription, 0)

	query := `SELECT ` + selectCols + ` FROM subscriptions ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	err := s.db.SelectContext(ctx, &subs, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionStore) CountAll(ctx context.Context) (int, error) {
	var total int

	err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM subscriptions`)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *SubscriptionStore) Update(ctx context.Context, sub *model.Subscription) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	var existing model.Subscription

	err = tx.GetContext(
		ctx, &existing,
		`SELECT `+selectCols+` FROM subscriptions WHERE id = $1 FOR UPDATE`,
		sub.ID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}

		return err
	}

	sub.UpdatedAt = time.Now()

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = $5
		WHERE id = $6
		`,
		sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.UpdatedAt, sub.ID,
	)

	if err != nil {
		return fmt.Errorf("update exec: %w", err)
	}

	return tx.Commit()
}

func (s *SubscriptionStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (s *SubscriptionStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Subscription, error) {
	subs := make([]model.Subscription, 0)

	query := `SELECT ` + selectCols + ` FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`

	err := s.db.SelectContext(ctx, &subs, query, userID)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionStore) GetByPeriod(
	ctx context.Context,
	start,
	end time.Time,
	userID *uuid.UUID,
	serviceName *string,
) ([]model.Subscription, error) {
	subs := make([]model.Subscription, 0)

	query := `SELECT ` + selectCols + ` FROM subscriptions
		WHERE start_date <= :query_end AND (end_date IS NULL OR end_date >= :query_start)`

	params := map[string]any{
		"query_start": start,
		"query_end":   end,
	}

	if userID != nil {
		query += ` AND user_id = :user_id`
		params["user_id"] = *userID
	}

	if serviceName != nil {
		query += ` AND service_name = :service_name`
		params["service_name"] = *serviceName
	}

	query += ` ORDER BY created_at DESC`

	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, err
	}

	err = s.db.SelectContext(ctx, &subs, s.db.Rebind(namedQuery), args...)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionStore) GetAggregatedCost(
	ctx context.Context,
	start,
	end time.Time,
	userID *uuid.UUID,
	serviceName *string,
) (int, error) {
	query := `
		SELECT COALESCE(SUM(price * (
			(EXTRACT(YEAR FROM eff_end) - EXTRACT(YEAR FROM eff_start)) * 12
			+ EXTRACT(MONTH FROM eff_end) - EXTRACT(MONTH FROM eff_start)
			+ 1
		)), 0)
		FROM (
			SELECT s.price,
			       GREATEST(s.start_date, CAST(:query_start AS date)) AS eff_start,
			       LEAST(COALESCE(s.end_date, CAST(:query_end AS date)), CAST(:query_end AS date)) AS eff_end
			FROM subscriptions s
			WHERE s.start_date <= CAST(:query_end AS date)
			  AND (s.end_date IS NULL OR s.end_date >= CAST(:query_start AS date))
	`

	params := map[string]any{
		"query_start": start,
		"query_end":   end,
	}

	if userID != nil {
		query += ` AND s.user_id = :user_id`
		params["user_id"] = *userID
	}

	if serviceName != nil {
		query += ` AND s.service_name = :service_name`
		params["service_name"] = *serviceName
	}

	query += `
		) calc
		WHERE calc.eff_start <= calc.eff_end
	`

	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return 0, err
	}

	var total sql.NullInt64

	err = s.db.GetContext(ctx, &total, s.db.Rebind(namedQuery), args...)
	if err != nil {
		return 0, err
	}

	if !total.Valid {
		return 0, nil
	}

	return int(total.Int64), nil
}
