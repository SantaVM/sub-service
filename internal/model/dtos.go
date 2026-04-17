package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CreateSubscriptionInput struct {
	ServiceName string  `json:"service_name" validate:"required"`
	Price       int     `json:"price" validate:"required,min=0"`
	UserID      string  `json:"user_id" validate:"required,uuid"`
	StartDate   string  `json:"start_date" validate:"required"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty"`
}

type CreateSubscription struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   MonthYear
	EndDate     *MonthYear
}

func (input CreateSubscriptionInput) ToDomain() (*CreateSubscription, error) {
	// --- ServiceName ---
	if strings.TrimSpace(input.ServiceName) == "" {
		return nil, fmt.Errorf("service_name is required")
	}

	// --- Price ---
	if input.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	// --- UserID ---
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// --- StartDate ---
	var startDate MonthYear
	if err := startDate.Parse(input.StartDate); err != nil {
		return nil, fmt.Errorf("invalid start_date format (expected MM-YYYY): %w", err)
	}

	// --- EndDate ---
	var endDate *MonthYear
	if input.EndDate != nil {
		trimmed := strings.TrimSpace(*input.EndDate)

		if trimmed != "" {
			var e MonthYear
			if err := e.Parse(trimmed); err != nil {
				return nil, fmt.Errorf("invalid end_date format (expected MM-YYYY): %w", err)
			}

			// --- Business rule validation ---
			if !startDate.Time.Before(e.Time) {
				return nil, ErrInvalidDateRange
			}

			endDate = &e
		}
	}

	return &CreateSubscription{
		ServiceName: strings.TrimSpace(input.ServiceName),
		Price:       input.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type UpdateSubscriptionInput struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty" validate:"omitempty"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty"`
}

type UpdateSubscription struct {
	ServiceName *string
	Price       *int
	StartDate   *MonthYear
	EndDate     *MonthYear
}

func (input UpdateSubscriptionInput) ToDomain() (*UpdateSubscription, error) {
	// --- ServiceName ---
	var serviceName *string
	if input.ServiceName != nil {
		trimmed := strings.TrimSpace(*input.ServiceName)
		if trimmed == "" {
			return nil, fmt.Errorf("service_name can not be emty")
		}
		serviceName = &trimmed
	}

	// --- Price ---
	var price *int
	if input.Price != nil {
		if *input.Price < 0 {
			return nil, fmt.Errorf("price must be non-negative")
		}
		price = input.Price
	}

	// --- StartDate ---
	var startDate *MonthYear
	if input.StartDate != nil {
		trimmed := strings.TrimSpace(*input.StartDate)

		if trimmed != "" {
			var e MonthYear
			if err := e.Parse(trimmed); err != nil {
				return nil, fmt.Errorf("invalid end_date format (expected MM-YYYY): %w", err)
			}
			startDate = &e
		}
	}

	// --- EndDate ---
	var endDate *MonthYear
	if input.EndDate != nil {
		trimmed := strings.TrimSpace(*input.EndDate)

		if trimmed != "" {
			var e MonthYear
			if err := e.Parse(trimmed); err != nil {
				return nil, fmt.Errorf("invalid end_date format (expected MM-YYYY): %w", err)
			}

			// --- Business rule validation ---
			if startDate != nil {
				if !startDate.Time.Before(e.Time) {
					return nil, ErrInvalidDateRange
				}
			}

			endDate = &e
		}
	}

	return &UpdateSubscription{
		ServiceName: serviceName,
		Price:       price,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type ListSubscriptionsQuery struct {
	UserID      *string
	ServiceName *string
	Size        int
	Page        int
}

type TotalCostQuery struct {
	UserID      *string
	ServiceName *string
	StartDate   MonthYear
	EndDate     MonthYear
}
