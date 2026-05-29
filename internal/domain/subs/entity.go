package subs

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int        `json:"id" db:"id"`
	ServiceName string     `json:"service_name" binding:"required" db:"service_name"`
	Price       int        `json:"price" binding:"required,gte=0" db:"price"`
	UserID      uuid.UUID  `json:"user_id" binding:"required,uuid" db:"user_id"`
	StartDate   time.Time  `json:"start_date" binding:"required" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	UpdatedAt   time.Time  `json:"-" db:"updated_at"`
}

type CreateSub struct {
	ServiceName string     `json:"service_name" binding:"required"`
	Price       int        `json:"price" binding:"required,gte=0"`
	UserID      uuid.UUID  `json:"user_id" binding:"required,uuid"`
	StartDate   MonthYear  `json:"start_date" binding:"required"` // MM-YYYY
	EndDate     *MonthYear `json:"end_date,omitempty"`            // MM-YYYY
}

type UpdateSub struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty" binding:"omitempty,gte=0"`
	StartDate   *MonthYear `json:"start_date,omitempty"` // MM-YYYY
	EndDate     *MonthYear `json:"end_date,omitempty"`   // MM-YYYY
}

type CostCalculation struct {
	UserID      uuid.UUID `form:"user_id" json:"user_id"`
	ServiceName string    `form:"service_name" json:"service_name"`
	StartDate   time.Time `form:"start_date" json:"start_date" binding:"required"`
	EndDate     time.Time `form:"end_date" json:"end_date" binding:"required"`
}

type MonthYear struct {
	time.Time
}

func (my *MonthYear) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" || s == "null" {
		return nil
	}

	t, err := time.Parse("01-2006", s)
	if err != nil {
		return fmt.Errorf("invalid month-year format, expected MM-YYYY: %s", s)
	}

	my.Time = t
	return nil
}

func (my MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", my.Format("01-2006"))), nil
}
