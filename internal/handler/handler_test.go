package handler

import (
	"context"
	"errors"
	"net/http"
	"sub-service/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCreateSubscription(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success",
			setup: func(svc *mockService) {
				svc.CreateSubscriptionFn = func(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
					return &model.Subscription{ID: 1}, nil
				}
			},
			req: NewRequest().
				Method(http.MethodPost).
				Path("/subscriptions").
				JSON(model.CreateSubscriptionInput{
					UserID:      "018f8f6e-0000-0000-0000-000000000000",
					ServiceName: "Netflix",
					Price:       10,
					StartDate:   "01-2026",
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out model.Subscription

				resp.
					Status(http.StatusCreated).
					JSON(&out)

				require.Equal(t, uint(1), out.ID)
			},
		},
		{
			name: "validation error",
			setup: func(svc *mockService) {
				svc.CreateSubscriptionFn = func(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodPost).
				Path("/subscriptions").
				JSON(model.CreateSubscriptionInput{
					UserID:      "wrong-uuid", // wrong
					ServiceName: "N",          // too short
					Price:       -10,          // negative
					StartDate:   "1-2026",     // wrong format
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Errors)
				require.Equal(t, len(out.Errors), 4)
			},
		},
		{
			name: "conflict - create",
			setup: func(svc *mockService) {
				svc.CreateSubscriptionFn = func(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error) {
					return nil, model.ErrSubscriptionOverlap
				}
			},
			req: NewRequest().
				Method(http.MethodPost).
				Path("/subscriptions").
				JSON(model.CreateSubscriptionInput{
					UserID:      "018f8f6e-0000-0000-0000-000000000000",
					ServiceName: "Netflix",
					Price:       10,
					StartDate:   "01-2026",
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusConflict).
					BodyContains("overlap")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.CreateSubscription, tt.req, resp)

			tt.check(t, resp)
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success",
			setup: func(svc *mockService) {
				svc.DeleteSubscriptionFn = func(ctx context.Context, id uint) error {
					require.Equal(t, uint(1), id)
					return nil
				}
			},
			req: NewRequest().
				Method(http.MethodDelete).
				Path("/subscriptions/1").
				URLParam("id", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.Status(http.StatusNoContent)
				require.Empty(t, resp.Rec.Body.String())
			},
		},
		{
			name: "missing id",
			setup: func(svc *mockService) {
				svc.DeleteSubscriptionFn = func(ctx context.Context, id uint) error {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodDelete).
				Path("/subscriptions"). // no parameter
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.Contains(t, out.Error, "missing ID")
			},
		},
		{
			name: "invalid id format",
			setup: func(svc *mockService) {
				svc.DeleteSubscriptionFn = func(ctx context.Context, id uint) error {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodPost).
				Path("/subscriptions/abc").
				URLParam("id", "abc").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.Contains(t, out.Error, "invalid subscription ID format")
			},
		},
		{
			name: "not found",
			setup: func(svc *mockService) {
				svc.DeleteSubscriptionFn = func(ctx context.Context, id uint) error {
					return errors.New("subscription not found")
				}
			},
			req: NewRequest().
				Method(http.MethodDelete).
				Path("/subscriptions/1").
				URLParam("id", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusNotFound).
					JSON(&out)

				require.Contains(t, out.Error, "not found")
			},
		},
		{
			name: "internal error",
			setup: func(svc *mockService) {
				svc.DeleteSubscriptionFn = func(ctx context.Context, id uint) error {
					return errors.New("db is down")
				}
			},
			req: NewRequest().
				Method(http.MethodDelete).
				Path("/subscriptions/1").
				URLParam("id", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusInternalServerError).
					JSON(&out)

				require.Contains(t, out.Error, "failed to delete")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.DeleteSubscription, tt.req, resp)

			tt.check(t, resp)
		})
	}
}

