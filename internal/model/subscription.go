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
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date" swaggertype:"string"`
	EndDate     *MonthYear `json:"end_date,omitempty" swaggertype:"string"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type MonthYear struct {
	time.Time
}

// парсим из Query
func (m *MonthYear) Parse(s string) error {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return err
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
		return err
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
