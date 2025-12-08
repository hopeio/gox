package gateway

import (
	"fmt"
	"net/http"
	"net/textproto"
	"slices"
	"strings"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func ForwardResponseMessage(w http.ResponseWriter, r *http.Request, md grpc.ServerMetadata, message proto.Message, marshaler httpx.Marshaler) error {
	HandleForwardResponseServerMetadata(w, md.Header)
	var wantsTrailers bool
	if te := r.Header.Get(httpx.HeaderTE); strings.Contains(strings.ToLower(te), "trailers") {
		wantsTrailers = true
		HandleForwardResponseTrailerHeader(w, md.Trailer)
		w.Header().Set(httpx.HeaderTransferEncoding, "chunked")
	}

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
	case httpx.ResponseBody:
		buf = rb.ResponseBody()
	case httpx.XXXResponseBody:
		buf, err = marshaler.Marshal(rb.XXX_ResponseBody())
	default:
		buf, err = marshaler.Marshal(message)
	}
	if err != nil {
		return err
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
		w.Header().Add("Trailer", tKey)
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
