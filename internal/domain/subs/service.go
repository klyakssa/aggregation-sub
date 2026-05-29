package subs

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	CreateSubscription(ctx context.Context, subs Subscription) (Subscription, error)
	GetSubscription(ctx context.Context, id int) (Subscription, error)
	GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	UpdateSubscription(ctx context.Context, subs Subscription) (Subscription, error)
	DeleteSubscription(ctx context.Context, id int) error
	ListSubscriptions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]Subscription, int, error)
	GetTotalCostByPeriod(ctx context.Context, cost CostCalculation) (int, error)
}
