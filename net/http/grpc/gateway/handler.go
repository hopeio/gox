package gateway

import (
	"context"
	"net/http"
	"reflect"
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

		resp, err := gprcHanlder(newMetadataContext(r.Context(), w.Header(), r.Header), &req)
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

		r = r.WithContext(grpc.NewContextWithServerTransportStream(newMetadataContext(r.Context(), w.Header(), r.Header), &grpcx.ServerTransportStream{}))

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
		r = r.WithContext(grpc.NewContextWithServerTransportStream(newMetadataContext(r.Context(), w.Header(), r.Header), &ts))

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

		r = r.WithContext(grpc.NewContextWithServerTransportStream(newMetadataContext(r.Context(), w.Header(), r.Header), &grpcx.ServerTransportStream{}))

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

// newMetadataContext 同时设置 incoming 和 outgoing metadata，
// 确保 handler 无论通过 FromIncomingContext 还是 FromOutgoingContext 都能读取请求头。
func newMetadataContext(ctx context.Context, wHeader,rheader http.Header) context.Context {
	return metadata.NewOutgoingContext(metadata.NewIncomingContext(ctx, metadata.MD(rheader)), metadata.MD(wHeader))
}

// ---------------------------------------------------------------------------
// RegisterGRPC — 将 gRPC 服务的所有方法转换为 HTTP 路由并注册到 ServeMux
// ---------------------------------------------------------------------------

// RegisterGRPC 根据 grpc.ServiceDesc 中描述的 unary 和 streaming 方法，
// 将每一个 gRPC 方法注册为 POST /package.Service/Method 格式的 HTTP 路由。
//
// 参数:
//   - mux: HTTP 路由注册目标
//   - server: gRPC 服务实现（必须是指针类型，与 RegisterService 时传入的一致）
//   - desc: gRPC 服务描述符（由 protoc-gen-go-grpc 生成的 _ServiceDesc 变量）
//
// 使用示例:
//
//	gateway.RegisterGRPC(mux, myServiceServerImpl, content.ContentService_ServiceDesc)
func RegisterGRPC[T any](mux *http.ServeMux, server T, desc grpc.ServiceDesc) {
	// unary methods
	for i := range desc.Methods {
		md := &desc.Methods[i]
		path := "/" + desc.ServiceName + "/" + md.MethodName
		reqType := extractReqType(md.Handler)
		if reqType == nil {
			continue
		}
		registerUnaryRoute(mux, path, server, md.Handler, reqType)
	}

	// streaming methods
	for i := range desc.Streams {
		sd := &desc.Streams[i]
		path := "/" + desc.ServiceName + "/" + sd.StreamName
		registerStreamRoute(mux, path, server, sd.Handler)
	}
}

// registerUnaryRoute 为单个 unary RPC 方法注册 HTTP handler。
// 通过反射从 grpc.MethodHandler 签名中提取请求类型，
// 使用 httpx.Bind 解析请求体后直接调用 MethodHandler。
func registerUnaryRoute(mux *http.ServeMux, path string, server any, handler grpc.MethodHandler, reqType reflect.Type) {
	mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		req := reflect.New(reqType).Interface()
		var md grpcx.ServerMetadata

		if err := httpx.Bind(r, req); err != nil {
			HttpError(w, r, err)
			return
		}

		ctx := newMetadataContext(r.Context(), w.Header(), r.Header)
		dec := func(v any) error { return nil } // 请求体已由 Bind 解析

		resp, err := handler(server, ctx, dec, nil)
		if err != nil {
			HttpError(w, r, err)
			return
		}

		ForwardResponseMessage(w, r, md, resp.(proto.Message), DefaultMarshal)
	})
}

// registerStreamRoute 为单个 streaming RPC 方法注册 HTTP handler。
// StreamHandler 签名为 func(srv any, stream grpc.ServerStream) error，
// 因此只需提供一个实现了 grpc.ServerStream 的适配器即可。
// 生成的 _Handler 会将该适配器包装为具体的 typed stream（如 ReadFileServer），
// typed 的 Send/Recv 最终会委托到适配器的 SendMsg/RecvMsg。
func registerStreamRoute(mux *http.ServeMux, path string, server any, handler grpc.StreamHandler) {
	mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		var err error

		r = r.WithContext(grpc.NewContextWithServerTransportStream(
			newMetadataContext(r.Context(), w.Header(), r.Header),
			&grpcx.ServerTransportStream{},
		))

		stream := newGatewayServerStream(w, r)
		defer func() {
			if stream.started {
				if err != nil {
					w.Header().Set(httpx.HeaderGrpcStatus, strconv.Itoa(int(status.Code(err))))
					w.Header().Set(httpx.HeaderGrpcMessage, err.Error())
				} else {
					w.Header().Set(httpx.HeaderGrpcStatus, "0")
				}
			}
		}()

		if err = handler(server, stream); err != nil {
			if !stream.started {
				HttpError(w, r, err)
			}
			return
		}
	})
}

// extractReqType 从 grpc.MethodHandler 的函数签名中提取请求结构体类型。
// handler 签名: func(srv any, ctx context.Context, dec func(ReqPtr) error, interceptor) (any, error)
func extractReqType(handler grpc.MethodHandler) reflect.Type {
	t := reflect.TypeOf(handler)
	if t == nil || t.Kind() != reflect.Func || t.NumIn() < 3 {
		return nil
	}
	decType := t.In(2) // dec func(ReqPtr) error
	if decType.Kind() != reflect.Func || decType.NumIn() != 1 {
		return nil
	}
	return decType.In(0).Elem() // *Req → Req
}

// ---------------------------------------------------------------------------
// gatewayServerStream — 实现 grpc.ServerStream，将 gRPC 流式操作适配到 HTTP
// ---------------------------------------------------------------------------

// gatewayServerStream 实现了 grpc.ServerStream 接口，将 Send/Recv 等操作
// 转发到 gateway 的 Stream 适配器。它会被生成的 _Handler 函数包装为
// 具体的 typed stream adapter（如 rfvServiceReadFileServer），
// typed adapter 的 Send/Recv 内部调用 SendMsg/RecvMsg，最终由此处处理。
type gatewayServerStream struct {
	w       http.ResponseWriter
	r       *http.Request
	header  metadata.MD
	trailer metadata.MD
	started bool
}

func newGatewayServerStream(w http.ResponseWriter, r *http.Request) *gatewayServerStream {
	return &gatewayServerStream{
		w:      w,
		r:      r,
		header: metadata.MD{},
	}
}

func (s *gatewayServerStream) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *gatewayServerStream) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *gatewayServerStream) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *gatewayServerStream) Context() context.Context {
	return s.r.Context()
}

func (s *gatewayServerStream) SendMsg(m any) error {
	msg, ok := m.(proto.Message)
	if !ok {
		return status.Errorf(codes.Internal, "SendMsg: %T is not proto.Message", m)
	}
	data, contentType := DefaultMarshal(s.r.Context(), msg)
	if !s.started {
		s.started = true
		s.w.Header().Set(httpx.HeaderContentType, contentType)
		s.w.WriteHeader(http.StatusOK)
	}
	if _, err := s.w.Write(data); err != nil {
		return err
	}
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func (s *gatewayServerStream) RecvMsg(m any) error {
	pm, ok := m.(proto.Message)
	if !ok {
		return status.Errorf(codes.Internal, "RecvMsg: %T is not proto.Message", m)
	}
	if err := httpx.Bind(s.r, pm); err != nil {
		return status.Errorf(codes.Internal, "RecvMsg: %v", err)
	}
	return nil
}
