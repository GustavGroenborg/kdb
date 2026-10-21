package fradrag

import (
	"math"

	"github.com/gcrg/kdb/service/pkg/models"
)

// DefaultRates returns the standard SKAT befordringsfradrag rates.
func DefaultRates(year int) models.TaxRate {
	return models.TaxRate{
		Year:             year,
		ThresholdKm:      24.0,
		BaseRate:         2.23, // 2024-2026 standard DKK/km
		Tier2ThresholdKm: 120.0,
		Tier2Rate:        1.12, // standard high-distance rate DKK/km
	}
}

// Calculate computes the Danish tax deduction (befordringsfradrag) in DKK
// for a given round-trip distance in kilometers.
func Calculate(distanceKm float64, rates models.TaxRate) float64 {
	if distanceKm <= rates.ThresholdKm || distanceKm <= 0 {
		return 0.0
	}

	threshold := rates.ThresholdKm
	tier2Thresh := rates.Tier2ThresholdKm
	baseRate := rates.BaseRate
	tier2Rate := rates.Tier2Rate

	// Fallback to standard rates if rates are zero
	if baseRate <= 0 {
		baseRate = 2.23
	}
	if tier2Thresh <= 0 {
		tier2Thresh = 120.0
	}
	if tier2Rate <= 0 {
		tier2Rate = 1.12
	}

	var fradrag float64
	if distanceKm <= tier2Thresh {
		// Between 24 km and 120 km: (km - 24) * baseRate
		fradrag = (distanceKm - threshold) * baseRate
	} else {
		// Above 120 km: (120 - 24) * baseRate + (km - 120) * tier2Rate
		tier1Deduction := (tier2Thresh - threshold) * baseRate
		tier2Deduction := (distanceKm - tier2Thresh) * tier2Rate
		fradrag = tier1Deduction + tier2Deduction
	}

	// Round to 2 decimal places
	return math.Round(fradrag*100) / 100
}
