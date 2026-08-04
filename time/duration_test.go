/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"context"
	"encoding/json"
	"testing"
	stdtime "time"
)

func TestDurationUnmarshalText(t *testing.T) {
	var d Duration
	if err := d.UnmarshalText([]byte("1h30m")); err != nil {
		t.Fatal(err)
	}
	if stdtime.Duration(d) != 90*stdtime.Minute {
		t.Fatalf("got %v", stdtime.Duration(d))
	}
	if err := d.UnmarshalText([]byte("bad")); err == nil {
		t.Fatal("want error")
	}
}

func TestDurationMarshalText(t *testing.T) {
	d := Duration(500 * stdtime.Millisecond)
	b, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "500ms" {
		t.Fatalf("got %q", b)
	}
}

func TestDurationJSONUnmarshal(t *testing.T) {
	var decoded Duration
	// UnmarshalJSON passes raw bytes to ParseDuration without stripping JSON quotes.
	if err := decoded.UnmarshalJSON([]byte(`500ms`)); err != nil {
		t.Fatal(err)
	}
	if stdtime.Duration(decoded) != 500*stdtime.Millisecond {
		t.Fatalf("got %v", decoded)
	}
	if err := json.Unmarshal([]byte(`null`), &decoded); err != nil {
		t.Fatal(err)
	}
}

func TestDurationMarshalTextJSONShape(t *testing.T) {
	// MarshalJSON returns raw duration string bytes (no extra quotes).
	d := Duration(2 * stdtime.Second)
	b, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "2s" {
		t.Fatalf("MarshalJSON = %q", b)
	}
}

func TestDurationShrink(t *testing.T) {
	d := Duration(10 * stdtime.Second)
	parent, cancel := context.WithTimeout(context.Background(), 50*stdtime.Millisecond)
	defer cancel()
	shrunk, ctx, shrinkCancel := d.Shrink(parent)
	defer shrinkCancel()
	if stdtime.Duration(shrunk) >= 10*stdtime.Second {
		t.Fatalf("shrink should reduce duration, got %v", shrunk)
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline on shrunk context")
	}
	if stdtime.Until(deadline) > 50*stdtime.Millisecond {
		t.Fatal("deadline should respect parent")
	}
}

func TestDurationShrinkNoParentDeadline(t *testing.T) {
	d := Duration(20 * stdtime.Millisecond)
	_, ctx, cancel := d.Shrink(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	if stdtime.Until(deadline) > 20*stdtime.Millisecond {
		t.Fatal("deadline too far")
	}
}

func TestNormalizeDuration(t *testing.T) {
	std := stdtime.Second
	if got := NormalizeDuration(0, std); got != 0 {
		t.Fatalf("zero stays zero, got %v", got)
	}
	if got := NormalizeDuration(5, std); got != 5*std {
		t.Fatalf("bare multiplier, got %v", got)
	}
	if got := NormalizeDuration(2*std, std); got != 2*std {
		t.Fatalf("large value unchanged, got %v", got)
	}
}
