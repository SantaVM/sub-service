package model

import "encoding/json"

/*
Nullable needs to specify:
  - field not provided,
  - field provided: value,
  - field provided: null

| Значение                     | JSON                   |
| ---------------------------- | ---------------------- |
| `Set=false`                  | поле отсутствует       |
| `Set=true, Value=nil`        | `"end_date": null`     |
| `Set=true, Value=&"01-2020"` | `"end_date":"01-2020"` |
*/
type Nullable[T any] struct {
	Value *T
	Set   bool
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Set = true

	if string(data) == "null" {
		n.Value = nil

		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	n.Value = &value

	return nil
}

// Эта функция нужна для корректного преобразования в тестах и отображения в Swagger UI.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.Set {
		// Обычно при наличии omitempty поле вообще не попадёт в JSON,
		// но если MarshalJSON вызван, можно вернуть null.
		return []byte("null"), nil
	}

	if n.Value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(*n.Value)
}

func (n Nullable[T]) IsZero() bool {
	return !n.Set
}
