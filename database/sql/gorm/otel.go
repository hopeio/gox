package gorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	sqlx "github.com/hopeio/gox/database/sql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const ScopeName = "github.com/hopeio/gox/database/sql/gorm"

const (
	pluginName   = "otelgorm"
	startTimeKey = "otel:gorm:start:"
)

var (
	firstWordRegex   = regexp.MustCompile(`^\w+`)
	cCommentRegex    = regexp.MustCompile(`(?is)/\*.*?\*/`)
	lineCommentRegex = regexp.MustCompile(`(?im)(?:--|#).*?$`)
	sqlPrefixRegex   = regexp.MustCompile(`^[\s;]*`)

	dbRowsAffected = attribute.Key("db.rows_affected")
)

type OTelPlugin struct {
	provider               trace.TracerProvider
	tracer                 trace.Tracer
	metrics                *GlobalMetrics
	defaultAttrs           []attribute.KeyValue
	customMetrics          []CustomMetric
	excludeQueryVars       bool
	excludeMetrics         bool
	recordStackTraceInSpan bool
	queryFormatter         func(query string) string
	serverAddressProvider  func(dialector gorm.Dialector) string
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

// GlobalGormMetrics returns the result.
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

// WithCustomMetrics updates or inserts a value.
func WithCustomMetrics(metrics ...CustomMetric) Option {
	return func(p *OTelPlugin) {
		p.customMetrics = append(p.customMetrics, metrics...)
	}
}

// WithAttributes updates or inserts a value.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(p *OTelPlugin) {
		p.defaultAttrs = append(p.defaultAttrs, attrs...)
	}
}

// WithTracerProvider updates or inserts a value.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(p *OTelPlugin) {
		p.provider = provider
	}
}

// WithDBSystem updates or inserts a value.
func WithDBSystem(name string) Option {
	return func(p *OTelPlugin) {
		p.defaultAttrs = append(p.defaultAttrs, semconv.DBSystemNameKey.String(name))
	}
}

// WithoutQueryVariables updates or inserts a value.
func WithoutQueryVariables() Option {
	return func(p *OTelPlugin) {
		p.excludeQueryVars = true
	}
}

// WithQueryFormatter updates or inserts a value.
func WithQueryFormatter(fn func(query string) string) Option {
	return func(p *OTelPlugin) {
		p.queryFormatter = fn
	}
}

// WithoutMetrics updates or inserts a value.
func WithoutMetrics() Option {
	return func(p *OTelPlugin) {
		p.excludeMetrics = true
	}
}

// WithRecordStackTrace updates or inserts a value.
func WithRecordStackTrace() Option {
	return func(p *OTelPlugin) {
		p.recordStackTraceInSpan = true
	}
}

// WithServerAddressProvider updates or inserts a value.
func WithServerAddressProvider(fn func(dialector gorm.Dialector) string) Option {
	return func(p *OTelPlugin) {
		p.serverAddressProvider = fn
	}
}

// NewOTelPlugin creates and returns a new instance.
func NewOTelPlugin(opts ...Option) *OTelPlugin {
	p := &OTelPlugin{metrics: GlobalGormMetrics()}
	for _, opt := range opts {
		opt(p)
	}
	if p.provider == nil {
		p.provider = otel.GetTracerProvider()
	}
	p.tracer = p.provider.Tracer(ScopeName)
	return p
}

// Name returns the result.
func (p *OTelPlugin) Name() string {
	return pluginName
}

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

