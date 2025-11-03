package id

import (
	"github.com/google/uuid"
	"testing"
)

func TestId(t *testing.T) {
	t.Log(uuid.New().String())
}
