package gerber

import (
	"os"
	"testing"
)

func TestGerber(t *testing.T) {
	path := `D:\Gerber_TopLayer.GTL`
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("skip: local gerber fixture missing: %v", err)
	}
	defer f.Close()

	p := LogProcessor{}
	if err := NewParser(p).Parse(f); err != nil {
		t.Error(err)
	}
}
