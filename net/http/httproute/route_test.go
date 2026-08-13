package httproute

import (
	"net/http/httptest"
	"testing"
)

func TestCollapseKeepsCardinalityLow(t *testing.T) {
	cases := map[string]string{
		"/api/user/12345": "/api/user/:v",
		"/api/exists/098f6bcd4621d373cade4e832627b4f6/12": "/api/exists/:v/:v",
		"/api/user/login":                "/api/user/login",
		"/upload/image/2026/08/12/a.jpg": "/upload/image/:v/:v/:v/a.jpg",
		"":                               "/",
	}
	for in, want := range cases {
		if got := Collapse(in); got != want {
			t.Errorf("Collapse(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLooksVariable(t *testing.T) {
	if !looksVariable("42") {
		t.Error("numbers are ids")
	}
	if !looksVariable("098f6bcd4621d373cade4e832627b4f6") {
		t.Error("md5 is an id")
	}
	if looksVariable("login") || looksVariable("abc") || looksVariable("") {
		t.Error("route words must be kept")
	}
}

func TestOfPrefersServeMuxPattern(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/upload/098f6bcd4621d373cade4e832627b4f6", nil)
	r.Pattern = "POST /api/upload/{md5}"
	if got := Of(r); got != "/api/upload/{md5}" {
		t.Fatalf("Of = %q", got)
	}
	plain := httptest.NewRequest("GET", "/api/user/7", nil)
	if got := Of(plain); got != "/api/user/:v" {
		t.Fatalf("Of fallback = %q", got)
	}
}
