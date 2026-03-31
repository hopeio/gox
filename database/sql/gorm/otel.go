package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	tracer   trace.Tracer
	meter    metric.Meter
	defaultAttrs []attribute.KeyValue
	duration metric.Float64Histogram
	requests metric.Int64Counter
	failures metric.Int64Counter
	rows     metric.Int64Histogram
	inflight metric.Int64UpDownCounter

	dbOpenConns        metric.Int64ObservableGauge
	dbInUseConns       metric.Int64ObservableGauge
	dbIdleConns        metric.Int64ObservableGauge
	dbWaitCount        metric.Int64ObservableCounter
	dbWaitDurationMs   metric.Float64ObservableGauge
	dbMaxOpenConns     metric.Int64ObservableGauge
	dbMaxIdleClosed    metric.Int64ObservableCounter
	dbMaxLifeClosed    metric.Int64ObservableCounter
	dbStatsReg         metric.Registration
	customMetrics      []CustomMetric
	collectors         []Collector
}

type Option func(*OTelPlugin)

type RecordContext struct {
	Ctx        context.Context
	Operation  string
	Tx         *gorm.DB
	Attrs      []attribute.KeyValue
	BaseAttrs  []attribute.KeyValue
	ErrorType  string
	Success    bool
	Start      time.Time
	DurationMs float64
}

type CustomMetric interface {
	Init(metric.Meter) error
	Record(*RecordContext)
}

type Collector interface {
	Init(db *gorm.DB, meter metric.Meter) error
	Close(context.Context) error
}

func WithCustomMetrics(metrics ...CustomMetric) Option {
	return func(p *OTelPlugin) {
		p.customMetrics = append(p.customMetrics, metrics...)
	}
}

func WithCollectors(collectors ...Collector) Option {
	return func(p *OTelPlugin) {
		p.collectors = append(p.collectors, collectors...)
	}
}

func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(p *OTelPlugin) {
		p.defaultAttrs = append(p.defaultAttrs, attrs...)
	}
}

func NewOTelPlugin(opts ...Option) *OTelPlugin {
	p := &OTelPlugin{tracer: otel.Tracer(ScopeName), meter: otel.Meter(ScopeName)}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *OTelPlugin) Name() string {
	return pluginName
}


func (p *OTelPlugin) Initialize(db *gorm.DB) error {
	if err := p.initMetrics(); err != nil {
		return err
	}
	if err := p.initCustomMetrics(); err != nil {
		return err
	}
	if err := p.registerDBStats(db); err != nil {
		return err
	}
	if err := p.initCollectors(db); err != nil {
		_ = p.Close(context.Background())
		return err
	}
	if err := p.register("create",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Create().Before("gorm:create").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Create().After(callbackAfter+"create").Register(name, fn) },
	); err != nil {
		return err
	}
	if err := p.register("query",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Query().Before("gorm:query").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Query().After(callbackAfter+"query").Register(name, fn) },
	); err != nil {
		return err
	}
	if err := p.register("update",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Update().Before("gorm:update").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Update().After(callbackAfter+"update").Register(name, fn) },
	); err != nil {
		return err
	}
	if err := p.register("delete",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Delete().Before("gorm:delete").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Delete().After(callbackAfter+"delete").Register(name, fn) },
	); err != nil {
		return err
	}
	if err := p.register("row",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Row().Before("gorm:row").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Row().After(callbackAfter+"row").Register(name, fn) },
	); err != nil {
		return err
	}
	return p.register("raw",
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Raw().Before("gorm:raw").Register(name, fn) },
		func(name string, fn func(*gorm.DB)) error { return db.Callback().Raw().After(callbackAfter+"raw").Register(name, fn) },
	)
}

func (p *OTelPlugin) initCollectors(db *gorm.DB) error {
	for _, c := range p.collectors {
		if err := c.Init(db, p.meter); err != nil {
			return err
		}
	}
	return nil
}

func (p *OTelPlugin) Close(ctx context.Context) error {
	var err error
	if p.dbStatsReg != nil {
		p.dbStatsReg.Unregister()
		p.dbStatsReg = nil
	}
	for _, c := range p.collectors {
		err = errors.Join(err, c.Close(ctx))
	}
	return err
}

func (p *OTelPlugin) initMetrics() error {
	var err error
	p.duration, err = p.meter.Float64Histogram("gorm.db.operation.duration_ms", metric.WithUnit("ms"))
	if err != nil {
		return err
	}
	p.requests, err = p.meter.Int64Counter("gorm.db.operation.requests")
	if err != nil {
		return err
	}
	p.failures, err = p.meter.Int64Counter("gorm.db.operation.failures")
	if err != nil {
		return err
	}
	p.rows, err = p.meter.Int64Histogram("gorm.db.operation.rows_affected")
	if err != nil {
		return err
	}
	p.inflight, err = p.meter.Int64UpDownCounter("gorm.db.operation.inflight")
	if err != nil {
		return err
	}
	return p.initDBStatsInstruments()
}

func (p *OTelPlugin) initCustomMetrics() error {
	for _, cm := range p.customMetrics {
		if err := cm.Init(p.meter); err != nil {
			return err
		}
	}
	return nil
}

