package gateway

import (
	"context"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	grpcx "github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/metadata"
)

type streamContextBinder interface {
	bindContext(context.Context)
}

func bindMetadataContext(
	r *http.Request, wHeader, rHeader http.Header, stream streamContextBinder,
) context.Context {
	ctx := NewMetadataContext(r.Context(), wHeader, rHeader)
	stream.bindContext(ctx)
	return ctx
}

func UnaryCall[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]](gprcHanlder grpcx.GrpcHandler[Req, Resp, ReqPtr, RespPtr]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req

		if err := httpx.Bind(r, &req); err != nil {
			HandleError(w, r, err)
			return
		}

		stream := NewServerTransportStream[Req, Resp, ReqPtr, RespPtr](w, r)
		ctx := bindMetadataContext(r, w.Header(), r.Header, stream)

		resp, err := gprcHanlder(ctx, &req)
		if err != nil {
			HandleError(w, r, err)
			return
		}

		HandleResponseMessage(w, r, resp)
	})
}

func ServerSideStreamCall[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp], S grpcx.ServerSideStream[Resp, RespPtr]](gprcHanlder grpcx.ServerSideStreamHandler[Req, Resp, ReqPtr, RespPtr, S]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req
		var err error

		if err = httpx.Bind(r, &req); err != nil {
			HandleError(w, r, err)
			return
		}

		stream := NewServerStream[Req, Resp, ReqPtr, RespPtr](w, r)
		stream.forServerSendOnly()
		r = r.WithContext(bindMetadataContext(r, w.Header(), r.Header, stream))

		defer FinalizeStreamTrailers(w, stream.Status(), err, stream.Trailer())
		if err = gprcHanlder(&req, any(stream).(S)); err != nil {
			HandleError(w, r, err)
			return
		}
	})
}

func ClientSideStreamCall[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp], S grpcx.ClientSideStream[Req, Resp, ReqPtr, RespPtr] ](gprcHanlder grpcx.ClientSideStreamHandler[Req, Resp, ReqPtr, RespPtr, S]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := NewServerStream[Req, Resp, ReqPtr, RespPtr](w, r)
		stream.forClientRecv()
		r = r.WithContext(bindMetadataContext(r, w.Header(), r.Header, stream))

		if err := gprcHanlder(any(stream).(S)); err != nil {
			HandleError(w, r, err)
			return
		}
	})
}

func BidiStreamCall[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp], S grpcx.BidiStream[Req, Resp, ReqPtr, RespPtr], GprcHandler grpcx.BidiStreamHandler[Req, Resp, ReqPtr, RespPtr, S]](gprcHanlder GprcHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		stream := NewServerStream[Req, Resp, ReqPtr, RespPtr](w, r)
		r = r.WithContext(bindMetadataContext(r, w.Header(), r.Header, stream))
		defer FinalizeStreamTrailers(w, stream.Status(), err, stream.Trailer())
		if err = gprcHanlder(any(stream).(S)); err != nil {
			HandleError(w, r, err)
			return
		}
	})
}

// NewMetadataContext 同时设置 incoming 和 outgoing metadata，
// 确保 handler 无论通过 FromIncomingContext 还是 FromOutgoingContext 都能读取请求头。
func NewMetadataContext(ctx context.Context, wHeader, rHeader http.Header) context.Context {
	return metadata.NewOutgoingContext(metadata.NewIncomingContext(ctx, metadata.MD(rHeader)), metadata.MD(wHeader))
}
