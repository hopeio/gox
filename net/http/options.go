/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"regexp"
	"strings"
)

// Using map for better lookup performance
type ExcludedExtensions map[string]bool

func NewExcludedExtensions(extensions []string) ExcludedExtensions {
	res := make(ExcludedExtensions)
	for _, e := range extensions {
		res[e] = true
	}
	return res
}

func (e ExcludedExtensions) Contains(target string) bool {
	_, ok := e[target]
	return ok
}

type ExcludedPaths []string

func NewExcludedPaths(paths []string) ExcludedPaths {
	return paths
}

func (e ExcludedPaths) Contains(requestURI string) bool {
	for _, path := range e {
		if strings.HasPrefix(requestURI, path) {
			return true
		}
	}
	return false
}

type ExcludedPathsRegex []*regexp.Regexp

func NewExcludedPathsRegex(regexes []string) ExcludedPathsRegex {
	result := make([]*regexp.Regexp, len(regexes), len(regexes))
	for i, reg := range regexes {
		result[i] = regexp.MustCompile(reg)
	}
	return result
}

func (e ExcludedPathsRegex) Contains(requestURI string) bool {
	for _, reg := range e {
		if reg.MatchString(requestURI) {
			return true
		}
	}
	return false
}