func TestUpdateSubscription(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success - update",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					require.Equal(t, uint(1), id)
					return &model.Subscription{ID: 1}, nil
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(model.UpdateSubscriptionInput{
					ServiceName: ptr("Netflix"),
					Price:       ptr(20),
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out model.Subscription

				resp.
					Status(http.StatusOK).
					JSON(&out)

				require.Equal(t, uint(1), out.ID)
			},
		},
		{
			name:  "missing id",
			setup: func(svc *mockService) {},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions").
				JSON(model.UpdateSubscriptionInput{}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.Contains(t, out.Error, "missing ID")
			},
		},
		{
			name:  "invalid id format",
			setup: func(svc *mockService) {},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/abc").
				URLParam("id", "abc").
				JSON(model.UpdateSubscriptionInput{}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.Contains(t, out.Error, "invalid subscription ID format")
			},
		},
		{
			name: "validation error",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(
					model.UpdateSubscriptionInput{
						StartDate: ptr("03-2000"),
						EndDate:   ptr("01-2000"), // before start_date
					},
				).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Errors)
			},
		},
		{
			name: "conflict - overlap",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					return nil, model.ErrSubscriptionOverlap
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(model.UpdateSubscriptionInput{
					ServiceName: ptr("Netflix"),
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusConflict).
					BodyContains("overlap")
			},
		},
		{
			name: "invalid date range",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					return nil, model.ErrInvalidDateRange
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(model.UpdateSubscriptionInput{
					ServiceName: ptr("Netflix"),
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusBadRequest).
					BodyContains("invalid")
			},
		},
		{
			name: "not found",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					return nil, nil
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(model.UpdateSubscriptionInput{
					ServiceName: ptr("Netflix"),
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusNotFound).
					JSON(&out)

				require.Contains(t, out.Error, "not found")
			},
		},
		{
			name: "internal error",
			setup: func(svc *mockService) {
				svc.UpdateSubscriptionFn = func(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error) {
					return nil, errors.New("db error")
				}
			},
			req: NewRequest().
				Method(http.MethodPut).
				Path("/subscriptions/1").
				URLParam("id", "1").
				JSON(model.UpdateSubscriptionInput{
					ServiceName: ptr("Netflix"),
				}).
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusInternalServerError).
					BodyContains("db error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.UpdateSubscription, tt.req, resp)

			tt.check(t, resp)
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success",
			setup: func(svc *mockService) {
				svc.ListSubscriptionsFn = func(ctx context.Context, q model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
					require.NotNil(t, q.UserID)
					require.Equal(t, "018f8f6e-0000-0000-0000-000000000000", *q.UserID)

					require.NotNil(t, q.ServiceName)
					require.Equal(t, "Netflix", *q.ServiceName)

					require.Equal(t, 10, q.Size)
					require.Equal(t, 1, q.Page)

					return model.NewPage([]*model.Subscription{{ID: uint(1)}}, 10, 1, 1), nil
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions").
				Query("user_id", "018f8f6e-0000-0000-0000-000000000000").
				Query("service_name", "Netflix").
				Query("size", "10").
				Query("page", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out model.Page[*model.Subscription]

				resp.
					Status(http.StatusOK).
					JSON(&out)

				require.Len(t, out.Content, 1)
				require.Equal(t, uint(1), out.Content[0].ID)
			},
		},
		{
			name: "validation error",
			setup: func(svc *mockService) {
				svc.ListSubscriptionsFn = func(ctx context.Context, q model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions").
				Query("user_id", "wrong-uuid"). // wrong format
				Query("size", "-10").           // negative
				Query("page", "-1").            // negative
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Errors)
				require.Equal(t, len(out.Errors), 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.ListSubscriptions, tt.req, resp)

			tt.check(t, resp)
		})
	}
}

func TestGetSubscription(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success",
			setup: func(svc *mockService) {
				svc.GetSubscriptionFn = func(ctx context.Context, id uint) (*model.Subscription, error) {
					require.Equal(t, uint(1), id)

					return &model.Subscription{ID: uint(1)}, nil
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/1").
				URLParam("id", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out model.Subscription

				resp.
					Status(http.StatusOK).
					JSON(&out)

				require.Equal(t, uint(1), out.ID)
			},
		},
		{
			name: "validation error",
			setup: func(svc *mockService) {
				svc.GetSubscriptionFn = func(ctx context.Context, id uint) (*model.Subscription, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/1").
				URLParam("id", "abc"). // wrong parameter
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Error)
				require.Contains(t, out.Error, "format")
			},
		},
		{
			name: "not found",
			setup: func(svc *mockService) {
				svc.GetSubscriptionFn = func(ctx context.Context, id uint) (*model.Subscription, error) {
					require.Equal(t, uint(1), id)

					return nil, nil
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/1").
				URLParam("id", "1").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusNotFound).
					JSON(&out)

				require.NotEmpty(t, out.Error)
				require.Contains(t, out.Error, "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.GetSubscription, tt.req, resp)

			tt.check(t, resp)
		})
	}
}

func TestGetTotalCost(t *testing.T) {
	tests := []struct {
		name  string
		setup func(svc *mockService)
		req   *http.Request
		check func(t *testing.T, resp *Response)
	}{
		{
			name: "success",
			setup: func(svc *mockService) {
				svc.GetTotalCostFn = func(ctx context.Context, query model.TotalCostQuery) (int, error) {
					require.NotNil(t, query)
					require.Equal(t, "01-2000", query.StartDate)

					return 100, nil
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions").
				Query("start_date", "01-2000").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out TotalCostResponse

				resp.
					Status(http.StatusOK).
					JSON(&out)

				require.Equal(t, 100, out.TotalCost)
			},
		},
		{
			name: "missing start_date",
			setup: func(svc *mockService) {
				svc.GetTotalCostFn = func(ctx context.Context, q model.TotalCostQuery) (int, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/total").
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusBadRequest).
					BodyContains("start_date is required")
			},
		},

		{
			name: "validation error - invalid date",
			setup: func(svc *mockService) {
				svc.GetTotalCostFn = func(ctx context.Context, q model.TotalCostQuery) (int, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/total").
				Query("start_date", "invalid").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Errors)
			},
		},

		{
			name: "validation error - end before start",
			setup: func(svc *mockService) {
				svc.GetTotalCostFn = func(ctx context.Context, q model.TotalCostQuery) (int, error) {
					panic("should not be called")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/total").
				Query("start_date", "02-2026").
				Query("end_date", "01-2026").
				Build(),
			check: func(t *testing.T, resp *Response) {
				var out ErrorResponse

				resp.
					Status(http.StatusBadRequest).
					JSON(&out)

				require.NotEmpty(t, out.Errors)
			},
		},

		{
			name: "service error",
			setup: func(svc *mockService) {
				svc.GetTotalCostFn = func(ctx context.Context, q model.TotalCostQuery) (int, error) {
					return 0, errors.New("db error")
				}
			},
			req: NewRequest().
				Method(http.MethodGet).
				Path("/subscriptions/total").
				Query("start_date", "01-2026").
				Build(),
			check: func(t *testing.T, resp *Response) {
				resp.
					Status(http.StatusInternalServerError).
					BodyContains("failed to calculate total cost")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{}
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := newTestHandler(mockSvc)
			resp := NewResponse(t)

			Call(h, h.GetTotalCost, tt.req, resp)

			tt.check(t, resp)
		})
	}
}
