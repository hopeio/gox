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

func TestTableFromSQL(t *testing.T) {
	sql := `SELECT b.ref_id,a.id,a.name FROM tag a LEFT JOIN content_tag b ON a.id = b.tag_id WHERE b.type = 1`
	if got := tableFromSQL(sql); got != "tag" {
		t.Fatalf("got %q want tag", got)
	}
}

func TestCollectionName(t *testing.T) {
	t.Run("table expr over dest schema dto", func(t *testing.T) {
		// Table("tag a").Find(&ContentTagRel) → Schema is content_tag_rel, FROM is tag.
		tx := &gorm.DB{Statement: &gorm.Statement{
			Table:     "a",
			Schema:    &schema.Schema{Table: "content_tag_rel"},
			TableExpr: &clause.Expr{SQL: "tag a"},
		}}
		if got := collectionName(tx); got != "tag" {
			t.Fatalf("got %q want tag", got)
		}
	})
	t.Run("schema when matches table", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{
			Table:  "users",
			Schema: &schema.Schema{Table: "users"},
		}}
		if got := collectionName(tx); got != "users" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("from sql when alias mismatches schema", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{
			Table:  "a",
			Schema: &schema.Schema{Table: "content_tag_rel"},
		}}
		tx.Statement.SQL.WriteString(`SELECT b.ref_id FROM tag a LEFT JOIN content_tag b ON a.id = b.tag_id`)
		if got := collectionName(tx); got != "tag" {
			t.Fatalf("got %q want tag", got)
		}
	})
	t.Run("plain table", func(t *testing.T) {
		tx := &gorm.DB{Statement: &gorm.Statement{Table: "orders"}}
		if got := collectionName(tx); got != "orders" {
			t.Fatalf("got %q", got)
		}
	})
}
