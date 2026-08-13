/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"log"
	"os"

	netx "github.com/hopeio/gox/net"
	"github.com/hopeio/gox/slices"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sort"
	"strconv"
	"time"
)

const (
	stdout = "stdout"
	stderr = "stderr"
)

const (
	EncodingJson    = "json"
	EncodingConsole = "console"
)

// NewProductionConfig creates and returns a new instance.
// 生产配置为性能禁用 caller/stacktrace，输出到 stdout。
func NewProductionConfig(appName string) *Config {
	c := &Config{
		Name:   appName,
		Config: zap.NewProductionConfig(),
	}
	c.DisableCaller = true
	c.DisableStacktrace = true
	c.OutputPaths = []string{stdout}
	// zap 生产默认 TimeKey 为 "ts"，统一为本库的字段名
	c.EncoderConfig.TimeKey = FieldTime
	c.Init()
	return c
}

// NewDevelopmentConfig creates and returns a new instance.
func NewDevelopmentConfig(appName string) *Config {
	c := &Config{
		Name:            appName,
		Config:          zap.NewDevelopmentConfig(),
		EncodeLevelType: EncodeLevelTypeCapitalColor,
	}
	c.OutputPaths = []string{stdout}
	c.Init()
	return c
}

type ZipConfig = zap.Config

type Config struct {
	Name        string `json:"name,omitempty"` //system name namespace.service
	LevelNumber int    `json:"levelNumber,omitempty"`
	// When EnableOtel is on, logs are also bridged to the OpenTelemetry log pipeline (via otelzap).
	EnableOtel bool
	Otel       OtelConfig `json:"otel,omitempty"`
	zap.Config
	EncodeLevelType string `json:"encodeLevelType,omitempty" comment:"capital;capitalColor;color"`
	TimeLayout      string
	Encoder         zapcore.Encoder
}

type OtelConfig struct {
	Version        string                 `json:"version,omitempty"`
	SchemaURL      string                 `json:"schemaURL,omitempty"`
	Attributes     []attribute.KeyValue   `json:"attributes,omitempty"`
	LoggerProvider otellog.LoggerProvider `json:"loggerProvider,omitempty"`
}

// Init performs the operation.
func (lc *Config) Init() {
	if lc.Name == "" {
		lc.Name = FieldApp
	}

	if lc.Level == (zap.AtomicLevel{}) {
		if lc.LevelNumber == 0 {
			if lc.Development {
				lc.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			} else {
				lc.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			}
		} else {
			lc.Level = zap.NewAtomicLevelAt(zapcore.Level(lc.LevelNumber))
		}
	}

	if !lc.Development {
		if lc.Name != "" && lc.EncoderConfig.NameKey == "" {
			lc.EncoderConfig.NameKey = FieldApp
		}
		if lc.EncoderConfig.EncodeName == nil {
			lc.EncoderConfig.EncodeName = zapcore.FullNameEncoder
		}
		if lc.EncoderConfig.FunctionKey == "" {
			lc.EncoderConfig.FunctionKey = FieldFunc
		}
	}
	if lc.Encoding == "" {
		if lc.Development {
			lc.Encoding = EncodingConsole
		} else {
			lc.Encoding = EncodingJson
		}
	}

	if len(lc.OutputPaths) == 0 && !lc.EnableOtel {
		lc.OutputPaths = []string{stdout}
	}

	if lc.EncoderConfig.TimeKey == "" {
		lc.EncoderConfig.TimeKey = FieldTime
	}

	if lc.EncoderConfig.LevelKey == "" {
		lc.EncoderConfig.LevelKey = FieldLevel
	}

	if lc.EncodeLevelType == "" && lc.Development {
		lc.EncodeLevelType = EncodeLevelTypeCapitalColor
	}

	if lc.EncoderConfig.EncodeLevel == nil {
		var el zapcore.LevelEncoder
		el.UnmarshalText([]byte(lc.EncodeLevelType))
		lc.EncoderConfig.EncodeLevel = el
	}

	if !lc.DisableCaller {
		if lc.EncoderConfig.CallerKey == "" {
			lc.EncoderConfig.CallerKey = FieldCaller
		}

		if lc.EncoderConfig.EncodeCaller == nil {
			if lc.Development {
				lc.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
			} else {
				lc.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
			}
		}
	}

	if !lc.DisableStacktrace {
		if lc.EncoderConfig.StacktraceKey == "" {
			lc.EncoderConfig.StacktraceKey = FieldStack
		}
	}
	if lc.EncoderConfig.MessageKey == "" {
		lc.EncoderConfig.MessageKey = FieldMsg
	}

	if lc.EncoderConfig.LineEnding == "" {
		lc.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	}

	if lc.EncoderConfig.ConsoleSeparator == "" {
		lc.EncoderConfig.ConsoleSeparator = "\t"
	}

	if lc.EncoderConfig.EncodeTime == nil {
		if lc.TimeLayout != "" {
			lc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(lc.TimeLayout)
		} else {
			lc.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
			}
		}
	}
	if lc.EncoderConfig.EncodeDuration == nil {
		lc.EncoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(strconv.FormatInt(d.Nanoseconds()/1e6, 10) + "ms")
		}
	}

	if lc.Sampling != nil && lc.Sampling.Initial == 0 && lc.Sampling.Thereafter == 0 {
		lc.Sampling = nil
	}
}

