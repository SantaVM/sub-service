package model

import (
	"strings"
)

// TODO: implement Nullable fields
type UpdateSubscriptionInput struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty,min=2"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty" validate:"omitempty,datetime=01-2006"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006"`
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
				return nil, err
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
				return nil, err
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
