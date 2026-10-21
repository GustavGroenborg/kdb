package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/fradrag"
	"github.com/gcrg/kdb/service/pkg/models"
)

type RatesHandler struct {
	DB *db.DB
}

func (h *RatesHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	var rates models.TaxRate
	err := h.DB.QueryRow(
		"SELECT year, threshold_km, base_rate, tier2_threshold_km, tier2_rate FROM tax_rates WHERE year = ?",
		year,
	).Scan(&rates.Year, &rates.ThresholdKm, &rates.BaseRate, &rates.Tier2ThresholdKm, &rates.Tier2Rate)

	if err != nil {
		rates = fradrag.DefaultRates(year)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

func (h *RatesHandler) UpdateRates(w http.ResponseWriter, r *http.Request) {
	var req models.TaxRate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Year <= 0 {
		req.Year = time.Now().Year()
	}
	if req.ThresholdKm <= 0 {
		req.ThresholdKm = 24.0
	}
	if req.BaseRate <= 0 {
		req.BaseRate = 2.23
	}
	if req.Tier2ThresholdKm <= 0 {
		req.Tier2ThresholdKm = 120.0
	}
	if req.Tier2Rate <= 0 {
		req.Tier2Rate = 1.12
	}

	_, err := h.DB.Exec(`
		INSERT INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(year) DO UPDATE SET
			threshold_km = excluded.threshold_km,
			base_rate = excluded.base_rate,
			tier2_threshold_km = excluded.tier2_threshold_km,
			tier2_rate = excluded.tier2_rate
	`, req.Year, req.ThresholdKm, req.BaseRate, req.Tier2ThresholdKm, req.Tier2Rate)

	if err != nil {
		http.Error(w, `{"error":"failed to update tax rates"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}
