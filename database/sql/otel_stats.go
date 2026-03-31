package sql

import (
	"context"
	stdsql "database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OTelDBStats struct {
	prefix string
	meter  metric.Meter

	openConns      metric.Int64ObservableGauge
	inUseConns     metric.Int64ObservableGauge
	idleConns      metric.Int64ObservableGauge
	waitCount      metric.Int64ObservableCounter
	waitDurationMs metric.Float64ObservableGauge
	maxOpenConns   metric.Int64ObservableGauge
	maxIdleClosed  metric.Int64ObservableCounter
	maxLifeClosed  metric.Int64ObservableCounter

	reg metric.Registration
}

func NewOTelDBStats(prefix string, meter metric.Meter) *OTelDBStats {
	return &OTelDBStats{prefix: prefix, meter: meter}
}

func (s *OTelDBStats) Register(db *stdsql.DB, attrs ...attribute.KeyValue) error {
	if err := s.initInstruments(); err != nil {
		return err
	}
	attrOpt := metric.WithAttributes(attrs...)
	reg, err := s.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		st := db.Stats()
		o.ObserveInt64(s.openConns, int64(st.OpenConnections), attrOpt)
		o.ObserveInt64(s.inUseConns, int64(st.InUse), attrOpt)
		o.ObserveInt64(s.idleConns, int64(st.Idle), attrOpt)
		o.ObserveInt64(s.waitCount, st.WaitCount, attrOpt)
		o.ObserveFloat64(s.waitDurationMs, float64(st.WaitDuration)/float64(time.Millisecond), attrOpt)
		o.ObserveInt64(s.maxOpenConns, int64(st.MaxOpenConnections), attrOpt)
		o.ObserveInt64(s.maxIdleClosed, st.MaxIdleClosed, attrOpt)
		o.ObserveInt64(s.maxLifeClosed, st.MaxLifetimeClosed, attrOpt)
		return nil
	},
		s.openConns,
		s.inUseConns,
		s.idleConns,
		s.waitCount,
		s.waitDurationMs,
		s.maxOpenConns,
		s.maxIdleClosed,
		s.maxLifeClosed,
	)
	if err != nil {
		return err
	}
	s.reg = reg
	return nil
}

func (s *OTelDBStats) Close() error {
	if s.reg != nil {
		s.reg.Unregister()
		s.reg = nil
	}
	return nil
}

func (s *OTelDBStats) initInstruments() error {
	var err error
	s.openConns, err = s.meter.Int64ObservableGauge(s.prefix + "db.pool.open_connections")
	if err != nil {
		return err
	}
	s.inUseConns, err = s.meter.Int64ObservableGauge(s.prefix + "db.pool.in_use")
	if err != nil {
		return err
	}
	s.idleConns, err = s.meter.Int64ObservableGauge(s.prefix + "db.pool.idle")
	if err != nil {
		return err
	}
	s.waitCount, err = s.meter.Int64ObservableCounter(s.prefix + "db.pool.wait_count")
	if err != nil {
		return err
	}
	s.waitDurationMs, err = s.meter.Float64ObservableGauge(s.prefix + "db.pool.wait_duration_ms")
	if err != nil {
		return err
	}
	s.maxOpenConns, err = s.meter.Int64ObservableGauge(s.prefix + "db.pool.max_open_connections")
	if err != nil {
		return err
	}
	s.maxIdleClosed, err = s.meter.Int64ObservableCounter(s.prefix + "db.pool.max_idletime_closed")
	if err != nil {
		return err
	}
	s.maxLifeClosed, err = s.meter.Int64ObservableCounter(s.prefix + "db.pool.max_lifetime_closed")
	return err
}
