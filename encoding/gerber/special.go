/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gerber

import (
	"errors"
	"strings"
)

// Oval ...
func (p *commandProcessor) Oval(lineIdx int, word string) error {
	var pe primitive
	tmpl := template{Line: lineIdx, Name: "Oval"}
	splitted := strings.Split(word, primitiveDelimiter)
	if len(splitted) == 0 {
		return errors.New("no splitted")
	}
	curLine := lineIdx
	if strings.Contains(splitted[0], "\n") {
		curLine++
	}
	pe.code = primitiveCodeOval
	pe.value = make([]primitiveValue, 6)
	pe.value[0] = primitiveValue{value: 1, varIndex: -1}
	for i := 0; i < 5; i++ {
		pe.value[i+1] = primitiveValue{varIndex: i}
	}
	tmpl.Primitives = append(tmpl.Primitives, pe)
	p.templates["Oval"] = tmpl
	return nil
}
