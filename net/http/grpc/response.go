package grpc

import (
	"errors"
	"reflect"

	errorsx "github.com/hopeio/gox/errors"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type CommonResp[T proto.Message] httpx.CommonResp[T]

func (r *CommonResp[T]) MarshalProto() ([]byte, error) {
	var buf []byte

	if r.Code != 0 {
		buf = protowire.AppendVarint(buf, 0x08)
		buf = protowire.AppendVarint(buf, uint64(r.Code))
	}

	// 编码 Msg 字段 (field number 2, string)
	if r.Msg != "" {
		buf = protowire.AppendVarint(buf, 0x12)
		buf = protowire.AppendString(buf, r.Msg)
	}

	// 编码 Data 字段 (field number 3, bytes)
	if !reflect.ValueOf(r.Data).IsNil() {
		buf = protowire.AppendVarint(buf, 0x1A)
		var err error
		buf, err = proto.MarshalOptions{}.MarshalAppend(buf, r.Data)
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

// UnmarshalProto 手动解码 protobuf 数据到 CommonProtoResp
func (r *CommonResp[T]) UnmarshalProto(data []byte) error {
	var pos int
	for pos < len(data) {
		// 解析标签 (tag 和 wire type)
		tag, n := protowire.ConsumeVarint(data[pos:])
		if n < 0 {
			return errors.New("invalid protobuf data: unable to consume varint")
		}
		pos += n

		fieldNum, wireType := protowire.DecodeTag(tag)
		switch fieldNum {
		case 1: // Code 字段
			if wireType != protowire.VarintType {
				return errors.New("invalid wire type for Code field")
			}
			code, n := protowire.ConsumeVarint(data[pos:])
			if n < 0 {
				return errors.New("invalid protobuf data: unable to consume Code varint")
			}
			r.Code = errorsx.ErrCode(code)
			pos += n

		case 2: // Msg 字段
			if wireType != protowire.BytesType {
				return errors.New("invalid wire type for Msg field")
			}
			msg, n := protowire.ConsumeString(data[pos:])
			if n < 0 {
				return errors.New("invalid protobuf data: unable to consume Msg string")
			}
			r.Msg = msg
			pos += n

		case 3: // Data 字段
			if wireType != protowire.BytesType {
				return errors.New("invalid wire type for Data field")
			}
			if reflect.ValueOf(r.Data).IsNil() {
				r.Data = r.Data.ProtoReflect().New().Interface().(T)
			}
			if len(data[pos:]) > 0 {
				return proto.Unmarshal(data[pos:], r.Data)
			}
			return nil
		}
	}

	return nil
}

type CommonProtoResp = CommonResp[proto.Message]

func NewCommonProtoResp(code errorsx.ErrCode, msg string, data proto.Message) *CommonProtoResp {
	return &CommonProtoResp{Code: code, Msg: msg, Data: data}
}
