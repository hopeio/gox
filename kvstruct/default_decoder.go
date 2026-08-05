package kvstruct

var defaultDecoder = NewDecoder("json")

// DefaultDecoder ...
func DefaultDecoder() *Decoder {
	return defaultDecoder
}

// SetAliasTag ...
func SetAliasTag(tag string) {
	defaultDecoder.SetAliasTag(tag)
}

// ZeroEmpty ...
func ZeroEmpty(z bool) {
	defaultDecoder.zeroEmpty = z
}

// IgnoreUnknownKeys ...
func IgnoreUnknownKeys(i bool) {
	defaultDecoder.ignoreUnknownKeys = i
}

// RegisterConverter registers a converter function for a custom type.
func RegisterConverter(value interface{}, converterFunc StringConverter) {
	defaultDecoder.cache.registerConverter(value, converterFunc)
}

// Decode ...
func Decode(dst any, src map[string][]string) error {
	return defaultDecoder.Decode(dst, src)
}
