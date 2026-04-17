package service

import (
	"context"
	"log/slog"

	"sub-service/internal/model"
	"sub-service/internal/repository"
)

type SubscriptionService struct {
	repo   *repository.SubscriptionRepository
	logger *slog.Logger
}

func New(repo *repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

// TODO: move to handler
func (s *SubscriptionService) CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
	s.logger.InfoContext(ctx, "creating subscription via service", "service_name", input.ServiceName)

	subsciption, err := input.ToDomain()
	if err != nil {
		s.logger.ErrorContext(ctx, "error convertion to domain", "create_sub", input.UserID)
		return nil, err
	}

	return s.repo.Create(ctx, *subsciption)
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uint) (*model.Subscription, error) {
	s.logger.InfoContext(ctx, "getting subscription via service", "subscription_id", id)
	return s.repo.GetByID(ctx, id)
}

// []*model.Subscription
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
	s.logger.InfoContext(ctx, "listing subscriptions via service")

	// Установка значений по умолчанию
	if query.Size <= 0 {
		query.Size = 10
	}

	if query.Size > 100 {
		query.Size = 100
	}

	if query.Page < 1 {
		query.Page = 1
	}

	return s.repo.List(ctx, query)
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
	s.logger.InfoContext(ctx, "updating subscription via service", "subscription_id", id)

	update, err := input.ToDomain()
	if err != nil {
		s.logger.ErrorContext(ctx, "error convertion to domain", "update_sub", id)
		return nil, err
	}

	return s.repo.Update(ctx, id, *update)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uint) error {
	s.logger.InfoContext(ctx, "deleting subscription via service", "subscription_id", id)
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, query model.TotalCostQuery) (int, error) {
	s.logger.InfoContext(ctx, "calculating total cost via service")
	return s.repo.GetTotalCost(ctx, query)
}
