package mtos

import (
	"github.com/go-viper/mapstructure/v2"
)

var decodeHook = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.TextUnmarshallerHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
)

type DecoderConfigOption func(*mapstructure.DecoderConfig)

func defaultDecoderConfig(output any, opts ...DecoderConfigOption) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		Squash:           true,
		DecodeHook:       decodeHook,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func Unmarshal(dst any, mapData map[string]any, opts ...DecoderConfigOption) error {
	config := defaultDecoderConfig(dst)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(mapData)
}
