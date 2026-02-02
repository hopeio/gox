/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"encoding/base64"
	"io"
	"math"
	"time"
	"unicode/utf8"

	jsonx "github.com/hopeio/gox/encoding/json"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func DefaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := jsonx.NewEncoder(w)
	// For consistency with our custom JSON encoder.
	enc.SetEscapeHTML(false)
	return enc
}

// For JSON-escaping; see CustomEncoder.safeAddString below.
const _hex = "0123456789abcdef"

type CustomEncoder struct {
	*CustomEncoderConfig
	buf            *buffer.Buffer
	openNamespaces int

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

type CustomEncoderConfig struct {
	*zapcore.EncoderConfig
	TransferKey func(key string) string
}

func NewCustomEncoder(cfg *CustomEncoderConfig) *CustomEncoder {
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = zapcore.DefaultLineEnding
	}

	// If no CustomEncoderConfig.NewReflectedEncoder is provided by the user, then use default
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = DefaultReflectedEncoder
	}
	return &CustomEncoder{
		CustomEncoderConfig: cfg,
		buf:                 buffer.NewPool().Get(),
	}
}

func (enc *CustomEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *CustomEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *CustomEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *CustomEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *CustomEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *CustomEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *CustomEncoder) AddComplex64(key string, val complex64) {
	enc.addKey(key)
	enc.AppendComplex64(val)
}

func (enc *CustomEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *CustomEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *CustomEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.AppendFloat32(val)
}

func (enc *CustomEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *CustomEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = buffer.NewPool().Get()
		enc.reflectEnc = enc.NewReflectedEncoder(enc.reflectBuf)
	} else {
		enc.reflectBuf.Reset()
	}
}

var nullLiteralBytes = []byte("null")

// Only invoke the standard JSON encoder if there is actually something to
// encode; otherwise write JSON null literal directly.
func (enc *CustomEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	enc.resetReflectBuf()
	if err := enc.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	enc.reflectBuf.TrimNewline()
	return enc.reflectBuf.Bytes(), nil
}

func (enc *CustomEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	enc.buf.AppendString("\n\n")
	return err
}

func (enc *CustomEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

func (enc *CustomEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
	enc.buf.AppendString("\n\n")
}

func (enc *CustomEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
	enc.buf.AppendString("\n\n")
}

func (enc *CustomEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
	enc.buf.AppendString("\n\n")
}

func (enc *CustomEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	enc.buf.AppendString("\n\n")
	return err
}

func (enc *CustomEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	// Close ONLY new openNamespaces that are created during
	// AppendObject().
	old := enc.openNamespaces
	enc.openNamespaces = 0
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	enc.closeOpenNamespaces()
	enc.openNamespaces = old
	enc.buf.AppendByte('\n')
	return err
}

func (enc *CustomEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *CustomEncoder) AppendByteString(val []byte) {
	enc.safeAddByteString(val)
}

// appendComplex appends the encoded form of the provided complex128 value.
// precision specifies the encoding precision for the real and imaginary
// components of the complex number.
func (enc *CustomEncoder) appendComplex(val complex128, precision int) {
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, precision)
	// If imaginary part is less than 0, minus (-) sign is added by default
	// by AppendFloat.
	if i >= 0 {
		enc.buf.AppendByte('+')
	}
	enc.buf.AppendFloat(i, precision)
	enc.buf.AppendByte('i')
}

func (enc *CustomEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	if e := enc.EncodeDuration; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *CustomEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *CustomEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *CustomEncoder) AppendString(val string) {
	enc.safeAddString(val)
}

func (enc *CustomEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.buf.AppendTime(time, layout)
}

