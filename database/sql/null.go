package sql

import (
	"database/sql"
	"database/sql/driver"

	jsonx "github.com/hopeio/gox/encoding/json"
)

type Null[T any] sql.Null[T]

// Scan performs the operation.
func (n *Null[T]) Scan(value any) error {
	return (*sql.Null[T])(n).Scan(value)
}

// Value returns the value.
func (n Null[T]) Value() (driver.Value, error) {
	return (sql.Null[T])(n).Value()
}

// MarshalJSON encodes the value.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return jsonx.Marshal(n.V)
}

// UnmarshalJSON decodes into the value.
func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		n.Valid = false
		return nil
	}
	if err := jsonx.Unmarshal(data, &n.V); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
