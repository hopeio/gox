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
	"google.golang.org/protobuf/proto"
)

// ResponseStream 供 ForwardResponseMessage 读取 handler 通过 SetTrailer 累积的 metadata。
type ResponseStream interface {
	Trailer() metadata.MD
}

func readGRPCFrame(body io.Reader) ([]byte, error) {
	hdr := make([]byte, 5)
	if _, err := io.ReadFull(body, hdr); err != nil {
		return nil, err
	}
	if hdr[0] != 0 {
		return nil, status.Error(codes.Unimplemented, "compressed frames not supported")
	}
	length := binary.BigEndian.Uint32(hdr[1:5])
	payload := make([]byte, length)
	if _, err := io.ReadFull(body, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// WriteGRPCFrame 将单条消息编码为 gRPC length-prefixed 帧写入 w。
func WriteGRPCFrame(w io.Writer, ctx context.Context, msg proto.Message) error {
	data, _, err := DefaultMarshal(ctx, msg)
	if err != nil {
		return err
	}
	return writeGRPCFrameData(w, data)
}

// WriteGRPCFrameData 写入已编码 payload 的 gRPC 帧。
func WriteGRPCFrameData(w io.Writer, data []byte) error {
	return writeGRPCFrameData(w, data)
}

func writeGRPCFrameData(w io.Writer, data []byte) error {
	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	_, err := w.Write(frame)
	return err
}

// BeginGRPCStream 声明 chunked 流式响应及 trailer 字段名（含 grpc-status / grpc-message）。
func BeginGRPCStream(w http.ResponseWriter, trailers metadata.MD) {
	w.Header().Add(httpx.HeaderTrailer, httpx.HeaderGrpcStatus)
	w.Header().Add(httpx.HeaderTrailer, httpx.HeaderGrpcMessage)
	HandleForwardResponseTrailerHeader(w, trailers)
	w.Header().Set(httpx.HeaderTransferEncoding, "chunked")
}

func writeResponseHeaderMD(w http.ResponseWriter, md metadata.MD) {
	for k, vs := range md {
		for _, v := range vs {
			if strings.HasSuffix(k, "-bin") {
				w.Header().Set(k, base64.StdEncoding.EncodeToString([]byte(v)))
			} else {
				w.Header().Set(k, v)
			}
		}
	}
}
