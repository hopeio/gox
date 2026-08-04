package svg

import (
	"github.com/hopeio/gox/encoding/gerber"
	"os"
	"testing"
)

func TestSvg(t *testing.T) {
	path := `D:\Gerber_TopLayer.GTL`
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("skip: local gerber fixture missing: %v", err)
	}
	defer f.Close()

	p := NewProcessor()
	if err := gerber.NewParser(p).Parse(f); err != nil {
		t.Error(err)
	}
	svgPath := `D:\Gerber_TopLayer.svg`
	svg, err := os.Create(svgPath)
	if err != nil {
		t.Skipf("skip: cannot create svg output: %v", err)
	}
	defer svg.Close()
	if err := p.Write(svg); err != nil {
		t.Error(err)
	}
}
