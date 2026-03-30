/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package structtag

// SubTag 适用于子tag形式等同于struct tag,区别是struct tag使用 " 而SubTag使用 '
// e.g. `key:"example:'1' type:'\\w' test:'\"'"`
type SubTag string

func SubTagLookup(subTag, key string) (value string, ok bool) {
	return SubTag(subTag).Lookup(key)
}

func (tag SubTag) Lookup(key string) (value string, ok bool) {

	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '\'' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '\'' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '\'' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			return qvalue[1 : len(qvalue)-1], true
		}
	}
	return "", false
}