// Initialize performs the operation.
func (p *OTelPlugin) Initialize(db *gorm.DB) error {
	if !p.excludeMetrics {
		if err := p.initMetrics(db); err != nil {
			return err
		}
		if err := p.registerDBStats(db); err != nil {
			return err
		}
	}

	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     func(*gorm.DB)
		name     string
	}{
		{cb.Create().Before("gorm:create"), p.before("gorm.Create", "create"), "before:create"},
		{cb.Create().After("gorm:create"), p.after("create"), "after:create"},
		{cb.Query().Before("gorm:query"), p.before("gorm.Query", "query"), "before:select"},
		{cb.Query().After("gorm:query"), p.after("query"), "after:select"},
		{cb.Delete().Before("gorm:delete"), p.before("gorm.Delete", "delete"), "before:delete"},
		{cb.Delete().After("gorm:delete"), p.after("delete"), "after:delete"},
		{cb.Update().Before("gorm:update"), p.before("gorm.Update", "update"), "before:update"},
		{cb.Update().After("gorm:update"), p.after("update"), "after:update"},
		{cb.Row().Before("gorm:row"), p.before("gorm.Row", "row"), "before:row"},
		{cb.Row().After("gorm:row"), p.after("row"), "after:row"},
		{cb.Raw().Before("gorm:raw"), p.before("gorm.Raw", "raw"), "before:raw"},
		{cb.Raw().After("gorm:raw"), p.after("raw"), "after:raw"},
	}

	var firstErr error
	for _, h := range hooks {
		if err := h.callback.Register("otel:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}
	return firstErr
}

// Close closes and releases resources.
func (p *OTelPlugin) Close(ctx context.Context) error {
	if p.excludeMetrics {
		return nil
	}
	if sqlx.GlobalOTelDBStats() != nil {
		return sqlx.GlobalOTelDBStats().Close()
	}
	return nil
}

// initMetrics performs the operation.
func (p *OTelPlugin) initMetrics(db *gorm.DB) error {
	if p.metrics == nil {
		p.metrics = GlobalGormMetrics()
	}
	attrs := append([]attribute.KeyValue{}, p.defaultAttrs...)
	if sys := dbSystem(db); sys.Valid() {
		attrs = append(attrs, sys)
	}
	return p.metrics.Register(db, attrs...)
}

// Register performs the operation.
func (m *GlobalMetrics) Register(db *gorm.DB, attrs ...attribute.KeyValue) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.targets {
		if t.db == db {
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
	return err
}

// registerDBStats performs the operation.
func (p *OTelPlugin) registerDBStats(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	attrs := append([]attribute.KeyValue{}, p.defaultAttrs...)
	if sys := dbSystem(db); sys.Valid() {
		attrs = append(attrs, sys)
	}
	return sqlx.GlobalOTelDBStats().Register(sqlDB, attrs...)
}

type contextWrapper struct {
	context.Context
	parent context.Context
}

// before performs the operation.
func (p *OTelPlugin) before(spanName, op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		parentCtx := getContext(tx)
		ctx, span := p.tracer.Start(parentCtx, spanName, trace.WithSpanKind(trace.SpanKindClient))
		tx.Statement.Context = contextWrapper{Context: ctx, parent: parentCtx}
		tx.Set(startTimeKey+op, time.Now())

		if p.serverAddressProvider != nil {
			span.SetAttributes(semconv.ServerAddress(p.serverAddressProvider(tx.Config.Dialector)))
		}
		if !p.excludeMetrics && p.metrics != nil && p.metrics.inflight != nil {
			p.metrics.inflight.Add(ctx, 1, metric.WithAttributes(p.metricBaseAttrs(op, tx)...))
		}
	}
}

// after performs the operation.
func (p *OTelPlugin) after(op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		defer func() {
			if c, ok := tx.Statement.Context.(contextWrapper); ok {
				tx.Statement.Context = c.parent
			}
		}()

		ctx := getContext(tx)
		span := trace.SpanFromContext(ctx)
		query := p.queryText(tx)
		operation := dbOperation(query)
		if operation == "" {
			operation = op
		}

		baseAttrs := p.metricBaseAttrs(operation, tx)
		errType := errorType(tx.Error)
		extraAttrs := p.extraAttrs(errType, !isError(tx.Error))
		attrs := append(baseAttrs, extraAttrs...)

		var start time.Time
		var durationMs float64
		if t, ok := getStartTime(tx, op); ok {
			start = t
			durationMs = float64(time.Since(start)) / float64(time.Millisecond)
		}
		if !p.excludeMetrics && p.metrics != nil {
			attrOpt := metric.WithAttributes(attrs...)
			baseAttrOpt := metric.WithAttributes(baseAttrs...)
			p.metrics.requests.Add(ctx, 1, attrOpt)
			if p.metrics.inflight != nil {
				p.metrics.inflight.Add(ctx, -1, baseAttrOpt)
			}
			if isError(tx.Error) {
				p.metrics.failures.Add(ctx, 1, attrOpt)
			}
			if tx.Statement != nil && tx.Statement.RowsAffected >= 0 {
				p.metrics.rows.Record(ctx, tx.Statement.RowsAffected, attrOpt)
			}
			if durationMs > 0 || !start.IsZero() {
				p.metrics.duration.Record(ctx, durationMs, attrOpt)
			}
		}

		p.recordCustomMetrics(&RecordContext{
			Ctx:        ctx,
			Operation:  operation,
			DB:         tx,
			Attrs:      attrs,
			BaseAttrs:  baseAttrs,
			ErrorType:  errType,
			Success:    !isError(tx.Error),
			StartTime:  start,
			DurationMs: durationMs,
		})

		if !span.IsRecording() {
			return
		}
		defer span.End(trace.WithStackTrace(p.recordStackTraceInSpan))

		spanAttrs := make([]attribute.KeyValue, 0, len(p.defaultAttrs)+6)
		spanAttrs = append(spanAttrs, p.defaultAttrs...)
		if sys := dbSystem(tx); sys.Valid() {
			spanAttrs = append(spanAttrs, sys)
		}
		formatQuery := p.formatQuery(query)
		if formatQuery != "" {
			spanAttrs = append(spanAttrs, semconv.DBQueryText(formatQuery))
		}
		spanAttrs = append(spanAttrs, semconv.DBOperationName(operation))
		if table := collectionName(tx); table != "" {
			spanAttrs = append(spanAttrs, semconv.DBCollectionName(table))
			summary := operation + " " + table
			spanAttrs = append(spanAttrs, semconv.DBQuerySummary(summary))
			span.SetName(summary)
		}
		if tx.Statement != nil && tx.Statement.RowsAffected != -1 {
			spanAttrs = append(spanAttrs, dbRowsAffected.Int64(tx.Statement.RowsAffected))
		}
		span.SetAttributes(spanAttrs...)

		if isError(tx.Error) {
			span.RecordError(tx.Error)
			span.SetStatus(codes.Error, tx.Error.Error())
		}
	}
}

