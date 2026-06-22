package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type ProtoMessage[T any] interface {
	*T
	proto.Message
}

type GrpcHandler[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] func(ctx context.Context, in ReqPtr, opts ...grpc.CallOption) (RespPtr, error)


type ServerSideStreamHandler[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] func(in ReqPtr, stream ServerSideStream[Resp, RespPtr]) error

type ServerSideStream[Resp any, RespPtr ProtoMessage[Resp]] interface {
	Send(RespPtr) error
	grpc.ServerStream
}

type ClientSideStreamHandler[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] func(stream ClientSideStream[Req, Resp, ReqPtr, RespPtr]) error

type ClientSideStream[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] interface {
	Recv() (ReqPtr, error)
	SendAndClose(RespPtr) error
	grpc.ServerStream
}

type BidiStreamHandler[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] func(stream BidiStream[Req, Resp, ReqPtr, RespPtr]) error

type BidiStream[Req, Resp any, ReqPtr ProtoMessage[Req], RespPtr ProtoMessage[Resp]] interface {
	Recv() (ReqPtr, error)
	Send(RespPtr) error
	grpc.ServerStream
}
