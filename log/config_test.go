/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package log

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewProductionConfig(t *testing.T) {
	cfg := NewProductionConfig("myapp")
	if cfg.Name != "myapp" {
		t.Fatalf("Name = %q, want myapp", cfg.Name)
	}
	if cfg.Level != zapcore.InfoLevel {
		t.Fatalf("Level = %v, want Info", cfg.Level)
	}
	if !cfg.DisableCaller || !cfg.DisableStacktrace {
		t.Fatal("production config should disable caller and stacktrace")
	}
	if cfg.Sampling == nil || cfg.Sampling.Initial != 100 || cfg.Sampling.Thereafter != 100 {
		t.Fatalf("unexpected sampling: %#v", cfg.Sampling)
	}
	if len(cfg.OutputPaths.Json) != 1 || cfg.OutputPaths.Json[0] != stdout {
		t.Fatalf("OutputPaths.Json = %#v", cfg.OutputPaths.Json)
	}
	if cfg.EncoderConfig.TimeKey != FieldTime {
		t.Fatalf("TimeKey = %q", cfg.EncoderConfig.TimeKey)
	}
}

func TestNewDevelopmentConfig(t *testing.T) {
	cfg := NewDevelopmentConfig("devapp")
	if !cfg.Development {
		t.Fatal("Development should be true")
	}
	if cfg.Level != zapcore.DebugLevel {
		t.Fatalf("Level = %v, want Debug", cfg.Level)
	}
	if cfg.EncodeLevelType != EncodeLevelTypeCapitalColor {
		t.Fatalf("EncodeLevelType = %q", cfg.EncodeLevelType)
	}
	if len(cfg.OutputPaths.Console) != 1 || cfg.OutputPaths.Console[0] != stdout {
		t.Fatalf("OutputPaths.Console = %#v", cfg.OutputPaths.Console)
	}
}

func TestConfigInitDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.Init()
	if cfg.Name != "app" {
		t.Fatalf("Name = %q, want app", cfg.Name)
	}
	if len(cfg.OutputPaths.Console) != 1 || cfg.OutputPaths.Console[0] != stdout {
		t.Fatalf("default console output = %#v", cfg.OutputPaths.Console)
	}
	if cfg.EncoderConfig.TimeKey != FieldTime {
		t.Fatalf("TimeKey = %q", cfg.EncoderConfig.TimeKey)
	}
	if cfg.EncoderConfig.LevelKey != FieldLevel {
		t.Fatalf("LevelKey = %q", cfg.EncoderConfig.LevelKey)
	}
	if cfg.EncoderConfig.MessageKey != FieldMsg {
		t.Fatalf("MessageKey = %q", cfg.EncoderConfig.MessageKey)
	}
	if cfg.EncoderConfig.LineEnding != zapcore.DefaultLineEnding {
		t.Fatal("LineEnding should be zap default")
	}
	if cfg.EncoderConfig.ConsoleSeparator != "\t" {
		t.Fatalf("ConsoleSeparator = %q", cfg.EncoderConfig.ConsoleSeparator)
	}
	if cfg.EncoderConfig.EncodeLevel == nil {
		t.Fatal("EncodeLevel should be set")
	}
	if cfg.EncoderConfig.EncodeTime == nil || cfg.EncoderConfig.EncodeDuration == nil {
		t.Fatal("EncodeTime and EncodeDuration should be set")
	}
}

func TestConfigInitProductionFields(t *testing.T) {
	cfg := &Config{Name: "svc"}
	cfg.Init()
	if cfg.EncoderConfig.NameKey != FieldApp {
		t.Fatalf("NameKey = %q", cfg.EncoderConfig.NameKey)
	}
	if cfg.EncoderConfig.FunctionKey != FieldFunc {
		t.Fatalf("FunctionKey = %q", cfg.EncoderConfig.FunctionKey)
	}
	if cfg.EncoderConfig.CallerKey != FieldCaller {
		t.Fatalf("CallerKey = %q", cfg.EncoderConfig.CallerKey)
	}
	if cfg.EncoderConfig.StacktraceKey != FieldStack {
		t.Fatalf("StacktraceKey = %q", cfg.EncoderConfig.StacktraceKey)
	}
}

