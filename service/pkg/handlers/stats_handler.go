package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type StatsHandler struct {
	DB *db.DB
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	yearParam := r.URL.Query().Get("year")
	if yearParam == "" {
		yearParam = time.Now().Format("2006")
	}

	var stats models.OverviewStats

	// 1. Overall totals for the specified year
	err := h.DB.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(distance_km), 0),
			COALESCE(SUM(calculated_fradrag), 0),
			COALESCE(AVG(distance_km), 0)
		FROM drives
		WHERE user_id = ? AND substr(date, 1, 4) = ?
	`, targetUserID, yearParam).Scan(
		&stats.TotalDrives,
		&stats.TotalDistanceKm,
		&stats.TotalFradrag,
		&stats.AverageKm,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to compute overall statistics"}`, http.StatusInternalServerError)
		return
	}

	// 2. Monthly breakdown
	monthlyRows, err := h.DB.Query(`
		SELECT 
			substr(date, 1, 7) AS month_name,
			COUNT(*),
			COALESCE(SUM(distance_km), 0),
			COALESCE(SUM(calculated_fradrag), 0)
		FROM drives
		WHERE user_id = ? AND substr(date, 1, 4) = ?
		GROUP BY substr(date, 1, 7)
		ORDER BY month_name ASC
	`, targetUserID, yearParam)

	stats.MonthlyBreakdown = make([]models.MonthlyStat, 0)
	if err == nil {
		defer monthlyRows.Close()
		for monthlyRows.Next() {
			var m models.MonthlyStat
			if err := monthlyRows.Scan(&m.MonthName, &m.DrivesCount, &m.DistanceKm, &m.TotalFradrag); err == nil {
				stats.MonthlyBreakdown = append(stats.MonthlyBreakdown, m)
			}
		}
	}

	// 3. Route breakdown
	routeRows, err := h.DB.Query(`
		SELECT 
			d.route_id,
			r.name,
			loc_dest.type,
			COUNT(d.id),
			COALESCE(SUM(d.distance_km), 0),
			COALESCE(SUM(d.calculated_fradrag), 0)
		FROM drives d
		JOIN routes r ON d.route_id = r.id
		JOIN locations loc_dest ON r.destination_location_id = loc_dest.id
		WHERE d.user_id = ? AND substr(d.date, 1, 4) = ?
		GROUP BY d.route_id, r.name, loc_dest.type
		ORDER BY SUM(d.calculated_fradrag) DESC
	`, targetUserID, yearParam)

	stats.RouteBreakdown = make([]models.RouteStat, 0)
	if err == nil {
		defer routeRows.Close()
		for routeRows.Next() {
			var rs models.RouteStat
			if err := routeRows.Scan(&rs.RouteID, &rs.RouteName, &rs.DestinationType, &rs.DrivesCount, &rs.DistanceKm, &rs.TotalFradrag); err == nil {
				stats.RouteBreakdown = append(stats.RouteBreakdown, rs)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *StatsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	yearParam := r.URL.Query().Get("year")

	query := `
		SELECT 
			d.date,
			r.name,
			loc_orig.name,
			loc_dest.name,
			loc_dest.type,
			d.distance_km,
			d.calculated_fradrag,
			d.notes,
			u.username
		FROM drives d
		JOIN routes r ON d.route_id = r.id
		JOIN locations loc_orig ON r.origin_location_id = loc_orig.id
		JOIN locations loc_dest ON r.destination_location_id = loc_dest.id
		JOIN users u ON d.registered_by_user_id = u.id
		WHERE d.user_id = ?
	`
	args := []interface{}{targetUserID}
	if yearParam != "" {
		query += " AND substr(d.date, 1, 4) = ?"
		args = append(args, yearParam)
	}
	query += " ORDER BY d.date ASC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch drive records for export"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fileName := "kdb-befordringsfradrag"
	if yearParam != "" {
		fileName += "-" + yearParam
	}
	fileName += ".csv"

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write CSV header
	_ = csvWriter.Write([]string{
		"Date",
		"Route",
		"Origin",
		"Destination",
		"Destination Type",
		"Round Trip (km)",
		"Fradrag (DKK)",
		"Notes",
		"Registered By",
	})

	for rows.Next() {
		var (
			date, routeName, originName, destName, destType, notes, registeredBy string
			distanceKm, fradrag                                                  float64
		)
		if err := rows.Scan(
			&date, &routeName, &originName, &destName, &destType,
			&distanceKm, &fradrag, &notes, &registeredBy,
		); err == nil {
			_ = csvWriter.Write([]string{
				date,
				routeName,
				originName,
				destName,
				destType,
				fmt.Sprintf("%.1f", distanceKm),
				fmt.Sprintf("%.2f", fradrag),
				notes,
				registeredBy,
			})
		}
	}
}
