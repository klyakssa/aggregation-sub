package subs

import (
	"github.com/google/uuid"
)

type SubscriptionService interface {
	CreateSubscription(subs Subscription) (Subscription, error)
	GetSubscription(id int) (Subscription, error)
	GetUserSubscriptions(userID uuid.UUID) ([]Subscription, error)
	UpdateSubscription(subs Subscription) (Subscription, error)
	DeleteSubscription(userID uuid.UUID, serviceName string) error
	ListSubscriptions(userID uuid.UUID, page, pageSize int) ([]Subscription, int, error)
}
