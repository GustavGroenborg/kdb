package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type RoutesHandler struct {
	DB *db.DB
}

type CreateRouteRequest struct {
	Name                  string  `json:"name"`
	OriginLocationID      string  `json:"origin_location_id"`
	DestinationLocationID string  `json:"destination_location_id"`
	DistanceOneWayKm      float64 `json:"distance_one_way_km"`
	DistanceRoundTripKm   float64 `json:"distance_round_trip_km"`
	IsDefault             bool    `json:"is_default"`
}

func (h *RoutesHandler) List(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())

	rows, err := h.DB.Query(`
		SELECT 
			r.id, r.user_id, r.name, r.origin_location_id, r.destination_location_id,
			loc_orig.name AS origin_name,
			loc_dest.name AS destination_name,
			r.distance_one_way_km, r.distance_round_trip_km, r.is_default, r.created_at
		FROM routes r
		JOIN locations loc_orig ON r.origin_location_id = loc_orig.id
		JOIN locations loc_dest ON r.destination_location_id = loc_dest.id
		WHERE r.user_id = ?
		ORDER BY r.is_default DESC, r.name ASC
	`, targetUserID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch routes"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	routes := make([]models.Route, 0)
	for rows.Next() {
		var route models.Route
		if err := rows.Scan(
			&route.ID, &route.UserID, &route.Name,
			&route.OriginLocationID, &route.DestinationLocationID,
			&route.OriginLocationName, &route.DestinationLocationName,
			&route.DistanceOneWayKm, &route.DistanceRoundTripKm,
			&route.IsDefault, &route.CreatedAt,
		); err != nil {
			http.Error(w, `{"error":"failed to parse route data"}`, http.StatusInternalServerError)
			return
		}
		routes = append(routes, route)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func (h *RoutesHandler) Create(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())

	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.OriginLocationID == "" || req.DestinationLocationID == "" {
		http.Error(w, `{"error":"name, origin, and destination are required"}`, http.StatusBadRequest)
		return
	}

	if req.DistanceRoundTripKm <= 0 {
		if req.DistanceOneWayKm > 0 {
			req.DistanceRoundTripKm = req.DistanceOneWayKm * 2
		} else {
			http.Error(w, `{"error":"distance must be greater than 0"}`, http.StatusBadRequest)
			return
		}
	}
	if req.DistanceOneWayKm <= 0 {
		req.DistanceOneWayKm = req.DistanceRoundTripKm / 2
	}

	// Check if this is the user's first route
	var existingCount int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM routes WHERE user_id = ?", targetUserID).Scan(&existingCount)
	if existingCount == 0 || req.IsDefault {
		req.IsDefault = true
		// Clear previous default if any
		_, _ = h.DB.Exec("UPDATE routes SET is_default = 0 WHERE user_id = ?", targetUserID)
	}

	route := models.Route{
		ID:                    uuid.New().String(),
		UserID:                targetUserID,
		Name:                  req.Name,
		OriginLocationID:      req.OriginLocationID,
		DestinationLocationID: req.DestinationLocationID,
		DistanceOneWayKm:      req.DistanceOneWayKm,
		DistanceRoundTripKm:   req.DistanceRoundTripKm,
		IsDefault:             req.IsDefault,
		CreatedAt:             time.Now().UTC(),
	}

	_, err := h.DB.Exec(`
		INSERT INTO routes (
			id, user_id, name, origin_location_id, destination_location_id,
			distance_one_way_km, distance_round_trip_km, is_default, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, route.ID, route.UserID, route.Name, route.OriginLocationID, route.DestinationLocationID,
		route.DistanceOneWayKm, route.DistanceRoundTripKm, route.IsDefault, route.CreatedAt)

	if err != nil {
		http.Error(w, `{"error":"failed to create route"}`, http.StatusInternalServerError)
		return
	}

	// Fetch location names for response
	_ = h.DB.QueryRow("SELECT name FROM locations WHERE id = ?", route.OriginLocationID).Scan(&route.OriginLocationName)
	_ = h.DB.QueryRow("SELECT name FROM locations WHERE id = ?", route.DestinationLocationID).Scan(&route.DestinationLocationName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(route)
}

func (h *RoutesHandler) Update(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	routeID := chi.URLParam(r, "id")

	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.OriginLocationID == "" || req.DestinationLocationID == "" {
		http.Error(w, `{"error":"name, origin, and destination are required"}`, http.StatusBadRequest)
		return
	}

	if req.DistanceRoundTripKm <= 0 {
		req.DistanceRoundTripKm = req.DistanceOneWayKm * 2
	}
	if req.DistanceOneWayKm <= 0 {
		req.DistanceOneWayKm = req.DistanceRoundTripKm / 2
	}

	if req.IsDefault {
		_, _ = h.DB.Exec("UPDATE routes SET is_default = 0 WHERE user_id = ?", targetUserID)
	}

	res, err := h.DB.Exec(`
		UPDATE routes SET
			name = ?, origin_location_id = ?, destination_location_id = ?,
			distance_one_way_km = ?, distance_round_trip_km = ?, is_default = ?
		WHERE id = ? AND user_id = ?
	`, req.Name, req.OriginLocationID, req.DestinationLocationID,
		req.DistanceOneWayKm, req.DistanceRoundTripKm, req.IsDefault, routeID, targetUserID)

	if err != nil {
		http.Error(w, `{"error":"failed to update route"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"route not found"}`, http.StatusNotFound)
		return
	}

	var route models.Route
	_ = h.DB.QueryRow(`
		SELECT 
			r.id, r.user_id, r.name, r.origin_location_id, r.destination_location_id,
			loc_orig.name AS origin_name,
			loc_dest.name AS destination_name,
			r.distance_one_way_km, r.distance_round_trip_km, r.is_default, r.created_at
		FROM routes r
		JOIN locations loc_orig ON r.origin_location_id = loc_orig.id
		JOIN locations loc_dest ON r.destination_location_id = loc_dest.id
		WHERE r.id = ?
	`, routeID).Scan(
		&route.ID, &route.UserID, &route.Name,
		&route.OriginLocationID, &route.DestinationLocationID,
		&route.OriginLocationName, &route.DestinationLocationName,
		&route.DistanceOneWayKm, &route.DistanceRoundTripKm,
		&route.IsDefault, &route.CreatedAt,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(route)
}

func (h *RoutesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	routeID := chi.URLParam(r, "id")

	res, err := h.DB.Exec("DELETE FROM routes WHERE id = ? AND user_id = ?", routeID, targetUserID)
	if err != nil {
		http.Error(w, `{"error":"failed to delete route"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"route not found"}`, http.StatusNotFound)
		return
	}

	// If there are other routes but none is default, set the latest as default
	var defaultCount int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM routes WHERE user_id = ? AND is_default = 1", targetUserID).Scan(&defaultCount)
	if defaultCount == 0 {
		_, _ = h.DB.Exec("UPDATE routes SET is_default = 1 WHERE id = (SELECT id FROM routes WHERE user_id = ? ORDER BY created_at DESC LIMIT 1)", targetUserID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"route deleted successfully"}`))
}
