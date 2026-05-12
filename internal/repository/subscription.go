package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"sub-service/internal/config"
	"sub-service/internal/models"
)

type SubscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(cfg config.DatabaseConfig) (*SubscriptionRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &SubscriptionRepository{db: db}, nil
}

func (r *SubscriptionRepository) Close() error {
	return r.db.Close()
}

func (r *SubscriptionRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	sub.ID = uuid.New()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	query := `SELECT * FROM subscriptions WHERE id = $1`

	err := r.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]models.Subscription, error) {
	var subs []models.Subscription
	query := `SELECT * FROM subscriptions ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &subs, query)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	sub, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		sub.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		sub.EndDate = req.EndDate
	}
	sub.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions 
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = $5
		WHERE id = $6
	`

	_, err = r.db.ExecContext(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
		sub.UpdatedAt,
		id,
	)

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	var subs []models.Subscription
	query := `SELECT * FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &subs, query, userID)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (r *SubscriptionRepository) GetTotalCostByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var total sql.NullInt64
	query := `SELECT SUM(price) FROM subscriptions WHERE user_id = $1`

	err := r.db.GetContext(ctx, &total, query, userID)
	if err != nil {
		return 0, err
	}

	if !total.Valid {
		return 0, nil
	}

	return int(total.Int64), nil
}
