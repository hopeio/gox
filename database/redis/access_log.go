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

func formatCmd(cmd redis.Cmder) string {
	var b strings.Builder
	appendCmd(&b, cmd)
	return b.String()
}

func formatCmds(cmds []redis.Cmder) string {
	var b strings.Builder
	for i, cmd := range cmds {
		if i > 0 {
			b.WriteString("; ")
		}
		appendCmd(&b, cmd)
	}
	return b.String()
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
		switch v := arg.(type) {
		case string:
			b.WriteString(v)
		case []byte:
			b.Write(v)
		case int:
			b.WriteString(strconv.Itoa(v))
		case int64:
			b.WriteString(strconv.FormatInt(v, 10))
		case bool:
			b.WriteString(strconv.FormatBool(v))
		default:
			fmt.Fprint(b, arg)
		}
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
