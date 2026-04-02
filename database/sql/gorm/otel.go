package gorm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	sqlx "github.com/hopeio/gox/database/sql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const ScopeName = "github.com/hopeio/gox/database/sql/gorm"

const (
	pluginName    = "otel:gorm"
	startTimeKey  = "otel:gorm:start"
	spanKey       = "otel:gorm:span"
	callbackAfter = "gorm:after_"
)

type OTelPlugin struct {
	tracer        trace.Tracer
	metrics       *GlobalMetrics
	defaultAttrs  []attribute.KeyValue
	customMetrics []CustomMetric
}

type target struct {
	db      *gorm.DB
	attrOpt metric.ObserveOption
}

type GlobalMetrics struct {
	meter    metric.Meter
	mu       sync.RWMutex
	targets  []target
	duration metric.Float64Histogram
	requests metric.Int64Counter
	failures metric.Int64Counter
	rows     metric.Int64Histogram
	inflight metric.Int64UpDownCounter
}

var globalMetrics = sync.OnceValue(func() *GlobalMetrics {
	meter := otel.GetMeterProvider().Meter(ScopeName)
	return &GlobalMetrics{meter: meter}
})

func GlobalGormMetrics() *GlobalMetrics {
	return globalMetrics()
}

type Option func(*OTelPlugin)

type RecordContext struct {
	Ctx        context.Context
	Operation  string
	DB         *gorm.DB
	Attrs      []attribute.KeyValue
	BaseAttrs  []attribute.KeyValue
	ErrorType  string
	Success    bool
	StartTime  time.Time
	DurationMs float64
}

type CustomMetric interface {
	Record(*RecordContext)
}

func WithCustomMetrics(metrics ...CustomMetric) Option {
	return func(p *OTelPlugin) {
		p.customMetrics = append(p.customMetrics, metrics...)
	}
}

func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(p *OTelPlugin) {
		p.defaultAttrs = append(p.defaultAttrs, attrs...)
	}
}

func NewOTelPlugin(opts ...Option) *OTelPlugin {
	p := &OTelPlugin{tracer: otel.Tracer(ScopeName), metrics: GlobalGormMetrics()}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *OTelPlugin) Name() string {
	return pluginName
}

func (p *OTelPlugin) Initialize(db *gorm.DB) error {
	if err := p.initMetrics(db); err != nil {
		return err
	}

	if err := p.registerDBStats(db); err != nil {
		return err
	}
	if err := p.register("create",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Create().Before("gorm:create").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Create().After(callbackAfter+"create").Register(name, fn)
		},
	); err != nil {
		return err
	}
	if err := p.register("query",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Query().Before("gorm:query").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Query().After(callbackAfter+"query").Register(name, fn)
		},
	); err != nil {
		return err
	}
	if err := p.register("update",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Update().Before("gorm:update").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Update().After(callbackAfter+"update").Register(name, fn)
		},
	); err != nil {
		return err
	}
	if err := p.register("delete",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Delete().Before("gorm:delete").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Delete().After(callbackAfter+"delete").Register(name, fn)
		},
	); err != nil {
		return err
	}
	if err := p.register("row",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Row().Before("gorm:row").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Row().After(callbackAfter+"row").Register(name, fn)
		},
	); err != nil {
		return err
	}
	return p.register("raw",
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Raw().Before("gorm:raw").Register(name, fn)
		},
		func(name string, fn func(*gorm.DB)) error {
			return db.Callback().Raw().After(callbackAfter+"raw").Register(name, fn)
		},
	)
}

func (p *OTelPlugin) Close(ctx context.Context) error {
	var err error
	if sqlx.GlobalOTelDBStats() != nil {
		err = errors.Join(err, sqlx.GlobalOTelDBStats().Close())
	}
	return err
}

func (p *OTelPlugin) initMetrics(db *gorm.DB) error {
	if p.metrics == nil {
		p.metrics = GlobalGormMetrics()
	}
	return p.metrics.Register(db, append(p.defaultAttrs, attribute.String("db.system", db.Dialector.Name()))...)
}

func (m *GlobalMetrics) Register(db *gorm.DB, attrs ...attribute.KeyValue) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, target := range m.targets {
		if target.db == db {
			return nil
		}
	}
	m.targets = append(m.targets, target{db: db, attrOpt: metric.WithAttributes(attrs...)})

	if m.duration != nil {
		return nil
	}
	var err error
	m.duration, err = m.meter.Float64Histogram("gorm.db.operation.duration_ms", metric.WithUnit("ms"))
	if err != nil {
		return err
	}
	m.requests, err = m.meter.Int64Counter("gorm.db.operation.requests")
	if err != nil {

		return err
	}
	m.failures, err = m.meter.Int64Counter("gorm.db.operation.failures")
	if err != nil {

		return err
	}
	m.rows, err = m.meter.Int64Histogram("gorm.db.operation.rows_affected")
	if err != nil {

		return err
	}
	m.inflight, err = m.meter.Int64UpDownCounter("gorm.db.operation.inflight")
	if err != nil {
		return err
	}

	return nil
}


