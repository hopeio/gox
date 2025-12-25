package gateway

import (
	"fmt"
	"net/http"
	"net/textproto"
	"slices"
	"strings"

	"github.com/hopeio/gox/errors"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type CommonResp[T proto.Message] httpx.CommonResp[T]

func (t *CommonResp[T]) MarshalProto() ([]byte, error) {
	buf := make([]byte, 64)
	if t.Code != 0 {
		buf = append(buf, 0x08)
		buf = protowire.AppendVarint(buf, uint64(t.Code))
	}
	if t.Msg != "" {
		buf = append(buf, 0x12)
		buf = protowire.AppendString(buf, t.Msg)
	}
	if t.Data.ProtoReflect().IsValid() {
		buf = append(buf, 0x1a)
		var err error
		buf, err = proto.MarshalOptions{}.MarshalAppend(buf, t.Data)
		if err != nil {
			return nil, err
		}

	}
	return buf, nil
}

func (t *CommonResp[T]) UnmarshalProto(data []byte) error {
	for {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		if num == 0 && typ == protowire.VarintType {
			code, m := protowire.ConsumeVarint(data)
			if m < 0 {
				return fmt.Errorf("invalid varint")
			}
			data = data[m:]
			t.Code = errors.ErrCode(code)
		}
		if num == 1 && typ == protowire.BytesType {
			msg, m := protowire.ConsumeString(data)
			if m < 0 {
				return fmt.Errorf("invalid string")
			}
			data = data[m:]
			t.Msg = msg
		}
		if num == 3 && typ == protowire.BytesType {
			msgData, m := protowire.ConsumeBytes(data)
			if m < 0 {
				return fmt.Errorf("invalid bytes")
			}
			v := t.Data.ProtoReflect().Type().New().Interface()
			return proto.Unmarshal(msgData, v)
		}
	}
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
		rb.Respond(r.Context(), w, r)
		return nil
	case httpx.ResponseBody:
		buf, contentType = rb.ResponseBody()
	case httpx.XXXResponseBody:
		buf, contentType = codec(r, rb.XXX_ResponseBody())
	default:
		buf, contentType = codec(r, message)
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
