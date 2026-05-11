package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sub-service/internal/model"
	"time"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	v.RegisterValidation("monthyear", validateMonthYear)

	v.RegisterStructValidation(validateStartEndDate, model.TotalCostQuery{})
	v.RegisterStructValidation(validateStartEndDate, model.UpdateSubscriptionInput{})
	v.RegisterStructValidation(validateStartEndDate, model.CreateSubscriptionInput{})

	return &Validator{v: v}
}

func (v *Validator) convertValidationErrors(err error) error {
	var ve validator.ValidationErrors

	if !errors.As(err, &ve) {
		return err
	}

	result := make([]model.ValidationError, 0, len(ve))

	for _, e := range ve {
		result = append(result, model.ValidationError{
			Field:   e.Field(),
			Message: v.mapMessage(e),
		})
	}

	return &model.ValidationErrors{
		Errors: result,
	}
}

func validateMonthYear(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if value == "" {
		return true // пусть "required" решает
	}

	_, err := time.Parse("01-2006", value)
	return err == nil
}

// Параметр dst принимает УКАЗАТЕЛЬ на структуру
func (v *Validator) BindAndValidate(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body %w", errors.Join(model.ErrInvalidArgument, err))
	}

	if err := v.v.Struct(dst); err != nil {
		return v.convertValidationErrors(err)
	}

	return nil
}

func (v *Validator) ValidateQuery(dst any) error {
	if err := v.v.Struct(dst); err != nil {
		return v.convertValidationErrors(err)
	}

	return nil
}

func (v *Validator) mapMessage(e validator.FieldError) string {
	switch e.Tag() {

	case "required":
		return e.Field() + " is required"

	case "uuid":
		return e.Field() + " must be a valid UUID"

	case "max":
		return e.Field() + " must be <= " + e.Param()

	case "min":
		return e.Field() + " must be >= " + e.Param()

	case "monthyear":
		return e.Field() + " must be in MM-YYYY format"

	case "after_start":
		return "end_date must be after start_date"

	default:
		return e.Field() + " is invalid"
	}
}

type HasStartEndDate interface {
	GetStartDate() *string
	GetEndDate() *string
}

var _ HasStartEndDate = (*model.UpdateSubscriptionInput)(nil)
var _ HasStartEndDate = (*model.CreateSubscriptionInput)(nil)
var _ HasStartEndDate = (*model.TotalCostQuery)(nil)

func validateStartEndDate(sl validator.StructLevel) {
	query, ok := sl.Current().Interface().(HasStartEndDate)
	if !ok {
		return
	}

	if query.GetStartDate() == nil {
		return
	}

	start, err := time.Parse("01-2006", *query.GetStartDate())
	if err != nil {
		return
	}

	if query.GetEndDate() == nil {
		return
	}

	end, err := time.Parse("01-2006", *query.GetEndDate())
	if err != nil {
		return
	}

	// сама проверка
	if !start.Before(end) {
		sl.ReportError(
			query.GetEndDate(),
			"end_date",
			"EndDate",
			"after_start",
			"",
		)
	}
}
