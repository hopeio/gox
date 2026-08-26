package sql

import "testing"

func TestStringArrayRoundTrip(t *testing.T) {
	cases := []StringArray{
		nil,
		{},
		{"a"},
		{"a", "b"},
		{`a"b`, `c\d`},
		{"image/6cbeb5c8-7160-4b6f-a342-d96d3c00367a.jpg"},
	}
	for _, in := range cases {
		v, err := in.Value()
		if err != nil {
			t.Fatalf("Value(%v): %v", in, err)
		}
		if in == nil {
			if v != nil {
				t.Fatalf("nil array should encode to nil, got %#v", v)
			}
			continue
		}
		s, ok := v.(string)
		if !ok {
			t.Fatalf("Value should return string, got %T", v)
		}
		var out StringArray
		if err := out.Scan(s); err != nil {
			t.Fatalf("Scan(%q): %v", s, err)
		}
		if len(out) != len(in) {
			t.Fatalf("len: got %d want %d (%q)", len(out), len(in), s)
		}
		for i := range in {
			if out[i] != in[i] {
				t.Fatalf("elem %d: got %q want %q (literal %q)", i, out[i], in[i], s)
			}
		}
	}
}

func TestStringArrayScanNativeSlice(t *testing.T) {
	var out StringArray
	if err := out.Scan([]string{"x", "y"}); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != "x" || out[1] != "y" {
		t.Fatalf("got %#v", out)
	}
}