// NewLogger creates and returns a new instance.
func (lc *Config) NewLogger(cores ...zapcore.Core) *Logger {
	logger := lc.initLogger(cores...)
	// Outside tests, include hostname and IP
	if !lc.Development {
		hostname, _ := os.Hostname()
		logger = logger.With(
			zap.String(FieldHostname, hostname),
			zap.String(FieldIP, netx.ExternalIPStr()),
		)
	}

	return &Logger{logger}
}

// initLogger returns the result.
func (lc *Config) initLogger(cores ...zapcore.Core) *zap.Logger {
	lc.Init()

	if len(lc.OutputPaths) > 0 {
		if lc.Encoder == nil {
			switch lc.Encoding {
			case EncodingJson:
				lc.Encoder = zapcore.NewJSONEncoder(lc.EncoderConfig)
			case EncodingConsole:
				lc.Encoder = zapcore.NewConsoleEncoder(lc.EncoderConfig)
			default:
				log.Fatal("invalid encoder")
			}
		}
		// If both stdout and stderr are set, warn and below go to stdout, error and above to stderr
		ustdout, ustderr := false, false
		consolePaths := make([]string, 0, len(lc.OutputPaths))
		slices.ForEachIndex(lc.OutputPaths, func(i int) {
			switch lc.OutputPaths[i] {
			case stdout:
				ustdout = true
			case stderr:
				ustderr = true
			default:
				consolePaths = append(consolePaths, lc.OutputPaths[i])
			}
		})
		if ustdout && ustderr {
			// 传 AtomicLevel（而非 Level() 静态快照），保证运行期 SetLevel 动态调级生效
			cores = append(cores, zapcore.NewCore(lc.Encoder, zapcore.AddSync(os.Stdout), splitLevel{base: lc.Level, err: false}),
				zapcore.NewCore(lc.Encoder, zapcore.AddSync(os.Stderr), splitLevel{base: lc.Level, err: true}))
		} else {
			if ustdout {
				consolePaths = append(consolePaths, stdout)
			}
			if ustderr {
				consolePaths = append(consolePaths, stderr)
			}
		}
		if len(consolePaths) > 0 {
			sink, _, err := zap.Open(consolePaths...)
			if err != nil {
				log.Fatal(err)
			}
			cores = append(cores, zapcore.NewCore(lc.Encoder, sink, lc.Level))
		}
	}

	if lc.EnableOtel {
		core := otelzap.NewCore(lc.Name,
			otelzap.WithLoggerProvider(lc.Otel.LoggerProvider),
			otelzap.WithVersion(lc.Otel.Version),
			otelzap.WithSchemaURL(lc.Otel.SchemaURL),
			otelzap.WithAttributes(lc.Otel.Attributes...),
		)
		if len(lc.InitialFields) > 0 {
			cores = append(cores, core.With(initialFields(lc.InitialFields)))
		} else {
			cores = append(cores, core)
		}

	}

	//If no output is set, default to the console
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(lc.Encoder, zapcore.AddSync(os.Stdout), lc.Level))
	}

	logger := zap.New(zapcore.NewTee(cores...), lc.hook()...)
	if lc.Name != "" {
		logger = logger.Named(lc.Name)
	}
	return logger
}

// hook returns the result.
func (lc *Config) hook() []zap.Option {
	var hooks []zap.Option

	if len(lc.ErrorOutputPaths) > 0 {
		errSink, _, err := zap.Open(lc.ErrorOutputPaths...)
		if err != nil {
			log.Fatal(err)
		}
		hooks = append(hooks, zap.ErrorOutput(errSink))
	}

	if lc.Development {
		hooks = append(hooks, zap.Development())
	}

	if !lc.DisableCaller {
		hooks = append(hooks, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	if !lc.DisableStacktrace {
		if lc.Development {
			hooks = append(hooks, zap.AddStacktrace(zapcore.DPanicLevel))
		} else {
			hooks = append(hooks, zap.AddStacktrace(zapcore.PanicLevel))
		}
	}
	if scfg := lc.Sampling; scfg != nil {
		hooks = append(hooks, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			var samplerOpts []zapcore.SamplerOption
			if scfg.Hook != nil {
				samplerOpts = append(samplerOpts, zapcore.SamplerHook(scfg.Hook))
			}
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				lc.Sampling.Initial,
				lc.Sampling.Thereafter,
				samplerOpts...,
			)
		}))
	}

	if len(lc.InitialFields) > 0 {
		hooks = append(hooks, zap.Fields(initialFields(lc.InitialFields)...))
	}

	return hooks
}

// splitLevel 组合动态 LevelEnabler（如 zap.AtomicLevel）做 stdout/stderr 分流：
// err=false 放行 [base, Error)，err=true 放行 [Error, ∞) ∩ [base, ∞)。
type splitLevel struct {
	base zapcore.LevelEnabler
	err  bool
}

// Enabled reports whether the given level should be logged by this side of the split.
func (s splitLevel) Enabled(lvl zapcore.Level) bool {
	if s.err {
		return lvl >= zapcore.ErrorLevel && s.base.Enabled(lvl)
	}
	return lvl < zapcore.ErrorLevel && s.base.Enabled(lvl)
}

func initialFields(fields map[string]any) []zapcore.Field {
	fs := make([]zapcore.Field, 0, len(fields))
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fs = append(fs, zap.Any(k, fields[k]))
	}
	return fs
}
