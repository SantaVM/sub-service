package service

import (
	"context"
	"log/slog"

	"sub-service/internal/model"
	"sub-service/internal/repository"
)

type Repository interface {
	Create(ctx context.Context, input model.CreateSubscription) (*model.Subscription, error)
	GetByID(ctx context.Context, id uint) (*model.Subscription, error)
	List(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error)
	Update(ctx context.Context, id uint, input model.UpdateSubscription) (*model.Subscription, error)
	Delete(ctx context.Context, id uint) error
	GetTotalCost(ctx context.Context, query model.TotalCostReq) (int, error)
}

var _ Repository = (*repository.SubscriptionRepository)(nil)

type SubscriptionService struct {
	repo   Repository
	logger *slog.Logger
}

func New(repo *repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, input model.CreateSubscription) (*model.Subscription, error) {
	const op = "SubscriptionService.CreateSubscription"
	log := s.logger.With("op", op)
	log.DebugContext(ctx, "creating subscription via service")

	return s.repo.Create(ctx, input)
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uint) (*model.Subscription, error) {
	const op = "SubscriptionService.GetSubscription"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "getting subscription via service")
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
	const op = "SubscriptionService.ListSubscriptions"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "listing subscriptions via service")

	return s.repo.List(ctx, query)
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscription) (*model.Subscription, error) {
	const op = "SubscriptionService.UpdateSubscription"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "updating subscription via service")

	return s.repo.Update(ctx, id, input)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uint) error {
	const op = "SubscriptionService.DeleteSubscription"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "deleting subscription via service")
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, query model.TotalCostReq) (int, error) {
	const op = "SubscriptionService.GetTotalCost"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "calculating total cost via service")

	// business logic
	if query.EndDate == nil {
		query.EndDate = &model.MonthYear{Time: query.StartDate.Time.AddDate(0, 1, 0)}
	}

	return s.repo.GetTotalCost(ctx, query)
}
