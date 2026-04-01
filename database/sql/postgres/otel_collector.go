package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OTelCollector struct {
	meter          metric.Meter
	db             *sql.DB

	lagGauge       metric.Float64ObservableGauge
	startGauge     metric.Float64ObservableGauge
	dbSizeGauge    metric.Float64ObservableGauge
	tableRowsGauge metric.Float64ObservableGauge
	reg            metric.Registration
	closeOnce      sync.Once
}

func NewOTelCollector(db *sql.DB, meter metric.Meter) *OTelCollector {
	return &OTelCollector{db: db, meter: meter}
}

func (c *OTelCollector) Init() error {
	if err := c.initInstruments(); err != nil {
		return err
	}
	reg, err := c.meter.RegisterCallback(c.observe, c.lagGauge, c.startGauge, c.dbSizeGauge, c.tableRowsGauge)
	if err != nil {
		return err
	}
	c.reg = reg
	return nil
}

func (c *OTelCollector) initInstruments() error {
	var err error
	c.lagGauge, err = c.meter.Float64ObservableGauge("db.postgres.replication_lag_seconds")
	if err != nil {
		return err
	}
	c.startGauge, err = c.meter.Float64ObservableGauge("db.postgres.postmaster_start_time_seconds")
	if err != nil {
		return err
	}
	c.dbSizeGauge, err = c.meter.Float64ObservableGauge("db.postgres.database_size_bytes")
	if err != nil {
		return err
	}
	c.tableRowsGauge, err = c.meter.Float64ObservableGauge("db.postgres.table_rows")
	return err
}



func (c *OTelCollector) Close(context.Context) error {
	c.closeOnce.Do(func() {
		if c.reg != nil {
			c.reg.Unregister()
		}
	})
	return nil
}

func (c *OTelCollector) observe(_ context.Context, o metric.Observer) error {
	o.ObserveFloat64(c.lagGauge, c.queryLag())
	o.ObserveFloat64(c.startGauge, c.queryStartTime())
	for datname, size := range c.queryDatabaseSize() {
		o.ObserveFloat64(c.dbSizeGauge, size, metric.WithAttributes(attribute.String("datname", datname)))
	}
	for key, rows := range c.queryTableRows() {
		schema, table := splitTableKey(key)
		o.ObserveFloat64(c.tableRowsGauge, rows, metric.WithAttributes(attribute.String("schema", schema), attribute.String("table", table)))
	}
	return nil
}

func (c *OTelCollector) queryLag() float64 {
	rows, err := c.db.Query("SELECT CASE WHEN NOT pg_is_in_recovery() THEN 0 ELSE GREATEST(0, EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))) END")
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

func (c *OTelCollector) queryStartTime() float64 {
	rows, err := c.db.Query("SELECT EXTRACT(EPOCH FROM pg_postmaster_start_time())")
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

func (c *OTelCollector) queryDatabaseSize() map[string]float64 {
	rows, err := c.db.Query("SELECT datname, pg_database_size(datname) AS size_bytes FROM pg_database")
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

func (c *OTelCollector) queryTableRows() map[string]float64 {
	rows, err := c.db.Query("SELECT schemaname, relname, COALESCE(n_live_tup,0) FROM pg_stat_user_tables")
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
