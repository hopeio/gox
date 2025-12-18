/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package protobuf

import (
	"fmt"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

var Unmarshal = proto.Unmarshal

type Kind int

const (
	KindVarint Kind = iota
	KindFixed32
	KindFixed64
	KindBytes
	KindString
	KindMessage
	KindPackedVarint
	KindPackedFixed32
	KindPackedFixed64
	KindGroup
)

type Field struct {
	Number protowire.Number
	Wire   protowire.Type
	Kind   Kind
	Value  any
}

func DecodeMessage(b []byte) ([]Field, error) {
	var fields []Field
	for len(b) > 0 {
		num, wt, n := protowire.ConsumeTag(b)
		if n < 0 {
			return nil, fmt.Errorf("consume tag: %d", n)
		}
		b = b[n:]
		var f Field
		f.Number = num
		f.Wire = wt
		switch wt {
		case protowire.VarintType:
			v, m := protowire.ConsumeVarint(b)
			if m < 0 {
				return nil, fmt.Errorf("consume varint: %d", m)
			}
			b = b[m:]
			f.Kind = KindVarint
			f.Value = v
		case protowire.Fixed32Type:
			v, m := protowire.ConsumeFixed32(b)
			if m < 0 {
				return nil, fmt.Errorf("consume fixed32: %d", m)
			}
			b = b[m:]
			f.Kind = KindFixed32
			f.Value = v
		case protowire.Fixed64Type:
			v, m := protowire.ConsumeFixed64(b)
			if m < 0 {
				return nil, fmt.Errorf("consume fixed64: %d", m)
			}
			b = b[m:]
			f.Kind = KindFixed64
			f.Value = v
		case protowire.BytesType:
			v, m := protowire.ConsumeBytes(b)
			if m < 0 {
				return nil, fmt.Errorf("consume bytes: %d", m)
			}
			b = b[m:]
			f.Kind = KindBytes
			f = classifyLengthDelimited(num, v)
		case protowire.StartGroupType:
			v, m := protowire.ConsumeGroup(num, b)
			if m < 0 {
				return nil, fmt.Errorf("consume group: %d", m)
			}
			b = b[m:]
			msg, err := DecodeMessage(v)
			if err != nil {
				return nil, err
			}
			f.Kind = KindGroup
			f.Value = msg
		case protowire.EndGroupType:
			return nil, fmt.Errorf("unexpected end group")
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func classifyLengthDelimited(num protowire.Number, v []byte) Field {

	if msg, err := DecodeMessage(v); err == nil && len(msg) > 0 {
		return Field{Number: num, Wire: protowire.BytesType, Kind: KindMessage, Value: msg}
	}

	if utf8.Valid(v) {
		return Field{Number: num, Wire: protowire.BytesType, Kind: KindString, Value: string(v)}
	}
	if len(v)%4 == 0 {
		pf := consumePackedFixed32(v)
		if pf != nil {
			return Field{Number: num, Wire: protowire.BytesType, Kind: KindPackedFixed32, Value: pf}
		}
	}
	if len(v)%8 == 0 {
		pf := consumePackedFixed64(v)
		if pf != nil {
			return Field{Number: num, Wire: protowire.BytesType, Kind: KindPackedFixed64, Value: pf}
		}
	}

	if pv, ok := consumePackedVarint(v); ok {
		return Field{Number: num, Wire: protowire.BytesType, Kind: KindPackedVarint, Value: pv}
	}
	return Field{Number: num, Wire: protowire.BytesType, Kind: KindBytes, Value: v}
}

func consumePackedVarint(b []byte) ([]uint64, bool) {
	var out []uint64
	for len(b) > 0 {
		v, n := protowire.ConsumeVarint(b)
		if n < 0 {
			return nil, false
		}
		out = append(out, v)
		b = b[n:]
	}
	return out, true
}

func consumePackedFixed32(b []byte) []uint32 {
	var out []uint32
	for len(b) > 0 {
		v, n := protowire.ConsumeFixed32(b)
		if n < 0 {
			return nil
		}
		out = append(out, v)
		b = b[n:]
	}
	return out
}

func consumePackedFixed64(b []byte) []uint64 {
	var out []uint64
	for len(b) > 0 {
		v, n := protowire.ConsumeFixed64(b)
		if n < 0 {
			return nil
		}
		out = append(out, v)
		b = b[n:]
	}
	return out
}

func wireName(t protowire.Type) string {
	switch t {
	case protowire.VarintType:
		return "varint"
	case protowire.Fixed32Type:
		return "fixed32"
	case protowire.Fixed64Type:
		return "fixed64"
	case protowire.BytesType:
		return "bytes"
	case protowire.StartGroupType:
		return "start_group"
	case protowire.EndGroupType:
		return "end_group"
	default:
		return "unknown"
	}
}

func kindName(k Kind) string {
	switch k {
	case KindVarint:
		return "varint"
	case KindFixed32:
		return "fixed32"
	case KindFixed64:
		return "fixed64"
	case KindBytes:
		return "bytes"
	case KindString:
		return "string"
	case KindMessage:
		return "message"
	case KindPackedVarint:
		return "packed_varint"
	case KindPackedFixed32:
		return "packed_fixed32"
	case KindPackedFixed64:
		return "packed_fixed64"
	case KindGroup:
		return "group"
	default:
		return "unknown"
	}
}

func FieldsToMap(fields []Field) map[protowire.Number]any {
	m := make(map[protowire.Number]any)
	for _, f := range fields {

		switch f.Kind {
		case KindVarint, KindFixed32, KindFixed64, KindString, KindPackedVarint, KindPackedFixed32, KindPackedFixed64, KindBytes:
			m[f.Number] = f.Value
		case KindMessage:
			if m[f.Number] != nil {
				if v, ok := m[f.Number].(map[protowire.Number]any); ok {
					m[f.Number] = append([]map[protowire.Number]any(nil), v, FieldsToMap(f.Value.([]Field)))
				} else {
					m[f.Number] = append(m[f.Number].([]map[protowire.Number]any), FieldsToMap(f.Value.([]Field)))
				}

			} else {
				m[f.Number] = FieldsToMap(f.Value.([]Field))
			}
		case KindGroup:
			m[f.Number] = FieldsToMap(f.Value.([]Field))
		}
	}
	return m
}
