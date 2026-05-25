package subs

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int        `json:"_"`
	ServiceName string     `json:"service_name" binding:"required"`
	Price       int        `json:"price" binding:"required,gte=0"`
	UserID      uuid.UUID  `json:"user_id" binding:"required,uuid"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSub struct {
	ServiceName string     `json:"service_name" binding:"required"`
	Price       int        `json:"price" binding:"required,gte=0"`
	UserID      uuid.UUID  `json:"user_id" binding:"required,uuid"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type UpdateSub struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty" binding:"omitempty,gte=0"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}
