package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestClientStreamUploadRoundTrip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := NewServerStream[
			wrapperspb.StringValue, wrapperspb.Int64Value,
			*wrapperspb.StringValue, *wrapperspb.Int64Value,
		](w, r)
		stream.forClientRecv()
		var n int64
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			n += int64(len(msg.GetValue()))
		}
		if err := stream.SendAndClose(wrapperspb.Int64(n)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	stream := NewClientStream[
		wrapperspb.StringValue, wrapperspb.Int64Value,
		*wrapperspb.StringValue, *wrapperspb.Int64Value,
	](req.Context(), srv.Client(), req)

	for _, v := range []string{"ab", "c"} {
		if err := stream.Send(wrapperspb.String(v)); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetValue() != 3 {
		t.Fatalf("sum: got %d want 3", resp.GetValue())
	}
	if got := stream.Trailer().Get("grpc-status"); len(got) > 0 && got[0] != "0" {
		t.Fatalf("grpc-status: %v", got)
	}
}

func TestClientStreamSatisfiesGRPCClientStream(t *testing.T) {
	var s *ClientStream[emptypb.Empty, emptypb.Empty, *emptypb.Empty, *emptypb.Empty]
	var _ grpc.ClientStream = s
}
