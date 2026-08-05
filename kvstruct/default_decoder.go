package kvstruct

var defaultDecoder = NewDecoder("json")

// DefaultDecoder returns the result.
func DefaultDecoder() *Decoder {
	return defaultDecoder
}

// SetAliasTag updates or inserts a value.
func SetAliasTag(tag string) {
	defaultDecoder.SetAliasTag(tag)
}

// ZeroEmpty performs the operation.
func ZeroEmpty(z bool) {
	defaultDecoder.zeroEmpty = z
}

// IgnoreUnknownKeys performs the operation.
func IgnoreUnknownKeys(i bool) {
	defaultDecoder.ignoreUnknownKeys = i
}

// RegisterConverter registers a converter function for a custom type.
func RegisterConverter(value interface{}, converterFunc StringConverter) {
	defaultDecoder.cache.registerConverter(value, converterFunc)
}

// Decode formats or converts the value.
func Decode(dst any, src map[string][]string) error {
	return defaultDecoder.Decode(dst, src)
}
