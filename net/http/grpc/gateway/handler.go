package gateway

import (
	"net/http"
	"strconv"

	httpx "github.com/hopeio/gox/net/http"
	grpcx "github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type ProtoMessage[T any] interface {
	*T
	proto.Message
}

func UnaryCall[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp], GprcHandler grpcx.GrpcHandler[Req, Resp, ReqPtr, RespPtr]](gprcHanlder GprcHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req
		var md grpcx.ServerMetadata

		if err := httpx.Bind(r, &req); err != nil {
			HttpError(w, r, err)
			return
		}

		resp, err := gprcHanlder(metadata.NewOutgoingContext(r.Context(), metadata.MD(r.Header)), &req, grpc.Header(&md.Header), grpc.Trailer(&md.Trailer))
		if err != nil {
			HttpError(w, r, err)
			return
		}

		ForwardResponseMessage(w, r, md, resp, DefaultMarshal)
	})
}

func ServerSideStreamCall[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp], GprcHandler grpcx.ServerSideStreamHandler[Req, Resp, ReqPtr, RespPtr]](gprcHanlder GprcHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req
		var err error

		if err = httpx.Bind(r, &req); err != nil {
			HttpError(w, r, err)
			return
		}

		r = r.WithContext(grpc.NewContextWithServerTransportStream(metadata.NewIncomingContext(r.Context(), metadata.MD(r.Header)), &grpcx.ServerTransportStream{}))

		stream := NewServerStream[Resp, RespPtr](w, r)
		defer func() {
			if stream.Status() {
				if err != nil {
					s := status.Code(err)
					w.Header().Set(httpx.HeaderGrpcStatus, strconv.Itoa(int(s)))
					w.Header().Set(httpx.HeaderGrpcMessage, err.Error())
				} else {
					w.Header().Set(httpx.HeaderGrpcStatus, "0")
				}
			}
		}()
		if err = gprcHanlder(&req, stream); err != nil {
			HttpError(w, r, err)
			return
		}
	})
}

func ClientSideStreamCall[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp], GprcHandler grpcx.ClientSideStreamHandler[Req, Resp, ReqPtr, RespPtr]](gprcHanlder GprcHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ts grpcx.ServerTransportStream
		r = r.WithContext(grpc.NewContextWithServerTransportStream(metadata.NewIncomingContext(r.Context(), metadata.MD(r.Header)), &ts))

		stream := NewClientStream[Req, Resp, ReqPtr, RespPtr](w, r)
		if err := gprcHanlder(stream); err != nil {
			HttpError(w, r, err)
			return
		}
		resp, ok := stream.Response()
		if !ok {
			HttpError(w, r, status.Error(codes.Internal, "no response from client streaming handler"))
			return
		}
		ForwardResponseMessage(w, r, ts.ServerMetadata(), resp, DefaultMarshal)
	})
}

func BidiStreamCall[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp], GprcHandler grpcx.BidiStreamHandler[Req, Resp, ReqPtr, RespPtr]](gprcHanlder GprcHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		r = r.WithContext(grpc.NewContextWithServerTransportStream(metadata.NewIncomingContext(r.Context(), metadata.MD(r.Header)), &grpcx.ServerTransportStream{}))

		stream := NewStream[Req, Resp, ReqPtr, RespPtr](w, r)
		defer func() {
			if stream.Status() {
				if err != nil {
					s := status.Code(err)
					w.Header().Set(httpx.HeaderGrpcStatus, strconv.Itoa(int(s)))
					w.Header().Set(httpx.HeaderGrpcMessage, err.Error())
				} else {
					w.Header().Set(httpx.HeaderGrpcStatus, "0")
				}
			}
		}()
		if err = gprcHanlder(stream); err != nil {
			HttpError(w, r, err)
			return
		}
	})
}
