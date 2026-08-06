/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"fmt"
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
func NewProductionConfig(appName string) *Config {
	return &Config{
		Name: appName,
		Config: zap.Config{
			Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
			OutputPaths:       []string{stdout},
			DisableCaller:     true,
			DisableStacktrace: true,
			EncoderConfig:     NewProductionEncoderConfig(),
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
		},
	}
}

// NewProductionEncoderConfig creates and returns a new instance.
func NewProductionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        FieldTime,
		LevelKey:       FieldLevel,
		NameKey:        FieldApp,
		CallerKey:      FieldCaller,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     FieldMsg,
		StacktraceKey:  FieldStack,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// NewDevelopmentConfig creates and returns a new instance.
func NewDevelopmentConfig(appName string) *Config {
	return &Config{
		Name: appName,
		Config: zap.Config{
			Development:   true,
			EncoderConfig: NewDevelopmentEncoderConfig(),
			Level:         zap.NewAtomicLevelAt(zapcore.DebugLevel),
			OutputPaths:   []string{stdout},
		},
		EncodeLevelType: EncodeLevelTypeCapitalColor,
	}
}

// NewDevelopmentEncoderConfig creates and returns a new instance.
func NewDevelopmentEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
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
			cores = append(cores, zapcore.NewCore(lc.Encoder, zapcore.AddSync(os.Stdout), StdOutLevel(lc.Config.Level.Level())),
				zapcore.NewCore(lc.Encoder, zapcore.AddSync(os.Stderr), StdErrLevel(lc.Config.Level.Level())))
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
			cores = append(cores, zapcore.NewCore(lc.Encoder, sink, lc.Config.Level.Level()))
		}
	}

	if lc.EnableOtel {
		attrs := append([]attribute.KeyValue{}, lc.Otel.Attributes...)
		attrs = append(attrs, initialFieldsToAttributes(lc.InitialFields)...)
		cores = append(cores, otelzap.NewCore(lc.Name,
			otelzap.WithLoggerProvider(lc.Otel.LoggerProvider),
			otelzap.WithVersion(lc.Otel.Version),
			otelzap.WithSchemaURL(lc.Otel.SchemaURL),
			otelzap.WithAttributes(attrs...),
		))
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
		fs := make([]zap.Field, 0, len(lc.InitialFields))
		keys := make([]string, 0, len(lc.InitialFields))
		for k := range lc.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, lc.InitialFields[k]))
		}
		hooks = append(hooks, zap.Fields(fs...))
	}

	return hooks
}

// initialFieldsToAttributes converts zap.Config.InitialFields into OTel attributes.
// Keys are sorted for stable attribute order.
func initialFieldsToAttributes(fields map[string]any) []attribute.KeyValue {
	if len(fields) == 0 {
		return nil
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attrs := make([]attribute.KeyValue, 0, len(keys))
	for _, k := range keys {
		attrs = append(attrs, anyToAttribute(k, fields[k]))
	}
	return attrs
}

func anyToAttribute(key string, v any) attribute.KeyValue {
	k := attribute.Key(key)
	switch x := v.(type) {
	case nil:
		return k.String("")
	case bool:
		return k.Bool(x)
	case string:
		return k.String(x)
	case int:
		return k.Int(x)
	case int8:
		return k.Int(int(x))
	case int16:
		return k.Int(int(x))
	case int32:
		return k.Int64(int64(x))
	case int64:
		return k.Int64(x)
	case uint:
		return k.Int64(int64(x))
	case uint8:
		return k.Int64(int64(x))
	case uint16:
		return k.Int64(int64(x))
	case uint32:
		return k.Int64(int64(x))
	case uint64:
		return k.Int64(int64(x))
	case float32:
		return k.Float64(float64(x))
	case float64:
		return k.Float64(x)
	case []string:
		return k.StringSlice(x)
	case []bool:
		return k.BoolSlice(x)
	case []int:
		return k.IntSlice(x)
	case []int64:
		return k.Int64Slice(x)
	case []float64:
		return k.Float64Slice(x)
	case fmt.Stringer:
		return k.String(x.String())
	default:
		return k.String(fmt.Sprint(x))
	}
}
