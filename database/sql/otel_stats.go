package sql

import (
	"context"
	stdsql "database/sql"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const ScopeName = "github.com/hopeio/gox/database/sql"

type OTelDBStats struct {
	meter metric.Meter

	openConns      metric.Int64ObservableGauge
	inUseConns     metric.Int64ObservableGauge
	idleConns      metric.Int64ObservableGauge
	waitCount      metric.Int64ObservableCounter
	waitDurationMs metric.Float64ObservableGauge
	maxOpenConns   metric.Int64ObservableGauge
	maxIdleClosed  metric.Int64ObservableCounter
	maxLifeClosed  metric.Int64ObservableCounter

	mu      sync.RWMutex
	targets []target
	reg     metric.Registration
}


type target struct {
	db *stdsql.DB
	attrOpt metric.ObserveOption
}

var globalOTelDBStats = sync.OnceValue(func() *OTelDBStats {
	meter := otel.GetMeterProvider().Meter(ScopeName)
	return &OTelDBStats{meter: meter, targets: make([]target, 0)}
})

func GlobalOTelDBStats() *OTelDBStats {
	return globalOTelDBStats()
}

func (s *OTelDBStats) Register(db *stdsql.DB, attrs ...attribute.KeyValue) error {
	if db == nil {
		return stdsql.ErrConnDone
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, target := range s.targets {
		if target.db == db {
			return nil
		}
	}
	s.targets = append(s.targets, target{db: db, attrOpt: metric.WithAttributes(attrs...)})
	if s.reg != nil {
		return nil
	}
	if err := s.initInstruments(); err != nil {
		return err
	}
	reg, err := s.meter.RegisterCallback(s.observe, s.openConns, s.inUseConns, s.idleConns, s.waitCount, s.waitDurationMs, s.maxOpenConns, s.maxIdleClosed, s.maxLifeClosed)
	if err != nil {
		return err
	}
	s.reg = reg
	return nil
}

func (s *OTelDBStats) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reg != nil {
		s.reg.Unregister()
		s.reg = nil
	}
	clear(s.targets)
	return nil
}

func (s *OTelDBStats) observe(_ context.Context, o metric.Observer) error {
	for _, target := range s.targets {
		st := target.db.Stats()
		o.ObserveInt64(s.openConns, int64(st.OpenConnections), target.attrOpt)
		o.ObserveInt64(s.inUseConns, int64(st.InUse), target.attrOpt)
		o.ObserveInt64(s.idleConns, int64(st.Idle), target.attrOpt)
		o.ObserveInt64(s.waitCount, st.WaitCount, target.attrOpt)
		o.ObserveFloat64(s.waitDurationMs, float64(st.WaitDuration)/float64(time.Millisecond), target.attrOpt)
		o.ObserveInt64(s.maxOpenConns, int64(st.MaxOpenConnections), target.attrOpt)
		o.ObserveInt64(s.maxIdleClosed, st.MaxIdleClosed, target.attrOpt)
		o.ObserveInt64(s.maxLifeClosed, st.MaxLifetimeClosed, target.attrOpt)
	}
	return nil
}

func (s *OTelDBStats) initInstruments() error {

	var err error
	s.openConns, err = s.meter.Int64ObservableGauge("db.pool.open_connections")
	if err != nil {
		return err
	}
	s.inUseConns, err = s.meter.Int64ObservableGauge("db.pool.in_use")
	if err != nil {
		return err
	}
	s.idleConns, err = s.meter.Int64ObservableGauge("db.pool.idle")
	if err != nil {
		return err
	}
	s.waitCount, err = s.meter.Int64ObservableCounter("db.pool.wait_count")
	if err != nil {
		return err
	}
	s.waitDurationMs, err = s.meter.Float64ObservableGauge("db.pool.wait_duration_ms")
	if err != nil {
		return err
	}
	s.maxOpenConns, err = s.meter.Int64ObservableGauge("db.pool.max_open_connections")
	if err != nil {
		return err
	}
	s.maxIdleClosed, err = s.meter.Int64ObservableCounter("db.pool.max_idletime_closed")
	if err != nil {
		return err
	}
	s.maxLifeClosed, err = s.meter.Int64ObservableCounter("db.pool.max_lifetime_closed")
	if err != nil {
		return err
	}
	return nil
}
