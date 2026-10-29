/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	logx "github.com/hopeio/gox/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// pkgDir is this package's source directory; frames below it belong to the hook
// itself and are skipped when resolving the caller.
var pkgDir string

// init initializes package state.
func init() {
	_, file, _, _ := runtime.Caller(0)
	pkgDir = strings.TrimSuffix(file, "access_log.go")
}

// AccessHook logs every Redis command in the same shape as GORM SQL access logs:
// elapsedms / cmd / rows / caller / trace_id / span_id.
type AccessHook struct {
	*zap.Logger
	// SlowThreshold, when > 0, upgrades successful cmds slower than this to Warn
	// (message "SLOW CMD >= …"); normal cmds stay Info. Not a filter — all cmds log.
	SlowThreshold time.Duration
}

// NewAccessHook builds a hook; pass a logger without caller (caller is computed here).
func NewAccessHook(loger *zap.Logger) *AccessHook {
	if loger == nil {
		loger = zap.NewNop()
	}
	return &AccessHook{Logger: loger.With(zap.String("component", "redis"))}
}

// DialHook is a pass-through; connect noise is not useful as access logs.
func (h *AccessHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook times and logs a single command after it finishes.
func (h *AccessHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		begin := time.Now()
		err := next(ctx, cmd)
		h.trace(ctx, begin, cmd, err)
		return err
	}
}

// ProcessPipelineHook times and logs a pipeline as one access line.
func (h *AccessHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		begin := time.Now()
		err := next(ctx, cmds)
		h.tracePipeline(ctx, begin, cmds, err)
		return err
	}
}

func (h *AccessHook) trace(ctx context.Context, begin time.Time, cmd redis.Cmder, err error) {
	ce, elapsed := h.entry(begin, err)
	if ce == nil {
		return
	}
	rows := int64(1)
	if errors.Is(err, redis.Nil) {
		rows = 0
	}
	h.write(ce, ctx, elapsed, formatCmd(cmd), rows)
}

func (h *AccessHook) tracePipeline(ctx context.Context, begin time.Time, cmds []redis.Cmder, err error) {
	ce, elapsed := h.entry(begin, err)
	if ce == nil {
		return
	}
	h.write(ce, ctx, elapsed, formatCmds(cmds), int64(len(cmds)))
}

// entry returns the checked entry for the outcome, or nil when the level is
// disabled — callers then skip formatting the command altogether.
func (h *AccessHook) entry(begin time.Time, err error) (*zapcore.CheckedEntry, time.Duration) {
	elapsed := time.Since(begin)
	level := zapcore.InfoLevel
	var msg string
	switch {
	case err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)):
		// Client disconnect / deadline: not a Redis fault.
		msg = err.Error()
	case err != nil && errors.Is(err, redis.Nil):
		// Key miss is a normal outcome for GET-like cmds.
	case err != nil:
		level = zapcore.ErrorLevel
		msg = err.Error()
	case h.SlowThreshold > 0 && elapsed > h.SlowThreshold:
		level = zapcore.WarnLevel
		msg = "SLOW CMD >= " + h.SlowThreshold.String()
	}
	return h.Check(level, msg), elapsed
}

// write emits one access line with the field shape shared by all commands.
func (h *AccessHook) write(ce *zapcore.CheckedEntry, ctx context.Context, elapsed time.Duration, cmd string, rows int64) {
	ce.Write(
		zap.Float64("elapsedms", float64(elapsed.Nanoseconds())/1e6),
		zap.String("cmd", cmd),
		zap.Int64("rows", rows),
		zap.String("caller", fileWithLineNum()),
		logx.Context(ctx),
	)
}

func formatCmd(cmd redis.Cmder) string { return SafeCommandString(cmd) }

func formatCmds(cmds []redis.Cmder) string { return SafeCommandStrings(cmds) }

// Caps for rendered commands. A SET carrying a megabyte payload must not become a
// megabyte log line, and a single huge argument must not hide the rest of the command.
const (
	maxCmdLen = 1024
	maxArgLen = 256
)

// SafeCommandString renders a command as valid UTF-8 with a bounded length.
//
// Redis arguments are frequently binary - protobuf blobs, compressed values, hashes.
// Writing them verbatim produced access-log lines containing invalid UTF-8: unreadable
// here, and OTLP exporters reject such strings wholesale ("string field contains invalid
// UTF-8"), which surfaced as an OTel SDK error per command. Non-UTF-8 arguments are
// rendered as "<binary:N B>" instead.
func SafeCommandString(cmd redis.Cmder) string {
	var b strings.Builder
	appendCmd(&b, cmd)
	return truncateCmd(b.String())
}

// SafeCommandStrings renders a pipeline (or transaction) as one bounded line.
func SafeCommandStrings(cmds []redis.Cmder) string {
	var b strings.Builder
	for i, cmd := range cmds {
		if i > 0 {
			b.WriteString("; ")
		}
		appendCmd(&b, cmd)
	}
	return truncateCmd(b.String())
}

// truncateCmd cuts s at maxCmdLen on a UTF-8 boundary and marks the cut.
func truncateCmd(s string) string {
	if len(s) <= maxCmdLen {
		return s
	}
	n := maxCmdLen
	for n > 0 && !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n] + "…"
}

// safeArg renders one argument: valid UTF-8, capped at maxArgLen, binary stubbed out.
func safeArg(v any) string {
	switch v := v.(type) {
	case string:
		return safeText(v)
	case []byte:
		// Valid UTF-8 byte slices are still worth showing (keys, JSON, small text
		// values); anything else is a payload whose text form is meaningless.
		return safeText(string(v))
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		return safeText(fmt.Sprint(v))
	}
}

func safeText(s string) string {
	if !utf8.ValidString(s) {
		return "<binary:" + strconv.Itoa(len(s)) + " B>"
	}
	if len(s) <= maxArgLen {
		return s
	}
	n := maxArgLen
	for n > 0 && !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n] + "…"
}

// appendCmd writes "name arg…" avoiding fmt for the types go-redis passes most.
func appendCmd(b *strings.Builder, cmd redis.Cmder) {
	args := cmd.Args()
	if len(args) == 0 {
		b.WriteString(cmd.Name())
		return
	}
	for i, arg := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(safeArg(arg))
	}
}

func fileWithLineNum() string {
	for i := 2; i < 24; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.HasPrefix(file, pkgDir) || strings.Contains(file, "/redis/go-redis/") {
			continue
		}
		if strings.HasSuffix(file, "/runtime/asm_arm64.s") ||
			strings.HasSuffix(file, "/runtime/asm_amd64.s") {
			continue
		}
		return file + ":" + strconv.FormatInt(int64(line), 10)
	}
	return ""
}
