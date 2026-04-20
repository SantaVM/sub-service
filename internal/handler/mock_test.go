package handler

import (
	"context"
	"sub-service/internal/model"
)

type mockService struct {
	CreateSubscriptionFn func(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error)
	GetSubscriptionFn    func(ctx context.Context, id uint) (*model.Subscription, error)
	ListSubscriptionsFn  func(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error)
	UpdateSubscriptionFn func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error)
	DeleteSubscriptionFn func(ctx context.Context, id uint) error
	GetTotalCostFn       func(ctx context.Context, query model.TotalCostQuery) (int, error)
}

var _ Service = (*mockService)(nil)

func (m *mockService) CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
	if m.CreateSubscriptionFn != nil {
		return m.CreateSubscriptionFn(ctx, input)
	}
	panic("CreateSubscription not implemented")
}

func (m *mockService) GetSubscription(ctx context.Context, id uint) (*model.Subscription, error) {
	if m.GetSubscriptionFn != nil {
		return m.GetSubscriptionFn(ctx, id)
	}
	panic("GetSubscription not implemented")
}

func (m *mockService) ListSubscriptions(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
	if m.ListSubscriptionsFn != nil {
		return m.ListSubscriptionsFn(ctx, query)
	}
	panic("ListSubscriptions not implemented")
}

func (m *mockService) UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
	if m.UpdateSubscriptionFn != nil {
		return m.UpdateSubscriptionFn(ctx, id, input)
	}
	panic("UpdateSubscription not implemented")
}

func (m *mockService) DeleteSubscription(ctx context.Context, id uint) error {
	if m.DeleteSubscriptionFn != nil {
		return m.DeleteSubscriptionFn(ctx, id)
	}
	panic("DeleteSubscription not implemented")
}

func (m *mockService) GetTotalCost(ctx context.Context, query model.TotalCostQuery) (int, error) {
	if m.GetTotalCostFn != nil {
		return m.GetTotalCostFn(ctx, query)
	}
	panic("GetTotalCost not implemented")
}
