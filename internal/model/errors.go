package model

import "errors"

var (
	ErrInvalidDateRange    = errors.New("invalid date range")
	ErrSubscriptionOverlap = errors.New("subscription overlap")
)

type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
