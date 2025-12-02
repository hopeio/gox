package gateway

import (
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(w http.ResponseWriter, r *http.Request, md grpc.ServerMetadata, message proto.Message, marshaler httpx.Marshaler) error {
	HandleForwardResponseServerMetadata(w, md.HeaderMD)
	HandleForwardResponseTrailerHeader(w, md.TrailerMD)

	contentType := marshaler.ContentType(message)
	w.Header().Set(httpx.HeaderContentType, contentType)

	var buf []byte
	var err error
	switch rb := message.(type) {
	case http.Handler:
		rb.ServeHTTP(w, r)
		return nil
	case httpx.Responder:
		rb.Respond(r.Context(), w)
		return nil
	case httpx.CommonResponder:
		rb.CommonRespond(r.Context(), httpx.ResponseWriterWrapper{ResponseWriter: w})
		return nil
	case ResponseBody:
		buf = rb.ResponseBody()
	case XXXResponseBody:
		buf, err = marshaler.Marshal(rb.XXX_ResponseBody())
	default:
		buf, err = marshaler.Marshal(message)
	}
	if err != nil {
		return err
	}
	w.Write(buf)
	return nil
}

type XXXResponseBody interface {
	XXX_ResponseBody() interface{}
}

type ResponseBody interface {
	ResponseBody() []byte
}