func (p *OTelPlugin) initDBStatsInstruments() error {
	var err error
	p.dbOpenConns, err = p.meter.Int64ObservableGauge("gorm.db.pool.open_connections")
	if err != nil {
		return err
	}
	p.dbInUseConns, err = p.meter.Int64ObservableGauge("gorm.db.pool.in_use")
	if err != nil {
		return err
	}
	p.dbIdleConns, err = p.meter.Int64ObservableGauge("gorm.db.pool.idle")
	if err != nil {
		return err
	}
	p.dbWaitCount, err = p.meter.Int64ObservableCounter("gorm.db.pool.wait_count")
	if err != nil {
		return err
	}
	p.dbWaitDurationMs, err = p.meter.Float64ObservableGauge("gorm.db.pool.wait_duration_ms")
	if err != nil {
		return err
	}
	p.dbMaxOpenConns, err = p.meter.Int64ObservableGauge("gorm.db.pool.max_open_connections")
	if err != nil {
		return err
	}
	p.dbMaxIdleClosed, err = p.meter.Int64ObservableCounter("gorm.db.pool.max_idletime_closed")
	if err != nil {
		return err
	}
	p.dbMaxLifeClosed, err = p.meter.Int64ObservableCounter("gorm.db.pool.max_lifetime_closed")
	return err
}

func (p *OTelPlugin) registerDBStats(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	attrOpt := metric.WithAttributes(attribute.String("db.system", db.Dialector.Name()))
	reg, err := p.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		st := sqlDB.Stats()
		o.ObserveInt64(p.dbOpenConns, int64(st.OpenConnections), attrOpt)
		o.ObserveInt64(p.dbInUseConns, int64(st.InUse), attrOpt)
		o.ObserveInt64(p.dbIdleConns, int64(st.Idle), attrOpt)
		o.ObserveInt64(p.dbWaitCount, st.WaitCount, attrOpt)
		o.ObserveFloat64(p.dbWaitDurationMs, float64(st.WaitDuration)/float64(time.Millisecond), attrOpt)
		o.ObserveInt64(p.dbMaxOpenConns, int64(st.MaxOpenConnections), attrOpt)
		o.ObserveInt64(p.dbMaxIdleClosed, st.MaxIdleClosed, attrOpt)
		o.ObserveInt64(p.dbMaxLifeClosed, st.MaxLifetimeClosed, attrOpt)
		return nil
	},
		p.dbOpenConns,
		p.dbInUseConns,
		p.dbIdleConns,
		p.dbWaitCount,
		p.dbWaitDurationMs,
		p.dbMaxOpenConns,
		p.dbMaxIdleClosed,
		p.dbMaxLifeClosed,
	)
	if err != nil {
		return err
	}
	p.dbStatsReg = reg
	return nil
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
		attrs := p.attrsFromBase(baseAttrs, "", true)
		ctx, span := p.tracer.Start(ctx, "gorm."+op, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
		tx.Statement.Context = ctx
		tx.InstanceSet(startTimeKey+op, time.Now())
		tx.InstanceSet(spanKey+op, span)
		p.inflight.Add(ctx, 1, metric.WithAttributes(baseAttrs...))
	}
}

func (p *OTelPlugin) after(op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		ctx := getContext(tx)
		errType := errorType(tx.Error)
		baseAttrs := p.baseAttrs(op, tx)
		attrs := p.attrsFromBase(baseAttrs, errType, tx.Error == nil)
		attrOpt := metric.WithAttributes(attrs...)
		baseAttrOpt := metric.WithAttributes(baseAttrs...)
		var start time.Time
		var durationMs float64
		p.requests.Add(ctx, 1, attrOpt)
		p.inflight.Add(ctx, -1, baseAttrOpt)
		if tx.Error != nil {
			p.failures.Add(ctx, 1, attrOpt)
		}
		if tx.RowsAffected >= 0 {
			p.rows.Record(ctx, tx.RowsAffected, attrOpt)
		}
		if start, ok := getStartTime(tx, op); ok {
			durationMs = float64(time.Since(start)) / float64(time.Millisecond)
			p.duration.Record(ctx, durationMs, attrOpt)
		}
		p.recordCustomMetrics(&RecordContext{
			Ctx:        ctx,
			Operation:  op,
			Tx:         tx,
			Attrs:      attrs,
			BaseAttrs:  baseAttrs,
			ErrorType:  errType,
			Success:    tx.Error == nil,
			Start:      start,
			DurationMs: durationMs,
		})
		finishSpan(tx, op)
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

func (p *OTelPlugin) attrsFor(op string, tx *gorm.DB, errType string, success bool) []attribute.KeyValue {
	return p.attrsFromBase(p.baseAttrs(op, tx), errType, success)
}

func (p *OTelPlugin) attrsFromBase(baseAttrs []attribute.KeyValue, errType string, success bool) []attribute.KeyValue {
	attrs := append(baseAttrs, attribute.Bool("db.success", success))
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
	val, ok := tx.InstanceGet(startTimeKey + op)
	if !ok {
		return time.Time{}, false
	}
	start, ok := val.(time.Time)
	return start, ok
}

func finishSpan(tx *gorm.DB, op string) {
	if tx == nil {
		return
	}
	val, ok := tx.InstanceGet(spanKey + op)
	if !ok {
		return
	}
	span, ok := val.(trace.Span)
	if !ok || span == nil {
		return
	}
	if tx.Error != nil {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, tx.Error.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
