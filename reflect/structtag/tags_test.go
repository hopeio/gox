/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package structtag

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {

	test := []struct {
		name    string
		tag     string
		exp     []*Tag
		invalid bool
	}{
		{
			name: "empty tag",
			tag:  "",
		},
		{
			name:    "tag with one key (invalid)",
			tag:     "json",
			invalid: true,
		},
		{
			name: "tag with one key (valid)",
			tag:  `json:""`,
			exp: []*Tag{
				{
					Key: "json",
				},
			},
		},
		{
			name: "tag with one key and dash name",
			tag:  `json:"-"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "-",
				},
			},
		},
		{
			name: "tag with key and name",
			tag:  `json:"foo"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
			},
		},
		{
			name: "tag with key, name and option",
			tag:  `json:"foo,omitempty"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo,omitempty",
				},
			},
		},
		{
			name: "tag with multiple keys",
			tag:  `json:"" hcl:""`,
			exp: []*Tag{
				{
					Key: "json",
				},
				{
					Key: "hcl",
				},
			},
		},
		{
			name: "tag with multiple keys and names",
			tag:  `json:"foo" hcl:"foo"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
				{
					Key:   "hcl",
					Value: "foo",
				},
			},
		},
		{
			name: "tag with multiple keys and names",
			tag:  `json:"foo" hcl:"foo"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
				{
					Key:   "hcl",
					Value: "foo",
				},
			},
		},
		{
			name: "tag with multiple keys and different names",
			tag:  `json:"foo" hcl:"bar"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
				{
					Key:   "hcl",
					Value: "bar",
				},
			},
		},
		{
			name: "tag with multiple keys, different names and options",
			tag:  `json:"foo,omitempty" structs:"bar,omitnested"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo,omitempty",
				},
				{
					Key:   "structs",
					Value: "bar,omitnested",
				},
			},
		},
		{
			name: "tag with multiple keys, different names and options",
			tag:  `json:"foo" structs:"bar,omitnested" hcl:"-"`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
				{
					Key:   "structs",
					Value: "bar,omitnested",
				},
				{
					Key:   "hcl",
					Value: "-",
				},
			},
		},
		{
			name: "tag with quoted name",
			tag:  `json:"foo,bar:\"baz\""`,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo,bar:\"baz\"",
				},
			},
		},
		{
			name: "tag with trailing space",
			tag:  `json:"foo" `,
			exp: []*Tag{
				{
					Key:   "json",
					Value: "foo",
				},
			},
		},
	}

	for _, ts := range test {
		t.Run(ts.name, func(t *testing.T) {
			tags, err := Parse(ts.tag)
			invalid := err != nil

			if invalid != ts.invalid {
				t.Errorf("invalid case\n\twant: %+v\n\tgot : %+v\n\terr : %s", ts.invalid, invalid, err)
			}

			if invalid {
				return
			}

			got := tags.Tags()
			if !reflect.DeepEqual(ts.exp, got) {
				t.Errorf("parse\n\twant: %#v\n\tgot : %#v", ts.exp, got)
			}

			trimmedInput := strings.TrimSpace(ts.tag)
			if trimmedInput != tags.String() {
				t.Errorf("parse string\n\twant: %#v\n\tgot : %#v", trimmedInput, tags.String())
			}
		})
	}
}

func TestTags_Get(t *testing.T) {
	tag := `json:"foo,omitempty" structs:"bar,omitnested"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	found, ok := tags.Get("json")
	if !ok {
		t.Fatal(ErrTagNotExist)
	}

	t.Run("String", func(t *testing.T) {
		want := `json:"foo,omitempty"`
		if found.String() != want {
			t.Errorf("get\n\twant: %#v\n\tgot : %#v", want, found.String())
		}
	})
	t.Run("Value", func(t *testing.T) {
		want := `foo`
		if found.Value != want {
			t.Errorf("get\n\twant: %#v\n\tgot : %#v", want, found.Value)
		}
	})
	t.Run("GoString", func(t *testing.T) {
		want := `{index: 0,key: 'json',value: 'foo',options: [omitempty]}`
		if found.GoString() != want {
			t.Errorf("get\n\twant: %s\n\tgot : %s", want, found.GoString())
		}
	})
}

