package excel

import (
	"fmt"
	"reflect"

	reflectx "github.com/hopeio/gox/reflect"
)

func export[T any](list []T, filename string) error {
	// TODO
	return nil
}

// TODO: support struct map array subField, merge cell
func ToRow(v any) (headers []string, values []any, err error) {
	rv := reflect.ValueOf(v)
	if kind := rv.Kind(); kind == reflect.Ptr || kind == reflect.Interface {
		rv = reflectx.DerefValue(rv)
	}
	if kind := rv.Kind(); !rv.IsValid() || (kind != reflect.Struct && kind != reflect.Map) {
		err = fmt.Errorf("invalid type %T", v)
		return
	}
	switch rv.Kind() {
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			if tag := field.Tag.Get("excel"); tag != "" {
				headers = append(headers, tag)
			} else if tag = field.Tag.Get("json"); tag != "" {
				headers = append(headers, tag)
			} else if tag = field.Tag.Get("comment"); tag != "" {
				headers = append(headers, tag)
			} else {
				headers = append(headers, field.Name)
			}

			values = append(values, rv.Field(i).Interface())
		}
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			headers = append(headers, key.Interface().(string))
			values = append(values, rv.MapIndex(key).Interface())
		}
	}
	return
}
