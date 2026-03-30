package structtag

import (
	"reflect"
	"testing"
)

type Bar2 struct {
	Field1 int    `mock:"example=1;type=\\w;test=\""`
	Field2 string `mock:"example:1;type:\\w;test:\""`
}

func TestSettingTag(t *testing.T) {
	var bar Bar2
	typ := reflect.TypeOf(bar)
	t.Log(ParseSettingTagToMap(typ.Field(0).Tag.Get("mock"), ";", "="))
	t.Log(ParseSettingTagToMap(typ.Field(1).Tag.Get("mock"), ";", ":"))
}
