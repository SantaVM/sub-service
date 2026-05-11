package handler

import (
	"errors"
	"net/http"
	"sub-service/internal/model"
)

func handleHTTPError(w http.ResponseWriter, err error) {

	var (
		status  int
		message string
		ve      *model.ValidationErrors
	)

	switch {
	case errors.As(err, &ve):
		_ = writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:  ve.Error(),
			Errors: ve.Errors,
		})
		return

	case errors.Is(err, model.ErrInvalidArgument):
		status = http.StatusBadRequest

	case errors.Is(err, model.ErrValidation):
		status = http.StatusBadRequest

	case errors.Is(err, model.ErrInvalidDateRange):
		status = http.StatusBadRequest

	case errors.Is(err, model.ErrSubscriptionOverlap):
		status = http.StatusConflict

	case errors.Is(err, model.ErrConflict):
		status = http.StatusConflict

	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound

	default:
		status = http.StatusInternalServerError
	}

	if status == http.StatusInternalServerError {
		message = "internal server error"
	} else {
		message = err.Error()
	}

	_ = writeJSON(w, status, ErrorResponse{
		Error: message,
	})
}
