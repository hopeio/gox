package gorm

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDBOperation(t *testing.T) {
	cases := []struct {
		q, want string
	}{
		{"SELECT * FROM users", "select"},
		{"/* comment */ INSERT INTO t VALUES (1)", "insert"},
		{"-- c\nUPDATE users SET a=1", "update"},
		{"", ""},
		{";  delete from t", "delete"},
	}
	for _, c := range cases {
		require.Equal(t, c.want, dbOperation(c.q), c.q)
	}
}

func TestIsError(t *testing.T) {
	require.False(t, isError(nil))
	require.False(t, isError(gorm.ErrRecordNotFound))
	require.False(t, isError(driver.ErrSkip))
	require.False(t, isError(io.EOF))
	require.False(t, isError(sql.ErrNoRows))
	require.True(t, isError(errors.New("boom")))
}

func TestErrorType(t *testing.T) {
	require.Equal(t, "", errorType(nil))
	require.Equal(t, "record_not_found", errorType(gorm.ErrRecordNotFound))
	require.Equal(t, "", errorType(sql.ErrNoRows))
	require.Equal(t, "unknown", errorType(errors.New("boom")))
}

func TestFormatQuery(t *testing.T) {
	p := NewOTelPlugin(WithQueryFormatter(func(q string) string { return "masked" }))
	require.Equal(t, "masked", p.formatQuery("select 1"))
	p2 := NewOTelPlugin()
	require.Equal(t, "select 1", p2.formatQuery("select 1"))
}
