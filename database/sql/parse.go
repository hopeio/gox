package sql

import (
	"database/sql"

	stringsx "github.com/hopeio/gox/strings"
)

func StringConvertFor[T any](str string) (T, error) {
	var t T
	a, ap := any(t), any(&t)
	isv, ok := a.(sql.Scanner)
	if !ok {
		isv, ok = ap.(sql.Scanner)
	}
	if ok {
		err := isv.Scan(str)
		if err != nil {
			return t, err
		}
		return t, nil
	}
	return stringsx.ParseFor[T](str)
}
