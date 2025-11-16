package gateway

import (
	"context"
	"net/http"

	"github.com/hopeio/gox/encoding/protobuf/jsonpb"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
	if v, ok := message.(httpx.ICommonRespond); ok {
		_, err := v.CommonRespond(httpx.CommonResponseWriter{ResponseWriter: writer})
		return err
	}
	if v, ok := message.(httpx.IRespond); ok {
		_, err := v.Respond(writer)
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
	httpx.HeaderSetCookie,
}
