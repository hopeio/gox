package gateway

import (
	"io"
	"net/http"

	grpcx "github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServerTransportStream 双向 streaming；Unary 路径也用它承载 metadata。
type ServerTransportStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]] struct {
	streamBase
	closed bool
}

func NewServerTransportStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]](w http.ResponseWriter, r *http.Request) *ServerTransportStream[Req, Resp, ReqPtr, RespPtr] {
	return &ServerTransportStream[Req, Resp, ReqPtr, RespPtr]{streamBase: newStreamBase(w, r)}
}

func (s *ServerTransportStream[Req, Resp, ReqPtr, RespPtr]) SetTrailer(md metadata.MD) {
	s.setTrailer(md)
}

func (s *ServerTransportStream[Req, Resp, ReqPtr, RespPtr]) Send(msg RespPtr) error {
	return s.sendFrame(msg)
}

func (s *ServerTransportStream[Req, Resp, ReqPtr, RespPtr]) Recv() (ReqPtr, error) {
	var msg Req
	if err := s.RecvMsg(&msg); err != nil {
		var zero ReqPtr
		return zero, err
	}
	return &msg, nil
}

func (s *ServerTransportStream[Req, Resp, ReqPtr, RespPtr]) RecvMsg(m any) error {
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

func (s *ServerTransportStream[Req, Resp, ReqPtr, RespPtr]) SendMsg(m any) error {
	msg, ok := m.(RespPtr)
	if !ok {
		return status.Errorf(codes.Internal, "SendMsg: unexpected message type %T, expected %T", m, new(Resp))
	}
	return s.Send(msg)
}
