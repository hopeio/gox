/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package kvstruct

import (
	"reflect"
	"testing"
)

func reflectTypeField(t *testing.T, sample any, name string) reflect.StructField {
	t.Helper()
	f, ok := reflect.TypeOf(sample).FieldByName(name)
	if !ok {
		t.Fatalf("field %q not found", name)
	}
	return f
}

func reflectValueField(t *testing.T, ptr any, name string) reflect.Value {
	t.Helper()
	v := reflect.ValueOf(ptr).Elem().FieldByName(name)
	if !v.IsValid() {
		t.Fatalf("field %q not found", name)
	}
	return v
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag       string
		wantAlias string
		wantOpt   *Options
	}{
		{"", "", nil},
		{"name", "name", &Options{}},
		{"alias,required", "alias", &Options{Required: true}},
		{"alias,omitempty", "alias", &Options{Omitempty: true}},
		{"alias,default=foo", "alias", &Options{Default: "foo"}},
		{"-", "-", &Options{}},
	}
	for _, tt := range tests {
		alias, opt := ParseTag(tt.tag)
		if alias != tt.wantAlias {
			t.Errorf("ParseTag(%q) alias = %q, want %q", tt.tag, alias, tt.wantAlias)
		}
		if tt.wantOpt == nil {
			if opt != nil {
				t.Errorf("ParseTag(%q) opt = %#v, want nil", tt.tag, opt)
			}
			continue
		}
		if opt == nil {
			t.Fatalf("ParseTag(%q) opt is nil", tt.tag)
		}
		if opt.Required != tt.wantOpt.Required || opt.Omitempty != tt.wantOpt.Omitempty || opt.Default != tt.wantOpt.Default {
			t.Errorf("ParseTag(%q) opt = %#v, want %#v", tt.tag, *opt, *tt.wantOpt)
		}
	}
}

func TestKVSourceTrySet(t *testing.T) {
	type cfg struct {
		Port int    `form:"port"`
		Name string `form:"name"`
	}
	var c cfg
	form := KVSource{"port": "8080", "name": "svc"}
	fieldPort := reflectTypeField(t, cfg{}, "Port")
	fieldName := reflectTypeField(t, cfg{}, "Name")

	ok, err := form.TrySet(reflectValueField(t, &c, "Port"), &fieldPort, "port", nil)
	if err != nil || !ok || c.Port != 8080 {
		t.Fatalf("port: ok=%v err=%v port=%d", ok, err, c.Port)
	}
	ok, err = form.TrySet(reflectValueField(t, &c, "Name"), &fieldName, "name", nil)
	if err != nil || !ok || c.Name != "svc" {
		t.Fatalf("name: ok=%v err=%v name=%q", ok, err, c.Name)
	}
}

func TestKVSourceTrySetDefault(t *testing.T) {
	type cfg struct {
		Level int `form:"level"`
	}
	var c cfg
	form := KVSource{}
	field := reflectTypeField(t, cfg{}, "Level")
	opt := &Options{Default: "3"}
	ok, err := form.TrySet(reflectValueField(t, &c, "Level"), &field, "level", opt)
	if err != nil || !ok || c.Level != 3 {
		t.Fatalf("ok=%v err=%v level=%d", ok, err, c.Level)
	}
}

func TestKVSourceTrySetMissing(t *testing.T) {
	type cfg struct {
		Level int `form:"level"`
	}
	var c cfg
	form := KVSource{}
	field := reflectTypeField(t, cfg{}, "Level")
	ok, err := form.TrySet(reflectValueField(t, &c, "Level"), &field, "level", nil)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}

func TestMapping(t *testing.T) {
	type inner struct {
		X int `form:"x"`
	}
	type cfg struct {
		inner
		Y string `form:"y"`
	}
	var c cfg
	form := KVSource{"x": "1", "y": "z"}
	if err := Mapping(&c, form, "form"); err != nil {
		t.Fatal(err)
	}
	if c.X != 1 || c.Y != "z" {
		t.Fatalf("got %#v", c)
	}
}