// queryText returns the result.
func (p *OTelPlugin) queryText(tx *gorm.DB) string {
	if tx == nil || tx.Statement == nil {
		return ""
	}
	sqlStr := tx.Statement.SQL.String()
	if sqlStr == "" {
		return ""
	}
	if p.excludeQueryVars || tx.Dialector == nil {
		return sqlStr
	}
	return tx.Dialector.Explain(sqlStr, tx.Statement.Vars...)
}

// formatQuery returns the result.
func (p *OTelPlugin) formatQuery(query string) string {
	if p.queryFormatter != nil {
		return p.queryFormatter(query)
	}
	return query
}

// recordCustomMetrics performs the operation.
func (p *OTelPlugin) recordCustomMetrics(record *RecordContext) {
	for _, cm := range p.customMetrics {
		cm.Record(record)
	}
}

// metricBaseAttrs returns the result.
func (p *OTelPlugin) metricBaseAttrs(op string, tx *gorm.DB) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(p.defaultAttrs)+3)
	attrs = append(attrs, p.defaultAttrs...)
	attrs = append(attrs, semconv.DBOperationName(op))
	if sys := dbSystem(tx); sys.Valid() {
		attrs = append(attrs, sys)
	}
	if table := collectionName(tx); table != "" {
		attrs = append(attrs, semconv.DBCollectionName(table))
	}
	return attrs
}

// extraAttrs returns the result.
func (p *OTelPlugin) extraAttrs(errType string, success bool) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.Bool("db.success", success)}
	if errType != "" {
		attrs = append(attrs, attribute.String("db.error_type", errType))
	}
	return attrs
}

