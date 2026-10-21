package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/fradrag"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type DrivesHandler struct {
	DB *db.DB
}

type CreateDriveRequest struct {
	RouteID    string  `json:"route_id"`
	Date       string  `json:"date"` // YYYY-MM-DD
	DistanceKm float64 `json:"distance_km"` // optional override
	Notes      string  `json:"notes"`
}

func (h *DrivesHandler) List(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())

	yearParam := r.URL.Query().Get("year")
	monthParam := r.URL.Query().Get("month")
	dateParam := r.URL.Query().Get("date")
	routeIDParam := r.URL.Query().Get("route_id")
	limitParam := r.URL.Query().Get("limit")

	query := `
		SELECT 
			d.id, d.user_id, d.registered_by_user_id, u.username AS registered_by_name,
			d.route_id, r.name AS route_name,
			d.date, d.distance_km, d.calculated_fradrag, d.notes, d.created_at
		FROM drives d
		JOIN routes r ON d.route_id = r.id
		JOIN users u ON d.registered_by_user_id = u.id
		WHERE d.user_id = ?
	`
	args := []interface{}{targetUserID}

	if dateParam != "" {
		query += " AND d.date = ?"
		args = append(args, dateParam)
	}

	if yearParam != "" {
		query += " AND substr(d.date, 1, 4) = ?"
		args = append(args, yearParam)
	}

	if monthParam != "" {
		if len(monthParam) == 7 { // "2026-10"
			query += " AND substr(d.date, 1, 7) = ?"
			args = append(args, monthParam)
		} else if len(monthParam) <= 2 { // "10"
			query += " AND substr(d.date, 6, 2) = ?"
			if len(monthParam) == 1 {
				monthParam = "0" + monthParam
			}
			args = append(args, monthParam)
		}
	}

	if routeIDParam != "" {
		query += " AND d.route_id = ?"
		args = append(args, routeIDParam)
	}

	query += " ORDER BY d.date DESC, d.created_at DESC"

	if limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			query += " LIMIT ?"
			args = append(args, limit)
		}
	} else {
		query += " LIMIT 100"
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch drives"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	drives := make([]models.Drive, 0)
	for rows.Next() {
		var d models.Drive
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.RegisteredByUserID, &d.RegisteredByName,
			&d.RouteID, &d.RouteName,
			&d.Date, &d.DistanceKm, &d.CalculatedFradrag, &d.Notes, &d.CreatedAt,
		); err != nil {
			http.Error(w, `{"error":"failed to parse drive data"}`, http.StatusInternalServerError)
			return
		}
		drives = append(drives, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drives)
}

func (h *DrivesHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUser, _ := middleware.GetAuthUser(r.Context())
	targetUserID := middleware.GetTargetUserID(r.Context())

	var req CreateDriveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.RouteID == "" {
		http.Error(w, `{"error":"route_id is required"}`, http.StatusBadRequest)
		return
	}

	// Fetch route details
	var routeName string
	var routeRoundTripKm float64
	err := h.DB.QueryRow(
		"SELECT name, distance_round_trip_km FROM routes WHERE id = ? AND user_id = ?",
		req.RouteID, targetUserID,
	).Scan(&routeName, &routeRoundTripKm)

	if err != nil {
		http.Error(w, `{"error":"route not found"}`, http.StatusNotFound)
		return
	}

	// Determine distance (prompt: "A drive counts for both back and forth.")
	distanceKm := req.DistanceKm
	if distanceKm <= 0 {
		distanceKm = routeRoundTripKm
	}

	// Default date to today
	driveDate := strings.TrimSpace(req.Date)
	if driveDate == "" {
		driveDate = time.Now().Format("2006-01-02")
	}

	// Parse year to fetch tax rates
	driveYear := time.Now().Year()
	if len(driveDate) >= 4 {
		if y, err := strconv.Atoi(driveDate[:4]); err == nil {
			driveYear = y
		}
	}

	// Fetch tax rates for this year
	var rates models.TaxRate
	err = h.DB.QueryRow(
		"SELECT year, threshold_km, base_rate, tier2_threshold_km, tier2_rate FROM tax_rates WHERE year = ?",
		driveYear,
	).Scan(&rates.Year, &rates.ThresholdKm, &rates.BaseRate, &rates.Tier2ThresholdKm, &rates.Tier2Rate)

	if err == sql.ErrNoRows {
		rates = fradrag.DefaultRates(driveYear)
	}

	calculatedFradrag := fradrag.Calculate(distanceKm, rates)

	drive := models.Drive{
		ID:                 uuid.New().String(),
		UserID:             targetUserID,
		RegisteredByUserID: authUser.ID,
		RegisteredByName:   authUser.Username,
		RouteID:            req.RouteID,
		RouteName:          routeName,
		Date:               driveDate,
		DistanceKm:         distanceKm,
		CalculatedFradrag:  calculatedFradrag,
		Notes:              strings.TrimSpace(req.Notes),
		CreatedAt:          time.Now().UTC(),
	}

	_, err = h.DB.Exec(`
		INSERT INTO drives (
			id, user_id, registered_by_user_id, route_id, date, distance_km, calculated_fradrag, notes, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, drive.ID, drive.UserID, drive.RegisteredByUserID, drive.RouteID,
		drive.Date, drive.DistanceKm, drive.CalculatedFradrag, drive.Notes, drive.CreatedAt)

	if err != nil {
		http.Error(w, `{"error":"failed to record drive"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(drive)
}

func (h *DrivesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	driveID := chi.URLParam(r, "id")

	res, err := h.DB.Exec("DELETE FROM drives WHERE id = ? AND user_id = ?", driveID, targetUserID)
	if err != nil {
		http.Error(w, `{"error":"failed to delete drive"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"drive not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"drive deleted successfully"}`))
}
