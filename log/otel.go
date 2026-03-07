package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
)

func NewOtelLogger(name string, provider *otellog.LoggerProvider) *Logger {
	return &Logger{zap.New(otelzap.NewCore(name, otelzap.WithLoggerProvider(provider)),zap.AddCallerSkip(1))}
}