package fradrag

import (
	"math"
	"testing"
)

func TestCalculate(t *testing.T) {
	rates := DefaultRates(2026)

	tests := []struct {
		name       string
		distanceKm float64
		expected   float64
	}{
		{"Zero km", 0, 0.0},
		{"Negative km", -10, 0.0},
		{"Threshold exact 24 km", 24.0, 0.0},
		{"Below threshold 15 km", 15.0, 0.0},
		{"Just above threshold 25 km", 25.0, 2.23},
		{"Commute 50 km", 50.0, 57.98},
		{"Commute 60 km", 60.0, 80.28},
		{"Tier 1 limit 120 km", 120.0, 214.08},
		{"Above tier 1: 140 km", 140.0, 236.48},
		{"Long commute: 200 km", 200.0, 303.68},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.distanceKm, rates)
			if math.Abs(got-tt.expected) > 0.01 {
				t.Errorf("Calculate(%f) = %f; expected %f", tt.distanceKm, got, tt.expected)
			}
		})
	}
}
