package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/klyakssa/aggregation-sub/internal/domain/subs"
)

type SubsService struct {
	repo subs.Repository
}

func NewSubsService(repo subs.Repository) *SubsService {
	return &SubsService{repo: repo}
}

func (s *SubsService) CreateSubscription(ctx context.Context, subs subs.Subscription) (subs.Subscription, error) {
	return s.repo.CreateSubscription(ctx, subs)
}

func (s *SubsService) GetSubscription(ctx context.Context, id int) (subs.Subscription, error) {
	return s.repo.GetSubscription(ctx, id)
}

func (s *SubsService) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]subs.Subscription, error) {
	return s.repo.GetUserSubscriptions(ctx, userID)
}

func (s *SubsService) UpdateSubscription(ctx context.Context, subs subs.Subscription) (subs.Subscription, error) {
	return s.repo.UpdateSubscription(ctx, subs)
}

func (s *SubsService) DeleteSubscription(ctx context.Context, id int) error {
	return s.repo.DeleteSubscription(ctx, id)
}

func (s *SubsService) ListSubscriptions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]subs.Subscription, int, error) {
	return s.repo.ListSubscriptions(ctx, userID, page, pageSize)
}

func (s *SubsService) GetTotalCostByPeriod(ctx context.Context, cost subs.CostCalculation) (int, error) {
	return s.repo.GetTotalCostByPeriod(ctx, cost)
}
