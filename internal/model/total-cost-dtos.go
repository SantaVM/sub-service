package model

import (
	"fmt"
	"strings"
)

type TotalCostQuery struct {
	UserID      *string `validate:"omitempty,uuid"`
	ServiceName *string
	StartDate   string  `validate:"required,monthyear"`
	EndDate     *string `validate:"omitempty,monthyear"`
}

type TotalCostReq struct {
	UserID      *string
	ServiceName *string
	StartDate   MonthYear
	EndDate     MonthYear
}

func (input TotalCostQuery) ToDomain() (*TotalCostReq, error) {
	// --- StartDate ---
	var startDate *MonthYear
	trimmed := strings.TrimSpace(input.StartDate)

	if trimmed != "" {
		var e MonthYear
		if err := e.Parse(trimmed); err != nil {
			return nil, fmt.Errorf("invalid end_date format (expected MM-YYYY): %w", err)
		}
		startDate = &e
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

	// business logic
	if endDate == nil {
		endDate = &MonthYear{Time: startDate.Time.AddDate(0, 1, 0)}
	}

	return &TotalCostReq{
		UserID:      input.UserID,
		ServiceName: input.ServiceName,
		StartDate:   *startDate,
		EndDate:     *endDate,
	}, nil
}

func (t TotalCostQuery) GetStartDate() *string {
	return &t.StartDate
}

func (t TotalCostQuery) GetEndDate() *string {
	return t.EndDate
}
