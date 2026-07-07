package gateway

import (
	"io"
	"net/http"

	grpcx "github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ grpc.ServerStream = (*ServerStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty])(nil)
var _ grpcx.ClientSideStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty] = (*ServerStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty])(nil)
var _ grpcx.ServerSideStream[emptypb.Empty, *emptypb.Empty] = (*ServerStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty])(nil)

// ServerStream 服务端 handler 持有的 gRPC stream（grpc.ServerStream 语义）。
// 通过 noRecv / unaryResponse 区分 server streaming、client streaming、bidi。
type ServerStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]] struct {
	streamBase
	closed        bool
	noRecv        bool
	unaryResponse bool
}

func NewServerStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]](w http.ResponseWriter, r *http.Request) *ServerStream[Req, Resp, ReqPtr, RespPtr] {
	return &ServerStream[Req, Resp, ReqPtr, RespPtr]{streamBase: newStreamBase(w, r)}
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) forServerSendOnly() {
	s.noRecv = true
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) forClientRecv() {
	s.unaryResponse = true
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) Send(msg RespPtr) error {
	return s.sendFrame(msg)
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) SetTrailer(md metadata.MD) {
	s.setTrailer(md)
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) Recv() (ReqPtr, error) {
	var msg Req
	if err := s.RecvMsg(&msg); err != nil {
		var zero ReqPtr
		return zero, err
	}
	return &msg, nil
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) RecvMsg(m any) error {
	if s.noRecv {
		return status.Error(codes.Internal, "RecvMsg not supported on server streaming")
	}
	if s.closed {
		return io.EOF
	}
	pm, ok := m.(ReqPtr)
	if !ok {
		return status.Errorf(codes.Internal, "RecvMsg: %T is not proto.Message", m)
	}
	data, err := s.recvFrame()
	if err != nil {
		return err
	}
	return DefaultUnmarshal(s.r.Context(), s.contentType, data, pm)
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) SendAndClose(msg RespPtr) error {
	if s.closed {
		return status.Error(codes.Internal, "SendAndClose called more than once")
	}
	s.closed = true
	if err := s.sendFrame(msg); err != nil {
		return err
	}
	s.finalize(nil)
	return nil
}

func (s *ServerStream[Req, Resp, ReqPtr, RespPtr]) SendMsg(m any) error {
	msg, ok := m.(RespPtr)
	if !ok {
		return status.Errorf(codes.Internal, "SendMsg: unexpected message type %T, expected %T", m, new(Resp))
	}
	if s.unaryResponse {
		return s.SendAndClose(msg)
	}
	return s.Send(msg)
}
