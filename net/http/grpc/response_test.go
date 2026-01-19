package grpc

import (
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCommonResp(t *testing.T) {
	data, _ := (&CommonProtoResp{
		Code: 1,
		Msg:  "1",
		Data: wrapperspb.Bool(true),
	}).MarshalProto()
	t.Log(data)
	var resp CommonResp[*wrapperspb.BoolValue]
	resp.UnmarshalProto(data)
	t.Log(resp)
}
