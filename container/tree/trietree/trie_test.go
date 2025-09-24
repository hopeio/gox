package trietree

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrie(t *testing.T) {
	node := &Node[int]{}
	node.Set("/static/*filepath", 1)

	node.Set("/", 2)
	node.Set("/apib", 3)
	node.Set("/api", 4)
	node.Set("/abc", 5)
	node.Set("/bcd", 6)

	node.Set("/abc/def", 7)
	node.Set("/test/id/path/path/*path", 8)
	node.Set("/id", 9)
	v, _, _ := node.Get("/static/filepath")
	assert.Equal(t, 1, v)
	v, _, _ = node.Get("/")
	assert.Equal(t, 2, v)
	v, _, _ = node.Get("/apib")
	assert.Equal(t, 3, v)
	v, _, _ = node.Get("/api")
	assert.Equal(t, 4, v)
	v, _, _ = node.Get("/abc")
	assert.Equal(t, 5, v)
	v, _, _ = node.Get("/bcd")
	assert.Equal(t, 6, v)
	v, _, _ = node.Get("/abc/def")
	assert.Equal(t, 7, v)
	v, p, _ := node.Get("/test/id/path/path/path")
	t.Log(p)
	assert.Equal(t, 8, v)
	v, _, _ = node.Get("/id")
	assert.Equal(t, 9, v)
	v, _, _ = node.Get("/id1")
	assert.Equal(t, 0, v)
}
