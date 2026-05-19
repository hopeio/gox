package math

import (
	"math"
	"testing"
)

func TestMax(t *testing.T) {
	if Max(1, 3, 2) != 3 {
		t.Errorf("Max(1, 3, 2) = %v, want 3", Max(1, 3, 2))
	}
	if Max(-1, -3) != -1 {
		t.Errorf("Max(-1, -3) = %v, want -1", Max(-1, -3))
	}
	if Max(5.0, 3.0, 8.0, 2.0) != 8.0 {
		t.Errorf("Max(5, 3, 8, 2) = %v, want 8", Max(5.0, 3.0, 8.0, 2.0))
	}
}

func TestMin(t *testing.T) {
	if Min(1, 3, 2) != 1 {
		t.Errorf("Min(1, 3, 2) = %v, want 1", Min(1, 3, 2))
	}
	if Min(-1, -3) != -3 {
		t.Errorf("Min(-1, -3) = %v, want -3", Min(-1, -3))
	}
}

func TestMinAndMax(t *testing.T) {
	min, max := MinAndMax(1, 5, 3, 2, 4)
	if min != 1 || max != 5 {
		t.Errorf("MinAndMax(1,5,3,2,4) = (%v, %v), want (1, 5)", min, max)
	}
}

func TestStandardDeviation(t *testing.T) {
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	sd := StandardDeviation(data, false)
	if sd <= 0 {
		t.Errorf("StandardDeviation should be positive, got %v", sd)
	}
}

func TestVariance(t *testing.T) {
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	v := Variance(data, false)
	if v <= 0 {
		t.Errorf("Variance should be positive, got %v", v)
	}
	// Empty slice
	v2 := Variance([]float64{}, false)
	if v2 != 0 {
		t.Errorf("Variance of empty slice = %v, want 0", v2)
	}
}

func TestDecimalPlaces_Extra(t *testing.T) {
	result := DecimalPlaces(3.14159, 2)
	if result != 3.14 {
		t.Errorf("DecimalPlaces(3.14159, 2) = %v, want 3.14", result)
	}
	result = DecimalPlaces(3.14159, 0)
	if result != 3.0 {
		t.Errorf("DecimalPlaces(3.14159, 0) = %v, want 3.0", result)
	}
}

func TestDecimalPlacesRound_Extra(t *testing.T) {
	result := DecimalPlacesRound(3.145, 2)
	if result != 3.15 {
		t.Errorf("DecimalPlacesRound(3.145, 2) = %v, want 3.15", result)
	}
	result = DecimalPlacesRound(3.144, 2)
	if result != 3.14 {
		t.Errorf("DecimalPlacesRound(3.144, 2) = %v, want 3.14", result)
	}
}

func TestFormatFloat_Extra(t *testing.T) {
	result := FormatFloat(3.14)
	if result != "3.14" {
		t.Errorf("FormatFloat(3.14) = %q, want %q", result, "3.14")
	}
	result = FormatFloat(math.Pi)
	if len(result) == 0 {
		t.Error("FormatFloat(Pi) should not be empty")
	}
}
