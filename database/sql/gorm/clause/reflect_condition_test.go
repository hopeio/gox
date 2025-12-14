package clause

import (
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils/tests"
)

type Report struct {
	gorm.Model
}

type ReportList struct {
	PaginationEmbedded
	LoadingTime *Range[time.Time]
	UserId      int `sqlcond:"-"`
	CarId       int `sqlcond:"emptyvalid"`
	TaskId      int
	Or          Condition `sqlcond:"or"`
	Embedded    Condition `sqlcond:"embedded"`
	And         Condition
}

type Condition struct {
	UserId int
	CarId  int `sqlcond:"operate:>"`
	TaskId int `sqlcond:"<"`
	Expr   int `sqlcond:"operate:id = ?;emptyvalid"`
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

	condition := ConditionsBy(&ReportList{
		LoadingTime: &Range[time.Time]{
			Field: "loading_time",
			Begin: time.Now(),
			End:   time.Now(),
		}, UserId: 1, Or: Condition{UserId: 1, CarId: 1, TaskId: 1}, Embedded: Condition{UserId: 1, CarId: 1, TaskId: 1}, And: Condition{
			UserId: 1,
			CarId:  1,
			TaskId: 1,
			Expr:   2,
		}})

	t.Log(db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var reports []*Report
		return tx.Clauses(condition...).Find(&reports)
	}))
}