// isError reports whether the condition holds.
func isError(err error) bool {
	switch {
	case err == nil,
		errors.Is(err, gorm.ErrRecordNotFound),
		errors.Is(err, driver.ErrSkip),
		errors.Is(err, io.EOF),
		errors.Is(err, sql.ErrNoRows):
		return false
	default:
		return true
	}
}

// errorType returns the result.
func errorType(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "record_not_found"
	case !isError(err):
		return ""
	default:
		return "unknown"
	}
}

// dbSystem returns the result.
func dbSystem(tx *gorm.DB) attribute.KeyValue {
	if tx == nil || tx.Dialector == nil {
		return attribute.KeyValue{}
	}
	switch tx.Dialector.Name() {
	case "mysql":
		return semconv.DBSystemNameMySQL
	case "mssql", "sqlserver":
		return semconv.DBSystemNameMicrosoftSQLServer
	case "postgres", "postgresql":
		return semconv.DBSystemNamePostgreSQL
	case "sqlite":
		return semconv.DBSystemNameSQLite
	case "clickhouse":
		return semconv.DBSystemNameClickHouse
	case "spanner":
		return semconv.DBSystemNameGCPSpanner
	default:
		return attribute.KeyValue{}
	}
}

// dbOperation returns the result.
func dbOperation(query string) string {
	if query == "" {
		return ""
	}
	s := cCommentRegex.ReplaceAllString(query, "")
	s = lineCommentRegex.ReplaceAllString(s, "")
	s = sqlPrefixRegex.ReplaceAllString(s, "")
	return strings.ToLower(firstWordRegex.FindString(s))
}

// collectionName returns the physical table for db.collection.name / span summary.
//
// Prefer the SQL FROM clause (TableExpr / parsed SQL) over Schema.Table: GORM sets
// Schema from the Dest model (e.g. ContentTagRel → content_tag_rel) which often is
// not a real table when using Table("tag a").Joins(...).Find(&dto).
func collectionName(tx *gorm.DB) string {
	if tx == nil || tx.Statement == nil {
		return ""
	}
	stmt := tx.Statement
	if stmt.TableExpr != nil {
		if name := baseTableName(stmt.TableExpr.SQL); name != "" {
			return name
		}
	}
	// Model-based query: Table and Schema.Table match (both the real table).
	if stmt.Schema != nil && stmt.Schema.Table != "" &&
		(stmt.Table == "" || stmt.Table == stmt.Schema.Table) {
		return stmt.Schema.Table
	}
	if sql := stmt.SQL.String(); sql != "" {
		if name := tableFromSQL(sql); name != "" {
			return name
		}
	}
	if name := baseTableName(stmt.Table); name != "" {
		return name
	}
	if stmt.Schema != nil {
		return stmt.Schema.Table
	}
	return ""
}

// fromTableRegexp captures the first identifier after FROM.
var fromTableRegexp = regexp.MustCompile(`(?i)\bFROM\s+(?:(?:\w+|["\x60][^"\x60]+["\x60])\.)?(?:(\w+)|["\x60]([^"\x60]+)["\x60])`)

func tableFromSQL(sql string) string {
	m := fromTableRegexp.FindStringSubmatch(sql)
	if len(m) < 3 {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

// baseTableName strips SQL aliases and quoting: "users AS u", "users u", `users`, schema.users.
func baseTableName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)
	if i := strings.LastIndex(lower, " as "); i >= 0 {
		return unquoteIdent(strings.TrimSpace(name[:i]))
	}
	if fields := strings.Fields(name); len(fields) >= 2 {
		return unquoteIdent(fields[0])
	}
	return unquoteIdent(name)
}

func unquoteIdent(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	if n := len(s); n >= 2 {
		if (s[0] == '`' && s[n-1] == '`') || (s[0] == '"' && s[n-1] == '"') {
			return s[1 : n-1]
		}
	}
	return s
}

// getContext returns the result.
func getContext(tx *gorm.DB) context.Context {
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		return tx.Statement.Context
	}
	return context.Background()
}

// getStartTime performs the operation.
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