func TestConfigInitDevelopmentSkipsProductionNameKey(t *testing.T) {
	cfg := &Config{Development: true, Name: "dev"}
	cfg.Init()
	if cfg.EncoderConfig.NameKey != "" {
		t.Fatalf("development should not auto-set NameKey, got %q", cfg.EncoderConfig.NameKey)
	}
	if cfg.EncodeLevelType != EncodeLevelTypeCapitalColor {
		t.Fatalf("EncodeLevelType = %q", cfg.EncodeLevelType)
	}
}

func TestConfigInitDisableCallerAndStacktrace(t *testing.T) {
	cfg := &Config{DisableCaller: true, DisableStacktrace: true}
	cfg.Init()
	if cfg.EncoderConfig.CallerKey != "" {
		t.Fatalf("CallerKey should stay empty, got %q", cfg.EncoderConfig.CallerKey)
	}
	if cfg.EncoderConfig.StacktraceKey != "" {
		t.Fatalf("StacktraceKey should stay empty, got %q", cfg.EncoderConfig.StacktraceKey)
	}
}

func TestConfigInitTimeLayout(t *testing.T) {
	cfg := &Config{TimeLayout: time.RFC3339}
	cfg.Init()
	if cfg.EncoderConfig.EncodeTime == nil {
		t.Fatal("EncodeTime should be set from TimeLayout")
	}
}

func TestConfigInitClearsZeroSampling(t *testing.T) {
	cfg := &Config{Sampling: &zap.SamplingConfig{}}
	cfg.Init()
	if cfg.Sampling != nil {
		t.Fatal("zero sampling config should be cleared")
	}
	cfg2 := &Config{Sampling: &zap.SamplingConfig{Initial: 10, Thereafter: 0}}
	cfg2.Init()
	if cfg2.Sampling == nil {
		t.Fatal("non-zero Initial should keep sampling")
	}
}

func TestConfigInitLoggerDevelopmentStdout(t *testing.T) {
	cfg := NewDevelopmentConfig("test")
	logger := cfg.initLogger()
	if logger == nil {
		t.Fatal("initLogger returned nil")
	}
	logger.Sync()
}

func TestConfigHookDevelopmentOptions(t *testing.T) {
	cfg := NewDevelopmentConfig("test")
	opts := cfg.hook()
	if len(opts) == 0 {
		t.Fatal("development hook should add options")
	}
}

func TestConfigHookInitialFieldsSorted(t *testing.T) {
	cfg := &Config{
		InitialFields: map[string]interface{}{
			"z": 1,
			"a": 2,
		},
	}
	opts := cfg.hook()
	if len(opts) < 1 {
		t.Fatalf("hook opts len = %d, want at least 1", len(opts))
	}
}

func TestStdOutStdErrLevel(t *testing.T) {
	base := StdOutLevel(zapcore.InfoLevel)
	errBase := StdErrLevel(zapcore.InfoLevel)

	cases := []struct {
		lvl      zapcore.Level
		stdout   bool
		stderr   bool
	}{
		{zapcore.DebugLevel, false, false},
		{zapcore.InfoLevel, true, false},
		{zapcore.WarnLevel, true, false},
		{zapcore.ErrorLevel, false, true},
		{zapcore.FatalLevel, false, true},
	}
	for _, c := range cases {
		if got := base.Enabled(c.lvl); got != c.stdout {
			t.Errorf("StdOutLevel(%v).Enabled(%v) = %v, want %v", base, c.lvl, got, c.stdout)
		}
		if got := errBase.Enabled(c.lvl); got != c.stderr {
			t.Errorf("StdErrLevel(%v).Enabled(%v) = %v, want %v", errBase, c.lvl, got, c.stderr)
		}
	}
}

func TestNewProductionEncoderConfig(t *testing.T) {
	enc := NewProductionEncoderConfig()
	if enc.TimeKey != FieldTime || enc.LevelKey != FieldLevel || enc.MessageKey != FieldMsg {
		t.Fatalf("unexpected encoder config: %#v", enc)
	}
	if enc.FunctionKey != zapcore.OmitKey {
		t.Fatalf("FunctionKey = %q", enc.FunctionKey)
	}
}

func TestNewDevelopmentEncoderConfig(t *testing.T) {
	enc := NewDevelopmentEncoderConfig()
	if enc.TimeKey != "T" || enc.LevelKey != "L" || enc.MessageKey != "M" {
		t.Fatalf("unexpected dev encoder config: %#v", enc)
	}
}
