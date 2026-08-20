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

// NewExcludedExtensions creates and returns a new instance.
func NewExcludedExtensions(extensions []string) ExcludedExtensions {
	res := make(ExcludedExtensions)
	for _, e := range extensions {
		res[e] = true
	}
	return res
}

// Contains reports whether the condition holds.
func (e ExcludedExtensions) Contains(target string) bool {
	_, ok := e[target]
	return ok
}

type ExcludedPaths []string

// NewExcludedPaths creates and returns a new instance.
func NewExcludedPaths(paths []string) ExcludedPaths {
	return paths
}

// Contains reports whether the condition holds.
func (e ExcludedPaths) Contains(requestURI string) bool {
	for _, path := range e {
		if strings.HasPrefix(requestURI, path) {
			return true
		}
	}
	return false
}

type ExcludedPathsRegex []*regexp.Regexp

// NewExcludedPathsRegex creates and returns a new instance.
func NewExcludedPathsRegex(regexes []string) ExcludedPathsRegex {
	result := make([]*regexp.Regexp, len(regexes))
	for i, reg := range regexes {
		result[i] = regexp.MustCompile(reg)
	}
	return result
}

// Contains reports whether the condition holds.
func (e ExcludedPathsRegex) Contains(requestURI string) bool {
	for _, reg := range e {
		if reg.MatchString(requestURI) {
			return true
		}
	}
	return false
}
