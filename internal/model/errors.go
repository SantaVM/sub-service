package model

import "errors"

var (
	ErrInvalidDateRange    = errors.New("invalid date range")
	ErrSubscriptionOverlap = errors.New("subscription overlap")
)
