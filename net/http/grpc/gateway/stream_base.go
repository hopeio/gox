package gateway

import (
	"context"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type streamBase struct {
	w           http.ResponseWriter
	r           *http.Request
	method      string
	trailers    metadata.MD
	started     bool
	contentType string
}

func newStreamBase(w http.ResponseWriter, r *http.Request) streamBase {
	return streamBase{
		w:           w,
		r:           r,
		contentType: r.Header.Get(httpx.HeaderContentType),
	}
}

func (b *streamBase) Context() context.Context { return b.r.Context() }

func (b *streamBase) Method() string { return b.method }

func (b *streamBase) Trailer() metadata.MD { return b.trailers }

func (b *streamBase) Status() bool { return b.started }

func (b *streamBase) SetHeader(md metadata.MD) error {
	writeResponseHeaderMD(b.w, md)
	return nil
}

func (b *streamBase) SendHeader(md metadata.MD) error {
	_ = b.SetHeader(md)
	return nil
}

func (b *streamBase) setTrailer(md metadata.MD) {
	b.trailers = metadata.Join(b.trailers, md)
	if b.started {
		HandleForwardResponseTrailerHeader(b.w, md)
	}
}

func (b *streamBase) bindContext(ctx context.Context) {
	if b.r != nil {
		b.r = b.r.WithContext(ctx)
	}
}

func (b *streamBase) sendFrame(msg proto.Message) error {
	data, ct, err := DefaultMarshal(b.r.Context(), msg)
	if err != nil {
		return err
	}
	if !b.started {
		b.started = true
		BeginGRPCStream(b.w, b.trailers)
		b.w.Header().Set(httpx.HeaderContentType, ct)
		b.w.WriteHeader(http.StatusOK)
	}
	if err := WriteGRPCFrameData(b.w, data); err != nil {
		return err
	}
	b.w.(http.Flusher).Flush()
	return nil
}

func (b *streamBase) recvFrame() ([]byte, error) {
	return readGRPCFrame(b.r.Body)
}

func (b *streamBase) finalize(err error) {
	FinalizeStreamTrailers(b.w, b.started, err, b.trailers)
}
