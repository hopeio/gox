package gateway

import (
	"context"
	"github.com/hopeio/gox/encoding/protobuf/jsonpb"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/consts"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
	"net/http"
)

func Response(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
	if v, ok := message.(httpx.ICommonResponseTo); ok {
		_, err := v.CommonResponse(httpx.CommonResponseWriter{writer})
		return err
	}
	if v, ok := message.(httpx.IHttpResponseTo); ok {
		_, err := v.Response(writer)
		return err
	}
	var buf []byte
	var err error
	switch rb := message.(type) {
	case responseBody:
		buf, err = jsonpb.JsonPb.Marshal(rb.ResponseBody())
	case xxxResponseBody:
		buf, err = jsonpb.JsonPb.Marshal(rb.XXX_ResponseBody())
	default:
		buf, err = jsonpb.JsonPb.Marshal(message)
	}

	if err != nil {
		grpclog.Infof("Marshal error: %v", err)
		return err
	}

	if _, err = writer.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}
	return nil
}

type xxxResponseBody interface {
	XXX_ResponseBody() interface{}
}

type responseBody interface {
	ResponseBody() interface{}
}

var OutGoingHeader = []string{
	consts.HeaderSetCookie,
}
