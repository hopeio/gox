package kvstruct

import (
	"reflect"
	"strings"

	"github.com/hopeio/gox/strconv"
)

// Setter tries to set value on a walking by fields of a struct
type Setter interface {
	TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error)
}

type Setters []Setter

func (receiver Setters) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	for _, arg := range receiver {
		if arg != nil {
			isSet, err = arg.TrySet(value, field, key, opt)
			if isSet {
				return
			}
		}
	}
	return
}

func MappingByTag(ptr any, setter Setter, tag string) error {
	_, err := mapping(reflect.ValueOf(ptr), nil, setter, tag)
	return err
}

func mapping(value reflect.Value, field *reflect.StructField, setter Setter, tag string) (bool, error) {
	var tagValue string
	if field != nil {
		tagValue = field.Tag.Get(tag)
	}
	if tagValue == "-" { // just ignoring this field
		return false, nil
	}

	var vKind = value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), &sf, setter, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}

	if field != nil && !field.Anonymous {
		ok, err := tryToSetValue(value, field, setter, tagValue)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

type Options struct {
	Default   string
	Required  bool
	Omitempty bool
}

func ParseTag(tagValue string) (string, *Options) {
	if tagValue == "" { // when field is "emptyField" variable
		return "", nil
	}
	alias, opts, _ := strings.Cut(tagValue, ",")
	var opt string
	var setOpt Options
	for len(opts) > 0 {
		opt, opts, _ = strings.Cut(opts, ",")

		if k, v, _ := strings.Cut(opt, "="); k == "default" {
			setOpt.Default = v
			break
		}
		if opt == "required" {
			setOpt.Required = true
			break
		}
		if opt == "omitempty" {
			setOpt.Omitempty = true
			break
		}
	}
	return alias, &setOpt
}

func tryToSetValue(value reflect.Value, field *reflect.StructField, setter Setter, tagValue string) (bool, error) {

	alias, opts := ParseTag(tagValue)
	if alias == "" { // default value is FieldName
		alias = field.Name
	}
	return setter.TrySet(value, field, alias, opts)
}

type Getter interface {
	Get(key string) (string, bool)
}

type GetFunc func(key string) (string, bool)

func (f GetFunc) Get(key string) (string, bool) {
	return f(key)
}

func (f GetFunc) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByGetter(value, field, f, key, opt)
}

type Getters []Getter

func (args Getters) Get(key string) (v string, ok bool) {
	for i := range args {
		if v, ok = args[i].Get(key); ok {
			return
		}
	}
	return
}
func (args Getters) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByGetter(value, field, args, key, opt)
}

type KVSource map[string]string

func (form KVSource) Get(key string) (string, bool) {
	v, ok := form[key]
	return v, ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByGetter(value, field, form, key, opt)
}

type KVsSource map[string][]string

var _ Setter = KVsSource(nil)

func (form KVsSource) Get(key string) ([]string, bool) {
	v, ok := form[key]
	return v, ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByValuesGetter(value, field, form, key, opt)
}

type ValuesGetter interface {
	Get(key string) ([]string, bool)
}

type ValuesGetters []ValuesGetter

func (args ValuesGetters) Get(key string) (v []string, ok bool) {
	for i := range args {
		if v, ok = args[i].Get(key); ok {
			return
		}
	}
	return
}

func (args ValuesGetters) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByValuesGetter(value, field, args, key, opt)
}

func SetValueByGetter(value reflect.Value, field *reflect.StructField, getter Getter, key string, opt *Options) (isSet bool, err error) {
	vs, _ := getter.Get(key)
	if vs == "" {
		if opt == nil || opt.Default == "" {
			return false, nil
		}
		vs = opt.Default
	}
	err = strconv.ParseStringSetReflectValue(value, vs, field)
	if err != nil {
		return false, err
	}
	return true, nil
}

func SetValueByValuesGetter(value reflect.Value, field *reflect.StructField, getter ValuesGetter, key string, opt *Options) (isSet bool, err error) {
	vals, _ := getter.Get(key)
	if len(vals) == 0 {
		if opt == nil || opt.Default == "" {
			return false, nil
		}
		vals = strings.Split(opt.Default, ",")
	}

	err = strconv.ParseStringsSetReflectValue(value, vals, field)
	if err != nil {
		return false, err
	}
	return true, nil
}
