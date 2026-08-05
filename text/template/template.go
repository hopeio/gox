/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package template

import (
	"io"
	"strings"
	"text/template"

	"github.com/hopeio/gox/log"
)

var CommonTemp = template.New("all")

// init initializes package state.
func init() {
	CommonTemp.Funcs(template.FuncMap{"join": strings.Join})
}

// Parse parses the input.
func Parse(tpl string) *template.Template {
	t, err := CommonTemp.Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// Execute executes the operation.
func Execute(wr io.Writer, name string, data interface{}) error {
	return CommonTemp.ExecuteTemplate(wr, name, data)
}
