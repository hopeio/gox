//go:build !(sonic && amd64) && !go_json

package json

import (
	"encoding/json"

	"github.com/hopeio/gox/strings"
)

var (
	Marshal = json.Marshal

	Unmarshal = json.Unmarshal

	MarshalIndent = json.MarshalIndent

	NewDecoder = json.NewDecoder

	NewEncoder = json.NewEncoder
)

func MarshalToString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.FromBytes(data), nil
}

func UnmarshalFromString(str string, v any) error {
	return json.Unmarshal(strings.ToBytes(str), v)
}
