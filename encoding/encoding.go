package encoding

import (
	"encoding"
	"encoding/json"
)

var (
	Marshal = json.Marshal

	Unmarshal = json.Unmarshal
)

func UnmarshalTextFor[T any](text []byte) error {
	var t T
	v, vp := any(t), any(&t)
	itv, ok := v.(encoding.TextUnmarshaler)
	if !ok {
		itv, ok = vp.(encoding.TextUnmarshaler)
	}
	if ok {
		err := itv.UnmarshalText(text)
		if err != nil {
			return err
		}
	}
	return nil
}
