package datatypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Null[T any] sql.Null[T]

func (n *Null[T]) Scan(value any) error {
	return (*sql.Null[T])(n).Scan(value)
}

func (n Null[T]) Value() (driver.Value, error) {
	return (sql.Null[T])(n).Value()
}

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.V)
}

func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		n.Valid = false
		return nil
	}
	return json.Unmarshal(data, &n.V)
}