func TestTags_Set(t *testing.T) {
	tag := `json:"foo,omitempty" structs:"bar,omitnested"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	tags.Set("json", "bar")
	if err != nil {
		t.Fatal(err)
	}

	found, ok := tags.Get("json")
	if !ok {
		t.Fatal(ErrTagNotExist)
	}

	want := `json:"bar"`
	if found.String() != want {
		t.Errorf("set\n\twant: %#v\n\tgot : %#v", want, found.String())
	}
}

func TestTags_Set_Append(t *testing.T) {
	tag := `json:"foo,omitempty"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	tags.Set("structs", "bar", "omitnested")

	found, ok := tags.Get("structs")
	if !ok {
		t.Fatal(ErrTagNotExist)
	}

	want := `structs:"bar,omitnested"`
	if found.String() != want {
		t.Errorf("set append\n\twant: %#v\n\tgot : %#v", want, found.String())
	}

	wantFull := `json:"foo,omitempty" structs:"bar,omitnested"`
	if tags.String() != wantFull {
		t.Errorf("set append\n\twant: %#v\n\tgot : %#v", wantFull, tags.String())
	}
}

func TestTags_Set_KeyDoesNotExist(t *testing.T) {
	tag := `json:"foo,omitempty" structs:"bar,omitnested"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("setting tag with a nonexisting key should error")
		}
	}()
	tags.Set("", "bar")
}

func TestTags_Delete(t *testing.T) {
	tag := `json:"foo,omitempty" structs:"bar,omitnested" hcl:"-"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	tags.Delete("structs")
	if tags.Len() != 2 {
		t.Fatalf("tag length should be 2, have %d", tags.Len())
	}

	found, ok := tags.Get("json")
	if !ok {
		t.Fatal(ErrTagNotExist)
	}

	want := `json:"foo,omitempty"`
	if found.String() != want {
		t.Errorf("delete\n\twant: %#v\n\tgot : %#v", want, found.String())
	}

	wantFull := `json:"foo,omitempty" hcl:"-"`
	if tags.String() != wantFull {
		t.Errorf("delete\n\twant: %#v\n\tgot : %#v", wantFull, tags.String())
	}
}

func TestTags_DeleteOptions(t *testing.T) {
	tag := `json:"foo,omitempty" structs:"bar,omitnested,omitempty" hcl:"-"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	tags.MustGet("json").DeleteOptions("omitempty")

	want := `json:"foo" structs:"bar,omitnested,omitempty" hcl:"-"`
	if tags.String() != want {
		t.Errorf("delete option\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}

	tags.MustGet("structs").DeleteOptions("omitnested")
	want = `json:"foo" structs:"bar,omitempty" hcl:"-"`
	if tags.String() != want {
		t.Errorf("delete option\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}
}

func TestTags_AddOption(t *testing.T) {
	tag := `json:"foo" structs:"bar,omitempty" hcl:"-"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	tags.MustGet("json").AddOptions("omitempty")

	want := `json:"foo,omitempty" structs:"bar,omitempty" hcl:"-"`
	if tags.String() != want {
		t.Errorf("add options\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}

	// this shouldn't change anything
	tags.MustGet("structs").AddOptions("omitempty")

	want = `json:"foo,omitempty" structs:"bar,omitempty" hcl:"-"`
	if tags.String() != want {
		t.Errorf("add options\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}

	// this should append to the existing
	tags.MustGet("structs").AddOptions("omitnested", "flatten")
	want = `json:"foo,omitempty" structs:"bar,omitempty,omitnested,flatten" hcl:"-"`
	if tags.String() != want {
		t.Errorf("add options\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}
}

func TestTags_String(t *testing.T) {
	tag := `json:"foo" structs:"bar,omitnested" hcl:"-"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	if tags.String() != tag {
		t.Errorf("string\n\twant: %#v\n\tgot : %#v", tag, tags.String())
	}
}

func TestTags_Sort(t *testing.T) {
	tag := `json:"foo" structs:"bar,omitnested" hcl:"-"`

	tags, err := Parse(tag)
	if err != nil {
		t.Fatal(err)
	}

	sort.Sort(tags)

	want := `hcl:"-" json:"foo" structs:"bar,omitnested"`
	if tags.String() != want {
		t.Errorf("string\n\twant: %#v\n\tgot : %#v", want, tags.String())
	}
}
