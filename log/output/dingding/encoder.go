/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingding

import (
	"encoding/base64"
	"math"
	"time"
	"unicode/utf8"

	"github.com/hopeio/gox/log/output"
	"go.uber.org/zap/zapcore"

	"go.uber.org/zap/buffer"
)

// For JSON-escaping; see dingEncoder.safeAddString below.
const _hex = "0123456789abcdef"

type dingEncoder struct {
	*zapcore.EncoderConfig
	buf            *buffer.Buffer
	openNamespaces int

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc zapcore.ReflectedEncoder
}

func NewDingEncoder(cfg *zapcore.EncoderConfig) zapcore.Encoder {
	return newDingEncoder(cfg)
}

func newDingEncoder(cfg *zapcore.EncoderConfig) *dingEncoder {
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = zapcore.DefaultLineEnding
	}

	// If no EncoderConfig.NewReflectedEncoder is provided by the user, then use default
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = output.DefaultReflectedEncoder
	}
	return &dingEncoder{
		EncoderConfig: cfg,
		buf:           buffer.NewPool().Get(),
	}
}

func (enc *dingEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *dingEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *dingEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *dingEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *dingEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *dingEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *dingEncoder) AddComplex64(key string, val complex64) {
	enc.addKey(key)
	enc.AppendComplex64(val)
}

func (enc *dingEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *dingEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *dingEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.AppendFloat32(val)
}

func (enc *dingEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *dingEncoder) resetReflectBuf() {
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
func (enc *dingEncoder) encodeReflected(obj interface{}) ([]byte, error) {
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

func (enc *dingEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	enc.buf.AppendString("\n\n")
	return err
}

func (enc *dingEncoder) OpenNamespace(key string) {
	enc.addKey(key)
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

func (enc *dingEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
	enc.buf.AppendString("\n\n")
}

func (enc *dingEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
	enc.buf.AppendString("\n\n")
}

func (enc *dingEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
	enc.buf.AppendString("\n\n")
}

func (enc *dingEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	enc.buf.AppendString("\n\n")
	return err
}

func (enc *dingEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
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

func (enc *dingEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *dingEncoder) AppendByteString(val []byte) {
	enc.safeAddByteString(val)
}

// appendComplex appends the encoded form of the provided complex128 value.
// precision specifies the encoding precision for the real and imaginary
// components of the complex number.
func (enc *dingEncoder) appendComplex(val complex128, precision int) {
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

func (enc *dingEncoder) AppendDuration(val time.Duration) {
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

func (enc *dingEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *dingEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *dingEncoder) AppendString(val string) {
	enc.safeAddString(val)
}

func (enc *dingEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.buf.AppendTime(time, layout)
}

func (enc *dingEncoder) AppendTime(val time.Time) {
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

func (enc *dingEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *dingEncoder) AddInt(k string, v int)         { enc.AddInt64(k, int64(v)) }
func (enc *dingEncoder) AddInt32(k string, v int32)     { enc.AddInt64(k, int64(v)) }
func (enc *dingEncoder) AddInt16(k string, v int16)     { enc.AddInt64(k, int64(v)) }
func (enc *dingEncoder) AddInt8(k string, v int8)       { enc.AddInt64(k, int64(v)) }
func (enc *dingEncoder) AddUint(k string, v uint)       { enc.AddUint64(k, uint64(v)) }
func (enc *dingEncoder) AddUint32(k string, v uint32)   { enc.AddUint64(k, uint64(v)) }
func (enc *dingEncoder) AddUint16(k string, v uint16)   { enc.AddUint64(k, uint64(v)) }
func (enc *dingEncoder) AddUint8(k string, v uint8)     { enc.AddUint64(k, uint64(v)) }
func (enc *dingEncoder) AddUintptr(k string, v uintptr) { enc.AddUint64(k, uint64(v)) }
func (enc *dingEncoder) AppendComplex64(v complex64)    { enc.appendComplex(complex128(v), 32) }
func (enc *dingEncoder) AppendComplex128(v complex128)  { enc.appendComplex(complex128(v), 64) }
func (enc *dingEncoder) AppendFloat64(v float64)        { enc.appendFloat(v, 64) }
func (enc *dingEncoder) AppendFloat32(v float32)        { enc.appendFloat(float64(v), 32) }
func (enc *dingEncoder) AppendInt(v int)                { enc.AppendInt64(int64(v)) }
func (enc *dingEncoder) AppendInt32(v int32)            { enc.AppendInt64(int64(v)) }
func (enc *dingEncoder) AppendInt16(v int16)            { enc.AppendInt64(int64(v)) }
func (enc *dingEncoder) AppendInt8(v int8)              { enc.AppendInt64(int64(v)) }
func (enc *dingEncoder) AppendUint(v uint)              { enc.AppendUint64(uint64(v)) }
func (enc *dingEncoder) AppendUint32(v uint32)          { enc.AppendUint64(uint64(v)) }
func (enc *dingEncoder) AppendUint16(v uint16)          { enc.AppendUint64(uint64(v)) }
func (enc *dingEncoder) AppendUint8(v uint8)            { enc.AppendUint64(uint64(v)) }
func (enc *dingEncoder) AppendUintptr(v uintptr)        { enc.AppendUint64(uint64(v)) }

func (enc *dingEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *dingEncoder) clone() *dingEncoder {
	clone := newDingEncoder(enc.EncoderConfig)
	return clone
}

func (enc *dingEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
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

func (enc *dingEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *dingEncoder) closeOpenNamespaces() {
	for range enc.openNamespaces {
		enc.buf.AppendByte('}')
	}
	enc.openNamespaces = 0
}

func (enc *dingEncoder) addKey(key string) {
	enc.buf.AppendString("**")
	enc.safeAddString(key)
	enc.buf.AppendString("**: ")
	enc.buf.AppendByte(' ')
}

func (enc *dingEncoder) appendFloat(val float64, bitSize int) {
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
func (enc *dingEncoder) safeAddString(s string) {
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
func (enc *dingEncoder) safeAddByteString(s []byte) {
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
func (enc *dingEncoder) tryAddRuneSelf(b byte) bool {
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

func (enc *dingEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}
