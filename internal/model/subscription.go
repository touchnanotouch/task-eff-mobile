package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Price       int        `db:"price" json:"price"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	StartDate   time.Time  `db:"start_date" json:"-"`
	EndDate     *time.Time `db:"end_date" json:"-"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func ParseMonth(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}

func FormatMonth(t time.Time) string {
	return t.Format("01-2006")
}