func (enc *CustomEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *CustomEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *CustomEncoder) AddInt(k string, v int)         { enc.AddInt64(k, int64(v)) }
func (enc *CustomEncoder) AddInt32(k string, v int32)     { enc.AddInt64(k, int64(v)) }
func (enc *CustomEncoder) AddInt16(k string, v int16)     { enc.AddInt64(k, int64(v)) }
func (enc *CustomEncoder) AddInt8(k string, v int8)       { enc.AddInt64(k, int64(v)) }
func (enc *CustomEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *CustomEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *CustomEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *CustomEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *CustomEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }
func (enc *CustomEncoder) AppendComplex64(v complex64)    { enc.appendComplex(complex128(v), 32) }
func (enc *CustomEncoder) AppendComplex128(v complex128)  { enc.appendComplex(complex128(v), 64) }
func (enc *CustomEncoder) AppendFloat64(v float64)        { enc.appendFloat(v, 64) }
func (enc *CustomEncoder) AppendFloat32(v float32)        { enc.appendFloat(float64(v), 32) }
func (enc *CustomEncoder) AppendInt(v int)                { enc.AppendInt64(int64(v)) }
func (enc *CustomEncoder) AppendInt32(v int32)            { enc.AppendInt64(int64(v)) }
func (enc *CustomEncoder) AppendInt16(v int16)            { enc.AppendInt64(int64(v)) }
func (enc *CustomEncoder) AppendInt8(v int8)              { enc.AppendInt64(int64(v)) }
func (enc *CustomEncoder) AppendUint(v uint)              { enc.AppendUint64(uint64(v)) }
func (enc *CustomEncoder) AppendUint32(v uint32)          { enc.AppendUint64(uint64(v)) }
func (enc *CustomEncoder) AppendUint16(v uint16)          { enc.AppendUint64(uint64(v)) }
func (enc *CustomEncoder) AppendUint8(v uint8)            { enc.AppendUint64(uint64(v)) }
func (enc *CustomEncoder) AppendUintptr(v uintptr)        { enc.AppendUint64(uint64(v)) }

func (enc *CustomEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *CustomEncoder) clone() *CustomEncoder {
	clone := NewCustomEncoder(enc.CustomEncoderConfig)
	return clone
}

func (enc *CustomEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc

	if final.LevelKey != "" && final.EncodeLevel != nil {
		final.addKey(final.LevelKey)
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeLevel was a no-op. Fall back to strings to keep
			// output JSON valid.
			final.AppendString(ent.Level.String())
		}
		enc.buf.AppendString("\n\n")
	}
	if final.TimeKey != "" {
		final.AddTime(final.TimeKey, ent.Time)
	}
	if ent.LoggerName != "" && final.NameKey != "" {
		final.addKey(final.NameKey)
		cur := final.buf.Len()
		nameEncoder := final.EncodeName

		// if no name encoder provided, fall back to FullNameEncoder for backwards
		// compatibility
		if nameEncoder == nil {
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeName was a no-op. Fall back to strings to
			// keep output JSON valid.
			final.AppendString(ent.LoggerName)
		}
		enc.buf.AppendString("\n\n")
	}
	if ent.Caller.Defined {
		if final.CallerKey != "" {
			final.addKey(final.CallerKey)
			cur := final.buf.Len()
			final.EncodeCaller(ent.Caller, final)
			if cur == final.buf.Len() {
				// User-supplied EncodeCaller was a no-op. Fall back to strings to
				// keep output JSON valid.
				final.AppendString(ent.Caller.String())
			}
			enc.buf.AppendString("\n\n")
		}
		if final.FunctionKey != "" {
			final.addKey(final.FunctionKey)
			final.AppendString(ent.Caller.Function)
			enc.buf.AppendString("\n\n")
		}
	}
	if final.MessageKey != "" {
		final.addKey(enc.MessageKey)
		final.AppendString(ent.Message)
		enc.buf.AppendString("\n\n")
	}

	for i := range fields {
		fields[i].AddTo(enc)
	}

	final.closeOpenNamespaces()
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
		enc.buf.AppendString("\n\n")
	}

	ret := final.buf
	return ret, nil
}

func (enc *CustomEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *CustomEncoder) closeOpenNamespaces() {
	for range enc.openNamespaces {
		enc.buf.AppendByte('}')
	}
	enc.openNamespaces = 0
}

func (enc *CustomEncoder) addKey(key string) {
	if enc.TransferKey != nil {
		enc.buf.AppendString(enc.TransferKey(key))
	} else {
		enc.safeAddString(key)
	}
}

func (enc *CustomEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *CustomEncoder) safeAddString(s string) {
	l := len(s)
	for i := 0; i < l; {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *CustomEncoder) safeAddByteString(s []byte) {
	l := len(s)
	for i := 0; i < l; {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.Write(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *CustomEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *CustomEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}
