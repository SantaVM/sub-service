package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uint       `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id" example:"b695926c-7d36-477b-b945-f14d05793c14"`
	StartDate   MonthYear  `json:"start_date" swaggertype:"string" example:"06-2026"`
	EndDate     *MonthYear `json:"end_date,omitempty" swaggertype:"string" example:"06-2027"`
	CreatedAt   time.Time  `json:"created_at" format:"date-time" example:"2026-06-19T00:26:11Z"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" format:"date-time" example:"2026-07-19T00:26:11Z"`
}

type MonthYear struct {
	time.Time
}

// парсим из Query
func (m *MonthYear) Parse(s string) error {
	trimmed := strings.TrimSpace(s)

	t, err := time.Parse("01-2006", trimmed)
	if err != nil {
		return fmt.Errorf("invalid date format (expected MM-YYYY): %w", err)
	}

	m.Time = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return nil
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	s := strings.Trim(string(data), `"`)

	t, err := time.Parse("01-2006", s)
	if err != nil {
		return fmt.Errorf("invalid date format (expected MM-YYYY): %w", err)
	}

	m.Time = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return nil
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Format("01-2006") + `"`), nil
}

// чтение из БД - Scanner
func (m *MonthYear) Scan(value any) error {
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan %T into MonthYear", value)
	}

	m.Time = t.UTC()
	return nil
}

// запись в БД - Valuer
func (m MonthYear) Value() (driver.Value, error) {
	return m.Time, nil
}
