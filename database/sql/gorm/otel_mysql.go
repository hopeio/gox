package gorm

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gorm.io/gorm"
)

type MySQLCollector struct {
	Interval      time.Duration
	VariableNames []string

	mu       sync.RWMutex
	values   map[string]float64
	allowSet map[string]struct{}
	gauge    metric.Float64ObservableGauge
	reg      metric.Registration
	stopCh   chan struct{}
	closeOnce sync.Once
}

func NewMySQLCollector(variableNames ...string) *MySQLCollector {
	return &MySQLCollector{Interval: 15 * time.Second, VariableNames: variableNames}
}

func (c *MySQLCollector) Init(db *gorm.DB, meter metric.Meter) error {
	if c.Interval <= 0 {
		c.Interval = 15 * time.Second
	}
	c.values = make(map[string]float64)
	c.allowSet = make(map[string]struct{}, len(c.VariableNames))
	c.stopCh = make(chan struct{})
	for _, name := range c.VariableNames {
		c.allowSet[name] = struct{}{}
	}
	gauge, err := meter.Float64ObservableGauge("gorm.db.mysql.status")
	if err != nil {
		return err
	}
	c.gauge = gauge
	reg, err := meter.RegisterCallback(c.observe, c.gauge)
	if err != nil {
		return err
	}
	c.reg = reg
	go c.run(db)
	return nil
}

func (c *MySQLCollector) run(db *gorm.DB) {
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

func (c *MySQLCollector) Close(context.Context) error {
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

func (c *MySQLCollector) observe(_ context.Context, o metric.Observer) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for name, val := range c.values {
		o.ObserveFloat64(c.gauge, val, metric.WithAttributes(attribute.String("variable", name)))
	}
	return nil
}

func (c *MySQLCollector) collect(db *gorm.DB) {
	rows, err := db.Raw("SHOW STATUS").Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	values := make(map[string]float64)
	for rows.Next() {
		var name, raw string
		if rows.Scan(&name, &raw) != nil || !c.allowed(name) {
			continue
		}
		val, ok := parseStatusNumber(raw)
		if !ok {
			continue
		}
		values[name] = val
	}
	c.mu.Lock()
	c.values = values
	c.mu.Unlock()
}

func (c *MySQLCollector) allowed(name string) bool {
	if len(c.allowSet) == 0 {
		return true
	}
	_, ok := c.allowSet[name]
	return ok
}

func parseStatusNumber(raw string) (float64, bool) {
	if raw == "" || strings.Contains(raw, ":") {
		return 0, false
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}
