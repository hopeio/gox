package gateway

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"strings"

	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Stream 将 HTTP 请求/响应包装为 gRPC stream，统一支持：
//   - server streaming：仅 Send（noRecv）
//   - client streaming：Recv + SendAndClose（bufferResponse）
//   - bidi streaming：Recv + Send
//
// 帧格式与 gRPC-over-HTTP2 一致：
//
//	[compressed-flag: 1 byte] [payload-length: 4 bytes big-endian] [payload: N bytes]
type Stream[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] struct {
	w              http.ResponseWriter
	r              *http.Request
	header         metadata.MD
	trailers       metadata.MD
	started        bool
	contentType    string
	closed         bool
	noRecv         bool
	bufferResponse bool
	response       RespPtr
	hasResp        bool
}

// NewStream 创建双向流适配器。
func NewStream[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]](w http.ResponseWriter, r *http.Request) *Stream[Req, Resp, ReqPtr, RespPtr] {
	return &Stream[Req, Resp, ReqPtr, RespPtr]{
		w:      w,
		r:      r,
		header: metadata.MD{},
	}
}

// NewServerStream 创建 server-side streaming 适配器（请求体已由 Bind 解析，流上只写出响应）。
func NewServerStream[Resp any, RespPtr ProtoMessage[Resp]](w http.ResponseWriter, r *http.Request) *Stream[emptypb.Empty, Resp, *emptypb.Empty, RespPtr] {
	s := NewStream[emptypb.Empty, Resp, *emptypb.Empty, RespPtr](w, r)
	s.noRecv = true
	return s
}

// NewClientStream 创建 client-side streaming 适配器。
func NewClientStream[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]](w http.ResponseWriter, r *http.Request) *Stream[Req, Resp, ReqPtr, RespPtr] {
	s := NewStream[Req, Resp, ReqPtr, RespPtr](w, r)
	s.bufferResponse = true
	return s
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) readFrame() ([]byte, error) {
	hdr := make([]byte, 5)
	if _, err := io.ReadFull(s.r.Body, hdr); err != nil {
		return nil, err
	}
	if hdr[0] != 0 {
		return nil, status.Error(codes.Unimplemented, "compressed frames not supported")
	}
	length := binary.BigEndian.Uint32(hdr[1:5])
	payload := make([]byte, length)
	if _, err := io.ReadFull(s.r.Body, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// Send 将一条消息编码后写入流式响应帧。
func (s *Stream[Req, Resp, ReqPtr, RespPtr]) Send(msg RespPtr) error {
	data, contentType := DefaultMarshal(s.r.Context(), msg)
	if !s.started {
		s.started = true
		s.contentType = contentType
		s.w.Header().Add(httpx.HeaderTrailer, httpx.HeaderGrpcStatus)
		s.w.Header().Add(httpx.HeaderTrailer, httpx.HeaderGrpcMessage)
		s.w.Header().Set(httpx.HeaderContentType, contentType)
		s.w.WriteHeader(http.StatusOK)
	}
	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	if _, err := s.w.Write(frame); err != nil {
		return err
	}
	s.w.(http.Flusher).Flush()
	return nil
}

// Status 返回流式响应是否已启动（至少发送过一帧）。
func (s *Stream[Req, Resp, ReqPtr, RespPtr]) Status() bool {
	return s.started
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) Context() context.Context {
	return s.r.Context()
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) SetHeader(md metadata.MD) error {
	for k, vs := range md {
		for _, v := range vs {
			if strings.HasSuffix(k, "-bin") {
				s.w.Header().Set(k, base64.StdEncoding.EncodeToString([]byte(v)))
			} else {
				s.w.Header().Set(k, v)
			}
		}
	}
	return nil
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) SendHeader(md metadata.MD) error {
	_ = s.SetHeader(md)
	return nil
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) SetTrailer(md metadata.MD) {
	s.trailers = metadata.Join(s.trailers, md)
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) SendMsg(m any) error {
	msg, ok := m.(RespPtr)
	if !ok {
		return status.Errorf(codes.Internal, "SendMsg: unexpected message type %T, expected %T", m, new(Resp))
	}
	if s.bufferResponse {
		if s.hasResp {
			return status.Error(codes.Internal, "SendAndClose called more than once")
		}
		s.response = msg
		s.hasResp = true
		return nil
	}
	return s.Send(msg)
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) Recv() (ReqPtr, error) {
	var msg Req
	if err := s.RecvMsg(&msg); err != nil {
		var zero ReqPtr
		return zero, err
	}
	return &msg, nil
}

func (s *Stream[Req, Resp, ReqPtr, RespPtr]) RecvMsg(m any) error {
	if s.noRecv {
		return status.Error(codes.Internal, "RecvMsg not supported on server streaming")
	}
	if s.closed {
		return io.EOF
	}
	data, err := s.readFrame()
	if err != nil {
		return err
	}
	pm, ok := m.(ReqPtr)
	if !ok {
		return status.Errorf(codes.Internal, "RecvMsg: %T is not proto.Message", m)
	}
	return Unmarshaller(s.r.Context(), s.r.Header.Get(httpx.HeaderContentType), data, pm)
}

// SendAndClose 提交 client streaming 的最终响应（由 gateway 层 ForwardResponseMessage 写出）。
func (s *Stream[Req, Resp, ReqPtr, RespPtr]) SendAndClose(msg RespPtr) error {
	if s.hasResp {
		return status.Error(codes.Internal, "SendAndClose called more than once")
	}
	s.closed = true
	return s.SendMsg(msg)
}

// Response 返回 SendAndClose 缓存的响应。
func (s *Stream[Req, Resp, ReqPtr, RespPtr]) Response() (RespPtr, bool) {
	return s.response, s.hasResp
}
