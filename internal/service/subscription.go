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

func (s *SubscriptionService) CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
	const op = "SubscriptionService.CreateSubscription"
	log := s.logger.With("op", op)
	log.DebugContext(ctx, "creating subscription via service")

	subsciption, err := input.ToDomain()
	if err != nil {
		log.ErrorContext(ctx, "error convertion to domain", "error", err)
		return nil, err
	}

	return s.repo.Create(ctx, *subsciption)
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

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
	const op = "SubscriptionService.UpdateSubscription"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "updating subscription via service")

	update, err := input.ToDomain()
	if err != nil {
		log.ErrorContext(ctx, "error convertion to domain", "error", err)
		return nil, err
	}

	return s.repo.Update(ctx, id, *update)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uint) error {
	const op = "SubscriptionService.DeleteSubscription"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "deleting subscription via service")
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, query model.TotalCostQuery) (int, error) {
	const op = "SubscriptionService.GetTotalCost"
	log := s.logger.With("op", op)

	log.DebugContext(ctx, "calculating total cost via service")

	dto, err := query.ToDomain()
	if err != nil {
		log.ErrorContext(ctx, "error convertion to domain", "error", err)
		return -1, err
	}

	return s.repo.GetTotalCost(ctx, *dto)
}
