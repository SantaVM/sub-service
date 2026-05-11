package model

import "errors"

var (
	ErrNotFound        = errors.New("entity not found")
	ErrConflict        = errors.New("conflict")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrValidation      = errors.New("validation error")

	ErrInvalidDateRange    = errors.New("invalid date range")
	ErrSubscriptionOverlap = errors.New("subscription overlap")
)

type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v *ValidationErrors) Error() string {
	return "validation failed"
}
