//go:build go_json

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package json

import json "github.com/goccy/go-json"

var (
	Marshal = json.Marshal

	Unmarshal = json.Unmarshal

	MarshalIndent = json.MarshalIndent

	NewDecoder = json.NewDecoder

	NewEncoder = json.NewEncoder
)

// MarshalToString performs the operation.
func MarshalToString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.ToString(data), nil
}

// UnmarshalFromString performs the operation.
func UnmarshalFromString(str string, v any) error {
	return json.Unmarshal(strings.ToBytes(str), v)
}
