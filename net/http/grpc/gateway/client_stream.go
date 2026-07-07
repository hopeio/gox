package gateway

import (
	"context"
	"io"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	grpcx "github.com/hopeio/gox/net/http/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ grpc.ClientStream = (*ClientStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty])(nil)

// ClientStream HTTP gateway 客户端流，实现 grpc.ClientStream。
// 通过 net/http.Client 向 gateway 发起 client / server / bidi streaming 请求。
type ClientStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]] struct {
	ctx         context.Context
	contentType string
	client      *http.Client
	req         *http.Request
	reqBody     *io.PipeWriter
	resp        *http.Response
	respBody    io.ReadCloser
	header      metadata.MD
	trailer     metadata.MD
	ready       chan struct{}
	err         error
	sendClosed  bool
}

// NewClientStream 以 pipe 作为请求体启动 HTTP 调用；SendMsg 写帧，CloseSend 结束发送并等待响应可读。
func NewClientStream[Req, Resp any, ReqPtr grpcx.ProtoMessage[Req], RespPtr grpcx.ProtoMessage[Resp]](
	ctx context.Context, client *http.Client, req *http.Request,
) *ClientStream[Req, Resp, ReqPtr, RespPtr] {
	if client == nil {
		client = http.DefaultClient
	}
	pr, pw := io.Pipe()
	req = req.Clone(ctx)
	req.Body = pr
	s := &ClientStream[Req, Resp, ReqPtr, RespPtr]{
		ctx:     ctx,
		client:  client,
		req:     req,
		reqBody: pw,
		ready:   make(chan struct{}),
	}
	go s.begin()
	return s
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) begin() {
	defer close(s.ready)
	s.resp, s.err = s.client.Do(s.req)
	if s.err != nil {
		return
	}
	s.header = metadata.MD(s.resp.Header)
	s.contentType = s.resp.Header.Get(httpx.HeaderContentType)
	s.respBody = s.resp.Body
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) waitReady() error {
	select {
	case <-s.ready:
		return s.err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) Header() (metadata.MD, error) {
	if err := s.waitReady(); err != nil {
		return nil, err
	}
	return s.header, nil
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) Trailer() metadata.MD {
	return s.trailer
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) CloseSend() error {
	if s.sendClosed {
		return nil
	}
	s.sendClosed = true
	return s.reqBody.Close()
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) Context() context.Context {
	return s.ctx
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) Send(msg ReqPtr) error {
	return s.SendMsg(msg)
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) SendMsg(m any) error {
	if s.sendClosed {
		return io.EOF
	}
	msg, ok := m.(proto.Message)
	if !ok {
		return status.Errorf(codes.Internal, "SendMsg: %T is not proto.Message", m)
	}
	return WriteGRPCFrame(s.reqBody, s.ctx, msg)
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) Recv() (RespPtr, error) {
	var resp Resp
	if err := s.RecvMsg(&resp); err != nil {
		var zero RespPtr
		return zero, err
	}
	return &resp, nil
}

func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) RecvMsg(m any) error {
	if err := s.waitReady(); err != nil {
		return err
	}
	data, err := readGRPCFrame(s.respBody)
	if err != nil {
		if err == io.EOF && s.resp != nil {
			s.trailer = metadata.MD(s.resp.Trailer)
		}
		return err
	}
	return DefaultUnmarshal(s.ctx, s.contentType, data, m)
}

// CloseAndRecv 结束发送并读取最终响应（client streaming RPC）。
func (s *ClientStream[Req, Resp, ReqPtr, RespPtr]) CloseAndRecv() (RespPtr, error) {
	if err := s.CloseSend(); err != nil {
		var zero RespPtr
		return zero, err
	}
	return s.Recv()
}
