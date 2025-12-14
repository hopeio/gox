/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package structtag

import (
	"reflect"
	"strings"

	encodingx "github.com/hopeio/gox/encoding/text"
)

/*
SettingTag is a tag for setting

	type example struct {
		db  string `key:"config:db;default:postgres;required"`
	}
*/
type SettingTag string

// ParseSettingTagToMap  parse setting tag, default sep ";" default assignSep ":"
func ParseSettingTagToMap(tag, sep, assignSep string) map[string]string {
	if tag == "" || tag == "-" {
		return nil
	}
	if sep == "" {
		sep = ";"
	}

	settings := map[string]string{}
	names := strings.Split(tag, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}
		if assignSep == "" {
			sep = "="
		}

		values := strings.Split(names[j], assignSep)
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			val := strings.Join(values[1:], assignSep)
			val = strings.ReplaceAll(val, `\"`, `"`)
			settings[k] = val
		} else if k != "" {
			settings[k] = "true"
		}
	}

	return settings
}

func ParseSettingTagToStruct[T any](tag, sep, assignSep string) (*T, error) {
	if tag == "-" {
		return nil, nil
	}
	settings := new(T)
	err := ParseSettingTagIntoStruct(tag, sep, assignSep, settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

const metaTag = "meta"

// ParseSettingTagIntoStruct 解析tag中的子tag, meta标识tag中都有哪些字段
// ParseTagSettingInto default sep ;
/*
type tag struct {
	ConfigName   string `meta:"config"`
	DefaultValue string `meta:"default"`
}
type example struct {
	db  string `specifyTagName:"config:db;default:postgres`
}
var tag tag
ParseSettingTagIntoStruct("tagName",";",":",&tag)
*/

func ParseSettingTagIntoStruct(tag, sep, assignSep string, settings any) error {
	if tag == "-" {
		return ErrTagIgnore
	}
	tagSettings := ParseSettingTagToMap(tag, sep, assignSep)
	if tagSettings == nil {
		return ErrTagNotExist
	}
	settingsValue := reflect.ValueOf(settings).Elem()
	settingsType := reflect.TypeOf(settings).Elem()
	for i := 0; i < settingsValue.NumField(); i++ {
		structField := settingsType.Field(i)
		var name string
		if metatag, ok := structField.Tag.Lookup(metaTag); ok {
			name = metatag
		} else {
			name = structField.Name
		}
		if flagtag, ok := tagSettings[strings.ToUpper(name)]; ok {
			err := encodingx.ParseSetReflectValue(settingsValue.Field(i), flagtag, &structField)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
