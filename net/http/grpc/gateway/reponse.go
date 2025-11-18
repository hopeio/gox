package gateway

import (
	"net/http"

	"github.com/hopeio/gox/log"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(w http.ResponseWriter, r *http.Request, message proto.Message) error {
	var buf []byte
	var err error
	switch rb := message.(type) {
	case http.Handler:
		rb.ServeHTTP(w, r)
		return nil
	case httpx.Responder:
		_, err = rb.Respond(r.Context(), w)
		return err
	case ResponseBody:
		buf = rb.ResponseBody()
	case XXXResponseBody:
		buf, err = JsonPb.Marshal(rb.XXX_ResponseBody())
	default:
		buf, err = JsonPb.Marshal(message)
	}

	if err != nil {
		log.Infof("Marshal error: %v", err)
		return err
	}

	if _, err = w.Write(buf); err != nil {
		log.Infof("Failed to write response: %v", err)
	}
	return nil
}

type XXXResponseBody interface {
	XXX_ResponseBody() interface{}
}

type ResponseBody interface {
	ResponseBody() []byte
}

var OutGoingHeader = []string{
	httpx.HeaderSetCookie,
}
