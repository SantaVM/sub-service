package model

import (
	"fmt"
	"strings"
)

// TODO: implement Nullable fields
type UpdateSubscriptionInput struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty,min=2"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty" validate:"omitempty,monthyear"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,monthyear"`
}

type UpdateSubscription struct {
	ServiceName *string
	Price       *int
	StartDate   *MonthYear
	EndDate     *MonthYear
}

func (u UpdateSubscriptionInput) GetStartDate() *string {
	return u.StartDate
}

func (u UpdateSubscriptionInput) GetEndDate() *string {
	return u.EndDate
}

func (input UpdateSubscriptionInput) ToDomain() (*UpdateSubscription, error) {
	// --- ServiceName ---
	var serviceName *string
	if input.ServiceName != nil {
		trimmed := strings.TrimSpace(*input.ServiceName)
		serviceName = &trimmed
	}

	// --- StartDate ---
	// TODO: refactor
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
	// TODO: refactor
	var endDate *MonthYear
	if input.EndDate != nil {
		trimmed := strings.TrimSpace(*input.EndDate)

		if trimmed != "" {
			var e MonthYear
			if err := e.Parse(trimmed); err != nil {
				return nil, fmt.Errorf("invalid end_date format (expected MM-YYYY): %w", err)
			}

			endDate = &e
		}
	}

	return &UpdateSubscription{
		ServiceName: serviceName,
		Price:       input.Price,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}
