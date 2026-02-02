package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"reflect"
	"slices"
	"strings"

	errorsx "github.com/hopeio/gox/errors"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type CommonResp[T proto.Message] httpx.CommonResp[T]

func (r *CommonResp[T]) MarshalProto() ([]byte, error) {
	buf := make([]byte, 0, 64)

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
	if r.Code == 0 {
		buf = append(buf, 0)
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
	if data[0] == 0 {
		if reflect.ValueOf(r.Data).IsNil() {
			r.Data = r.Data.ProtoReflect().New().Interface().(T)
		}
		if len(data[pos:]) > 0 {
			return proto.Unmarshal(data[1:], r.Data)
		}
		return nil
	}
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
		}
	}

	return nil
}

type CommonProtoResp = CommonResp[proto.Message]

func NewCommonProtoResp(code errorsx.ErrCode, msg string, data proto.Message) *CommonProtoResp {
	return &CommonProtoResp{Code: code, Msg: msg, Data: data}
}

func ForwardResponseMessage(w http.ResponseWriter, r *http.Request, md grpc.ServerMetadata, message proto.Message, codec httpx.MarshalFunc) error {
	HandleForwardResponseServerMetadata(w, md.Header)
	var wantsTrailers bool
	if te := r.Header.Get(httpx.HeaderTE); strings.Contains(strings.ToLower(te), "trailers") {
		wantsTrailers = true
		HandleForwardResponseTrailerHeader(w, md.Trailer)
		w.Header().Set(httpx.HeaderTransferEncoding, "chunked")
	}
	var contentType string
	var buf []byte
	switch rb := message.(type) {
	case http.Handler:
		rb.ServeHTTP(w, r)
		return nil
	case httpx.Responder:
		rb.Respond(r.Context(), w)
		return nil
	case httpx.ResponseBody:
		buf, contentType = rb.ResponseBody()
	case httpx.XXXResponseBody:
		buf, contentType = codec(r.Context(), rb.XXX_ResponseBody())
	default:
		buf, contentType = codec(r.Context(), message)
	}
	w.Header().Set(httpx.HeaderContentType, contentType)
	ow := w
	if uw, ok := w.(httpx.Unwrapper); ok {
		ow = uw.Unwrap()
	}
	if recorder, ok := ow.(httpx.RecordBody); ok {
		recorder.RecordBody(buf, message)
	}
	w.Write(buf)
	if wantsTrailers {
		HandleForwardResponseTrailer(w, md.Trailer)
	}
	return nil
}

func InComingHeaderMatcher(key string) (string, bool) {
	if slices.Contains(InComingHeader, key) {
		return key, true
	}
	return "", false
}

func OutgoingHeaderMatcher(key string) (string, bool) {
	if slices.Contains(OutgoingHeader, key) {
		return key, true
	}
	return "", false
}

func HandleForwardResponseServerMetadata(w http.ResponseWriter, md metadata.MD) {
	for _, k := range OutgoingHeader {
		if vs, ok := md[k]; ok {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
}

func HandleForwardResponseTrailerHeader(w http.ResponseWriter, md metadata.MD) {
	for k := range md {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", grpc.MetadataTrailerPrefix, k))
		w.Header().Add(httpx.HeaderTrailer, tKey)
	}
}

func HandleForwardResponseTrailer(w http.ResponseWriter, md metadata.MD) {
	for k, vs := range md {
		tKey := fmt.Sprintf("%s%s", grpc.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}
