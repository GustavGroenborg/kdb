package fradrag

import (
	"math"
	"testing"

	"github.com/gcrg/kdb/service/pkg/models"
)

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func TestCalculate_2024_2025(t *testing.T) {
	rates := DefaultRates(2025)

	tests := []struct {
		name          string
		distanceKm    float64
		isYderkommune bool
		expected      float64
	}{
		{"Below threshold (20 km)", 20.0, false, 0.0},
		{"Exact threshold (24 km)", 24.0, false, 0.0},
		{"Standard [25-120 km] (50 km)", 50.0, false, 57.98},
		{"Standard >120 km (150 km)", 150.0, false, 247.68},
		{"Yderkommune <=24 km (20 km)", 20.0, true, 0.0},
		{"Yderkommune (50 km)", 50.0, true, 57.98},
		{"Yderkommune >120 km (150 km)", 150.0, true, 280.98},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := round2(Calculate(tt.distanceKm, rates, tt.isYderkommune))
			if got != tt.expected {
				t.Errorf("Calculate(%v, isYder=%v) = %v; want %v", tt.distanceKm, tt.isYderkommune, got, tt.expected)
			}
		})
	}
}

func TestCalculate_2026(t *testing.T) {
	rates := DefaultRates(2026)

	tests := []struct {
		name          string
		distanceKm    float64
		isYderkommune bool
		expected      float64
	}{
		{"Below threshold (24 km)", 24.0, false, 0.0},
		{"Standard (50 km)", 50.0, false, 82.42},
		{"Standard (150 km)", 150.0, false, 352.02},
		{"Yderkommune (50 km)", 50.0, true, 91.26},
		{"Yderkommune (150 km)", 150.0, true, 442.26},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := round2(Calculate(tt.distanceKm, rates, tt.isYderkommune))
			if got != tt.expected {
				t.Errorf("Calculate(%v, isYder=%v) = %v; want %v", tt.distanceKm, tt.isYderkommune, got, tt.expected)
			}
		})
	}
}

func TestCalculate_2027Extensible(t *testing.T) {
	rates := models.TaxRate{
		Year:             2027,
		ThresholdKm:      24.0,
		BaseRate:         3.25,
		Tier2ThresholdKm: 120.0,
		Tier2Rate:        1.65,
		YderkommuneRate:  3.60,
	}

	if got := round2(Calculate(50.0, rates, false)); got != 84.50 {
		t.Errorf("got %v, want 84.50", got)
	}

	if got := round2(Calculate(50.0, rates, true)); got != 93.60 {
		t.Errorf("got %v, want 93.60", got)
	}
}
