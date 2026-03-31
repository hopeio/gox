package mysql

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OTelCollector struct {
	VariableNames []string
	prefix        string
	meter         metric.Meter
	db            *sql.DB
	allowSet      map[string]struct{}
	gauge         metric.Int64ObservableGauge
	reg           metric.Registration
	closeOnce     sync.Once
}

func NewOTelCollector(prefix string, db *sql.DB, meter metric.Meter, variableNames ...string) *OTelCollector {
	return &OTelCollector{prefix: prefix, db: db, meter: meter, VariableNames: variableNames}
}

func (c *OTelCollector) Init() error {
	c.allowSet = make(map[string]struct{}, len(c.VariableNames))
	for _, name := range c.VariableNames {
		c.allowSet[name] = struct{}{}
	}
	gauge, err := c.meter.Int64ObservableGauge(c.prefix + "db.mysql.status")
	if err != nil {
		return err
	}
	c.gauge = gauge
	reg, err := c.meter.RegisterCallback(c.observe, c.gauge)
	if err != nil {
		return err
	}
	c.reg = reg
	return nil
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
	rows, err := c.db.Query("SHOW STATUS")
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var name, raw string
		if rows.Scan(&name, &raw) != nil || !c.allowed(name) {
			continue
		}
		val, ok := parseStatusNumber(raw)
		if !ok {
			continue
		}
		o.ObserveInt64(c.gauge, val, metric.WithAttributes(attribute.String("variable", name)))
	}
	return nil
}

func (c *OTelCollector) allowed(name string) bool {
	if len(c.allowSet) == 0 {
		return true
	}
	_, ok := c.allowSet[name]
	return ok
}

func parseStatusNumber(raw string) (int64, bool) {
	if raw == "" || strings.Contains(raw, ":") {
		return 0, false
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}
