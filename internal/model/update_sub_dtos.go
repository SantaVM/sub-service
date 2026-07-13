package model

import (
	"strings"
)

type UpdateSubscriptionInput struct {
	ServiceName *string          `json:"service_name,omitempty" validate:"omitempty,min=2"`
	Price       *int             `json:"price,omitempty" validate:"omitempty,min=0"`
	StartDate   *string          `json:"start_date,omitempty" validate:"omitempty,datetime=01-2006" example:"06-2026"`
	EndDate     Nullable[string] `json:"end_date" swaggertype:"string" example:"06-2027"`
}

type UpdateSubscription struct {
	ServiceName *string
	Price       *int
	StartDate   *MonthYear
	EndDate     Nullable[MonthYear]
}

func (u UpdateSubscriptionInput) GetStartDate() *string {
	return u.StartDate
}

func (u UpdateSubscriptionInput) GetEndDate() *string {
	return u.EndDate.Value
}

func (input UpdateSubscriptionInput) ToDomain() (*UpdateSubscription, error) {
	// --- ServiceName ---
	var serviceName *string
	if input.ServiceName != nil {
		trimmed := strings.TrimSpace(*input.ServiceName)
		serviceName = &trimmed
	}

	// --- StartDate ---
	var startDate *MonthYear
	if input.StartDate != nil {
		var e MonthYear

		if err := e.Parse(*input.StartDate); err != nil {
			return nil, err
		}

		startDate = &e
	}

	// --- EndDate ---
	var nullable Nullable[MonthYear]
	if input.EndDate.Set {
		nullable.Set = true

		if input.EndDate.Value != nil {
			var end MonthYear

			if err := end.Parse(*input.EndDate.Value); err != nil {
				return nil, err
			}

			nullable.Value = &end
		}
	}

	return &UpdateSubscription{
		ServiceName: serviceName,
		Price:       input.Price,
		StartDate:   startDate,
		EndDate:     nullable,
	}, nil
}

/*
Валидируем Nullable поля только
*/
func (u UpdateSubscriptionInput) Validate() *ValidationErrors {
	var validationErrors ValidationErrors

	if u.EndDate.Value != nil {
		var endDate MonthYear
		if err := endDate.Parse(*u.EndDate.Value); err != nil {
			validationErrors.Errors = append(validationErrors.Errors, ValidationError{
				Field:   "end_date",
				Message: "end date must be in MM-YYYY format",
			})
		}
	}

	if len(validationErrors.Errors) > 0 {
		return &validationErrors
	}

	return nil
}
