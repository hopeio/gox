package sql

import (
	"database/sql"

	strconvx "github.com/hopeio/gox/strconv"
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
	return strconvx.ParseFor[T](str)
}
