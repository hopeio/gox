package gorm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gorm.io/gorm"
)

type PostgresCollector struct {
	Interval time.Duration

	mu             sync.RWMutex
	replicationLag float64
	startTime      float64
	databaseSize   map[string]float64
	tableRows      map[string]float64

	lagGauge       metric.Float64ObservableGauge
	startGauge     metric.Float64ObservableGauge
	dbSizeGauge    metric.Float64ObservableGauge
	tableRowsGauge metric.Float64ObservableGauge
	reg            metric.Registration
	stopCh         chan struct{}
	closeOnce      sync.Once
}

func NewPostgresCollector() *PostgresCollector {
	return &PostgresCollector{Interval: 15 * time.Second}
}

func (c *PostgresCollector) Init(db *gorm.DB, meter metric.Meter) error {
	if c.Interval <= 0 {
		c.Interval = 15 * time.Second
	}
	c.databaseSize = make(map[string]float64)
	c.tableRows = make(map[string]float64)
	c.stopCh = make(chan struct{})
	if err := c.initInstruments(meter); err != nil {
		return err
	}
	reg, err := meter.RegisterCallback(c.observe, c.lagGauge, c.startGauge, c.dbSizeGauge, c.tableRowsGauge)
	if err != nil {
		return err
	}
	c.reg = reg
	go c.run(db)
	return nil
}

func (c *PostgresCollector) initInstruments(meter metric.Meter) error {
	var err error
	c.lagGauge, err = meter.Float64ObservableGauge("gorm.db.postgres.replication_lag_seconds")
	if err != nil {
		return err
	}
	c.startGauge, err = meter.Float64ObservableGauge("gorm.db.postgres.postmaster_start_time_seconds")
	if err != nil {
		return err
	}
	c.dbSizeGauge, err = meter.Float64ObservableGauge("gorm.db.postgres.database_size_bytes")
	if err != nil {
		return err
	}
	c.tableRowsGauge, err = meter.Float64ObservableGauge("gorm.db.postgres.table_rows")
	return err
}

func (c *PostgresCollector) run(db *gorm.DB) {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	c.collect(db)
	for {
		select {
		case <-ticker.C:
			c.collect(db)
		case <-c.stopCh:
			return
		}
	}
}

func (c *PostgresCollector) Close(context.Context) error {
	c.closeOnce.Do(func() {
		if c.reg != nil {
			c.reg.Unregister()
		}
		if c.stopCh != nil {
			close(c.stopCh)
		}
	})
	return nil
}

func (c *PostgresCollector) observe(_ context.Context, o metric.Observer) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	o.ObserveFloat64(c.lagGauge, c.replicationLag)
	o.ObserveFloat64(c.startGauge, c.startTime)
	for datname, size := range c.databaseSize {
		o.ObserveFloat64(c.dbSizeGauge, size, metric.WithAttributes(attribute.String("datname", datname)))
	}
	for key, rows := range c.tableRows {
		schema, table := splitTableKey(key)
		o.ObserveFloat64(c.tableRowsGauge, rows, metric.WithAttributes(attribute.String("schema", schema), attribute.String("table", table)))
	}
	return nil
}

func (c *PostgresCollector) collect(db *gorm.DB) {
	lag := c.queryLag(db)
	start := c.queryStartTime(db)
	sizes := c.queryDatabaseSize(db)
	rows := c.queryTableRows(db)
	c.mu.Lock()
	c.replicationLag = lag
	c.startTime = start
	c.databaseSize = sizes
	c.tableRows = rows
	c.mu.Unlock()
}

func (c *PostgresCollector) queryLag(db *gorm.DB) float64 {
	rows, err := db.Raw("SELECT CASE WHEN NOT pg_is_in_recovery() THEN 0 ELSE GREATEST(0, EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))) END").Rows()
	if err != nil {
		return 0
	}
	defer rows.Close()
	var lag float64
	if rows.Next() {
		_ = rows.Scan(&lag)
	}
	return lag
}

func (c *PostgresCollector) queryStartTime(db *gorm.DB) float64 {
	rows, err := db.Raw("SELECT EXTRACT(EPOCH FROM pg_postmaster_start_time())").Rows()
	if err != nil {
		return 0
	}
	defer rows.Close()
	var start float64
	if rows.Next() {
		_ = rows.Scan(&start)
	}
	return start
}

func (c *PostgresCollector) queryDatabaseSize(db *gorm.DB) map[string]float64 {
	rows, err := db.Raw("SELECT datname, pg_database_size(datname) AS size_bytes FROM pg_database").Rows()
	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()
	out := make(map[string]float64)
	for rows.Next() {
		var datname string
		var size float64
		if rows.Scan(&datname, &size) == nil {
			out[datname] = size
		}
	}
	return out
}

func (c *PostgresCollector) queryTableRows(db *gorm.DB) map[string]float64 {
	query := "SELECT schemaname, relname, COALESCE(n_live_tup,0) FROM pg_stat_user_tables"
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()
	out := make(map[string]float64)
	for rows.Next() {
		var schema, table string
		var cnt float64
		if rows.Scan(&schema, &table, &cnt) == nil {
			out[fmt.Sprintf("%s.%s", schema, table)] = cnt
		}
	}
	return out
}

func splitTableKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i], key[i+1:]
		}
	}
	return "", key
}
