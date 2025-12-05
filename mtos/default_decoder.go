package mtos

import (
	reflectx "github.com/hopeio/gox/encoding"
)

var defaultDecoder = NewDecoder("json")

func DefaultDecoder() *Decoder {
	return defaultDecoder
}

func SetAliasTag(tag string) {
	defaultDecoder.SetAliasTag(tag)
}

func ZeroEmpty(z bool) {
	defaultDecoder.zeroEmpty = z
}

func IgnoreUnknownKeys(i bool) {
	defaultDecoder.ignoreUnknownKeys = i
}

// RegisterConverter registers a converter function for a custom type.
func RegisterConverter(value interface{}, converterFunc reflectx.StringConverter) {
	defaultDecoder.cache.registerConverter(value, converterFunc)
}

func Decode(dst any, src map[string][]string) error {
	return defaultDecoder.Decode(dst, src)
}
