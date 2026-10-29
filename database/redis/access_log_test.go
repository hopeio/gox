package redis

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
)

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
