package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CreateSubscriptionInput struct {
	ServiceName string  `json:"service_name" validate:"required,min=2"`
	Price       int     `json:"price" validate:"required,min=0"`
	UserID      string  `json:"user_id" validate:"required,uuid"`
	StartDate   string  `json:"start_date" validate:"required,monthyear"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,monthyear"`
}

type CreateSubscription struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   MonthYear
	EndDate     *MonthYear
}

func (c CreateSubscriptionInput) GetStartDate() *string {
	return &c.StartDate
}

func (c CreateSubscriptionInput) GetEndDate() *string {
	return c.EndDate
}

func (input CreateSubscriptionInput) ToDomain() (*CreateSubscription, error) {
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
