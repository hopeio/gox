package datatypes

import (
	"testing"
)

func TestArray(t *testing.T) {
	data := `{{{{1,2},{3,4}},{{5,6},{7,8}}}}`
	arr := Array[Array[Array[Array[int]]]]{}
	err := arr.Scan(data)
	if err != nil {
		t.Error(err)
	}
	t.Log(arr)
	t.Log(arr.Value())
}
