package gorm

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestProtoJSONSerializerValueStructObject(t *testing.T) {
	s, err := structpb.NewStruct(map[string]any{"a": "1", "n": 2.0})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ProtoJSONSerializer{}.Value(context.Background(), nil, reflect.Value{}, s)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := raw.([]byte)
	if !ok {
		t.Fatalf("Value type %T", raw)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["a"] != "1" {
		t.Fatalf("got %#v", m)
	}
	rawNil, err := ProtoJSONSerializer{}.Value(context.Background(), nil, reflect.Value{}, (*structpb.Struct)(nil))
	if err != nil {
		t.Fatal(err)
	}
	if rawNil != nil {
		t.Fatalf("nil Struct Value = %#v", rawNil)
	}
}
