package encoding

import (
	"encoding"
	"encoding/json"
	"errors"
)

var (
	Marshal = json.Marshal

	Unmarshal = json.Unmarshal
)

// UnmarshalTextFor decodes text into a new T and returns it.
// 旧版解码进局部变量后丢弃且无返回值，调用方永远拿不到结果。
func UnmarshalTextFor[T any](text []byte) (T, error) {
	var t T
	itv, ok := any(&t).(encoding.TextUnmarshaler)
	if !ok {
		return t, errors.New("encoding: type does not implement encoding.TextUnmarshaler")
	}
	if err := itv.UnmarshalText(text); err != nil {
		return t, err
	}
	return t, nil
}
