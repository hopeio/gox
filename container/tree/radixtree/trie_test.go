package radixtree

import (
	"fmt"
	"testing"

	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/stretchr/testify/assert"
)

func TestTrie(t *testing.T) {
	paths := []string{
		"/static/*filepath",
		"/",
		"/apib",
		"/api",
		"/abc",
		"/bcd",
		"/abc/def",
		"/test/id/path/path/*path",
		"/id",
	}
	value := []string{
		"/static/*filepath",
		"/",
		"/apib",
		"/api",
		"/abc",
		"/bcd",
		"/abc/def",
		"/test/id/path/path/*path",
		"/id",
	}
	node := &Node[string]{}
	for i, path := range paths {
		node.Set(path, value[i])
	}

	data, err := jsonx.Marshal(node)
	if err != nil {
		t.Log(err)
	}
	fmt.Println(string(data))
	fmt.Printf("%#v\n", node)
	for i, path := range paths {
		v, params, _ := node.Get(path)
		t.Logf("path: %s, value: %s, params: %v", path, v, params)
		assert.Equal(t, v, value[i])
	}
}