func (p *OTelPlugin) registerDBStats(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlx.GlobalOTelDBStats().Register(sqlDB, append(p.defaultAttrs, attribute.String("db.system", db.Dialector.Name()))...)
}

type registerHook func(name string, fn func(*gorm.DB)) error

func (p *OTelPlugin) register(op string, beforeHook, afterHook registerHook) error {
	beforeName := fmt.Sprintf("%s:before_%s", pluginName, op)
	afterName := fmt.Sprintf("%s:after_%s", pluginName, op)
	if err := beforeHook(beforeName, p.before(op)); err != nil {
		return err
	}
	return afterHook(afterName, p.after(op))
}

func (p *OTelPlugin) before(op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		ctx := getContext(tx)
		baseAttrs := p.baseAttrs(op, tx)
		ctx, span := p.tracer.Start(ctx, "gorm."+op, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(baseAttrs...))
		tx.Statement.Context = ctx
		tx.Set(startTimeKey+op, time.Now())
		tx.Set(spanKey+op, span)
		p.metrics.inflight.Add(ctx, 1, metric.WithAttributes(baseAttrs...))
	}
}

func (p *OTelPlugin) after(op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		ctx := getContext(tx)
		errType := errorType(tx.Error)
		baseAttrs := p.baseAttrs(op, tx)
		extraAttrs := p.extraAttrs(errType, tx.Error == nil)
		attrs := append(baseAttrs, extraAttrs...)
		attrOpt := metric.WithAttributes(attrs...)
		baseAttrOpt := metric.WithAttributes(baseAttrs...)
		var start time.Time
		var durationMs float64
		p.metrics.requests.Add(ctx, 1, attrOpt)
		p.metrics.inflight.Add(ctx, -1, baseAttrOpt)
		if tx.Error != nil {
			p.metrics.failures.Add(ctx, 1, attrOpt)
		}
		if tx.RowsAffected >= 0 {
			p.metrics.rows.Record(ctx, tx.RowsAffected, attrOpt)
		}
		if start, ok := getStartTime(tx, op); ok {
			durationMs = float64(time.Since(start)) / float64(time.Millisecond)
			p.metrics.duration.Record(ctx, durationMs, attrOpt)
		}
		p.recordCustomMetrics(&RecordContext{
			Ctx:        ctx,
			Operation:  op,
			DB:         tx,
			Attrs:      attrs,
			BaseAttrs:  baseAttrs,
			ErrorType:  errType,
			Success:    tx.Error == nil,
			StartTime:  start,
			DurationMs: durationMs,
		})
		finishSpan(tx, op, extraAttrs...)
	}
}

func (p *OTelPlugin) recordCustomMetrics(record *RecordContext) {
	for _, cm := range p.customMetrics {
		cm.Record(record)
	}
}

func (p *OTelPlugin) baseAttrs(op string, tx *gorm.DB) []attribute.KeyValue {
	attrs := append([]attribute.KeyValue{}, p.defaultAttrs...)
	attrs = append(attrs, attribute.String("db.operation", op))
	if tx == nil {
		return attrs
	}
	if tx.Dialector != nil {
		attrs = append(attrs, attribute.String("db.system", tx.Dialector.Name()))
	}
	if tx.Statement != nil && tx.Statement.Table != "" {
		attrs = append(attrs, attribute.String("db.table", tx.Statement.Table))
	}
	return attrs
}

func (p *OTelPlugin) extraAttrs(errType string, success bool) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.Bool("db.success", success)}
	if errType != "" {
		attrs = append(attrs, attribute.String("db.error_type", errType))
	}
	return attrs
}

func errorType(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "record_not_found"
	default:
		return "unknown"
	}
}

func getContext(tx *gorm.DB) context.Context {
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		return tx.Statement.Context
	}
	return context.Background()
}

func getStartTime(tx *gorm.DB, op string) (time.Time, bool) {
	if tx == nil {
		return time.Time{}, false
	}
	val, ok := tx.Get(startTimeKey + op)
	if !ok {
		return time.Time{}, false
	}
	start, ok := val.(time.Time)
	return start, ok
}

func finishSpan(tx *gorm.DB, op string, attrs ...attribute.KeyValue) {
	if tx == nil {
		return
	}
	val, ok := tx.Get(spanKey + op)
	if !ok {
		return
	}
	span, ok := val.(trace.Span)
	if !ok || span == nil {
		return
	}
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	if tx.Error != nil {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, tx.Error.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
