package fradrag

import (
	"github.com/gcrg/kdb/service/pkg/models"
)

// DefaultRates returns the official SKAT befordringsfradrag rates for the given year.
// In 2026, standard rates are 3.17 DKK/km (25-120 km) and 1.59 DKK/km (>120 km),
// while commuters in yderkommuner receive the full rate of 3.51 DKK/km regardless of distance.
func DefaultRates(year int) models.TaxRate {
	if year >= 2026 {
		return models.TaxRate{
			Year:             year,
			ThresholdKm:      24.0,
			BaseRate:         3.17, // 2026 standard DKK/km (25-120 km)
			Tier2ThresholdKm: 120.0,
			Tier2Rate:        1.59, // 2026 standard high-distance rate DKK/km (>120 km)
			YderkommuneRate:  3.51, // 2026 yderkommune rate DKK/km (>24 km)
		}
	}

	// 2024-2025 rates
	return models.TaxRate{
		Year:             year,
		ThresholdKm:      24.0,
		BaseRate:         2.23, // 2024-2025 standard DKK/km (25-120 km)
		Tier2ThresholdKm: 120.0,
		Tier2Rate:        1.12, // 2024-2025 standard high-distance rate DKK/km (>120 km)
		YderkommuneRate:  2.23, // 2024-2025 yderkommune rate DKK/km (>24 km)
	}
}

// Calculate computes the Danish tax deduction (befordringsfradrag) in DKK
// for a given round-trip distance in kilometers.
func Calculate(distanceKm float64, rates models.TaxRate, isYderkommune bool) float64 {
	if distanceKm <= rates.ThresholdKm {
		return 0.0
	}

	if isYderkommune {
		return (distanceKm - rates.ThresholdKm) * rates.YderkommuneRate
	}

	if distanceKm <= rates.Tier2ThresholdKm {
		return (distanceKm - rates.ThresholdKm) * rates.BaseRate
	}

	tier1 := (rates.Tier2ThresholdKm - rates.ThresholdKm) * rates.BaseRate
	tier2 := (distanceKm - rates.Tier2ThresholdKm) * rates.Tier2Rate
	return tier1 + tier2
}
