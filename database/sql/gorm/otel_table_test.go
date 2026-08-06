package gorm

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

func TestBaseTableName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"users", "users"},
		{"users AS u", "users"},
		{"users as u", "users"},
		{"users u", "users"},
		{"`users` AS u", "users"},
		{`"users" AS u`, "users"},
		{"public.users AS u", "users"},
		{"public.`users` AS u", "users"},
		{"u", "u"},
		{"", ""},
	}
	for _, c := range cases {
		if got := baseTableName(c.in); got != c.want {
			t.Fatalf("baseTableName(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestCollectionName(t *testing.T) {
	t.Run("schema wins over alias", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{
			Table:  "u",
			Schema: &schema.Schema{Table: "users"},
			TableExpr: &clause.Expr{SQL: "users AS u"},
		}}
		if got := collectionName(tx); got != "users" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("table expr when no schema", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{
			Table:     "u",
			TableExpr: &clause.Expr{SQL: "users AS u"},
		}}
		if got := collectionName(tx); got != "users" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("plain table", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{Table: "orders"}}
		if got := collectionName(tx); got != "orders" {
			t.Fatalf("got %q", got)
		}
	})
}
