package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newJSONHook builds a hook whose lines land in buf, so a test can read the
// rendered record instead of trusting the field list.
func newJSONHook(buf *bytes.Buffer) *AccessHook {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)
	return NewAccessHook(zap.New(core))
}

// An access line carries trace_id/span_id only when the ctx handed to the command
// holds a valid span: logx.Context renders nothing otherwise. Pin both shapes so a
// missing trace id in production reads as "that call had no span on its ctx" rather
// than a broken hook.
func TestAccessHook_TraceFieldsComeFromContext(t *testing.T) {
	const (
		wantTrace = "0af7651916cd43dd8448eb211c80319c"
		wantSpan  = "577586f6ae9d0f4c"
	)
	tid, err := trace.TraceIDFromHex(wantTrace)
	if err != nil {
		t.Fatal(err)
	}
	sid, err := trace.SpanIDFromHex(wantSpan)
	if err != nil {
		t.Fatal(err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	var buf bytes.Buffer
	h := newJSONHook(&buf)
	h.trace(ctx, time.Now(), redis.NewCmd(ctx, "get", "rfv:demo"), nil)

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("decode log line: %v", err)
	}
	if line["trace_id"] != wantTrace {
		t.Fatalf("trace_id=%v want %s (full line %s)", line["trace_id"], wantTrace, buf.Bytes())
	}
	if line["span_id"] != wantSpan {
		t.Fatalf("span_id=%v want %s", line["span_id"], wantSpan)
	}

	// No span on the ctx: the keys stay absent instead of rendering zeros.
	buf.Reset()
	bare := context.Background()
	h.trace(bare, time.Now(), redis.NewCmd(bare, "get", "rfv:demo"), nil)
	line = map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("decode log line: %v", err)
	}
	if _, ok := line["trace_id"]; ok {
		t.Fatalf("a spanless ctx must not emit trace_id, got %s", buf.Bytes())
	}
	if line["cmd"] != "get rfv:demo" {
		t.Fatalf("cmd=%v want 'get rfv:demo'", line["cmd"])
	}
}

func TestSafeCommandStringBinaryArg(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "set", "rfv:hub:instance:1", []byte{0xff, 0x00, 0xfe})
	got := SafeCommandString(cmd)
	if !utf8.ValidString(got) {
		t.Fatalf("rendered command is not valid UTF-8: %q", got)
	}
	if !strings.Contains(got, "<binary:3 B>") {
		t.Fatalf("binary payload must be stubbed, got %q", got)
	}
	if !strings.Contains(got, "rfv:hub:instance:1") {
		t.Fatalf("plain arguments must stay readable, got %q", got)
	}
}

func TestSafeCommandStringInvalidUTF8String(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "set", "k", "bad\xffvalue")
	got := SafeCommandString(cmd)
	if !utf8.ValidString(got) {
		t.Fatalf("rendered command is not valid UTF-8: %q", got)
	}
	if !strings.Contains(got, "<binary:9 B>") {
		t.Fatalf("got %q", got)
	}
}

func TestSafeCommandStringCapsLongArg(t *testing.T) {
	long := strings.Repeat("a", maxArgLen*2)
	got := SafeCommandString(redis.NewCmd(context.Background(), "set", "k", long))
	if len(got) > maxCmdLen+len("…") {
		t.Fatalf("command not capped: %d bytes", len(got))
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("truncation must be marked, got suffix %q", got[len(got)-8:])
	}
}

func TestSafeCommandStringsCapsWholeLine(t *testing.T) {
	cmds := make([]redis.Cmder, 0, 20)
	for i := 0; i < 20; i++ {
		cmds = append(cmds, redis.NewCmd(context.Background(), "set", "k", strings.Repeat("v", maxArgLen)))
	}
	got := SafeCommandStrings(cmds)
	if len(got) > maxCmdLen+len("…") {
		t.Fatalf("pipeline not capped: %d bytes", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatal("pipeline rendering must be valid UTF-8")
	}
}

func TestSafeCommandStringKeepsPlainText(t *testing.T) {
	got := SafeCommandString(redis.NewCmd(context.Background(), "incr", "counter"))
	if got != "incr counter" {
		t.Fatalf("got %q", got)
	}
}
