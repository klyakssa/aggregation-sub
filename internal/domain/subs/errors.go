package subs

import "errors"

var (
	ErrFailedToFoundSubscription  = errors.New("subscription not found")
	ErrFailedToCreateSubscription = errors.New("failed to create subscription")
	ErrFailedToGetSubscription    = errors.New("failed to get subscription")
	ErrFailedToUpdateSubscription = errors.New("failed to update subscription")
	ErrFailedToDeleteSubscription = errors.New("failed to delete subscription")
	ErrFailedToListSubscriptions  = errors.New("failed to list subscriptions")
	ErrFailedToGetRowsAffected    = errors.New("failed to get rows affected")
	ErrFailedToCountSubscriptions = errors.New("failed to count subscriptions")
	ErrFailedToCalculateTotalCost = errors.New("failed to calculate total cost")
	ErrUserIDIsNil                = errors.New("user_id is nil")
)
