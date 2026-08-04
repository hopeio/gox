package structtag

import "testing"

func TestParse_OptionsNilWhenAbsent(t *testing.T) {
	tags, err := Parse(`json:"foo"`)
	if err != nil {
		t.Fatal(err)
	}
	tag, ok := tags.Get("json")
	if !ok {
		t.Fatal("missing json")
	}
	if tag.Options != nil {
		t.Fatalf("Options want nil, got %#v", tag.Options)
	}
	if tag.String() != `json:"foo"` {
		t.Fatalf("String=%q", tag.String())
	}
}

func TestMustParse_OKAndPanic(t *testing.T) {
	tags := MustParse(`json:"foo"`)
	if _, ok := tags.Get("json"); !ok {
		t.Fatal("missing")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustParse invalid tag should panic")
		}
	}()
	_ = MustParse("json")
}
