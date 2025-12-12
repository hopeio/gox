package clause

import (
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils/tests"
)

type ReportList struct {
	PaginationEmbedded `sqlcondi:"-"`
	LoadingTime        *Range[time.Time]
	UserId             int
	CarId              int
	TaskId             int
	RouteId            int
	Diff               float64 `sqlcondi:"-"`
	Outlier            bool    `sqlcondi:"-"`
}

func TestConditionsBy(t *testing.T) {
	var db, err = gorm.Open(tests.DummyDialector{}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Statement.SQL = strings.Builder{}
	t.Log(db.Statement.SQL)
	condition := AndConditionBy(&ReportList{UserId: 1})

	t.Log(condition)
	condition.Build(db.Statement)
	t.Log(db.Statement.SQL.String(), err)
}
