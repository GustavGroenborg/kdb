package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gcrg/kdb/service/pkg/dar"
	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type LocationsHandler struct {
	DB        *db.DB
	DARClient *dar.Client
}

func (h *LocationsHandler) resolveKommune(ctx context.Context, address string) (string, string, bool) {
	if h.DARClient == nil {
		log.Printf("DARClient is nil; skipping municipality resolution for %q", address)
		return "", "", false
	}
	res, err := h.DARClient.LookupAddress(ctx, address)
	if err != nil {
		log.Printf("DAR address lookup error for %q: %v", address, err)
		return "", "", false
	}
	log.Printf("DAR resolved %q -> Kommune %s (%s), isYder: %v", address, res.KommuneCode, res.KommuneName, res.IsYderkommune)
	return res.KommuneCode, res.KommuneName, res.IsYderkommune
}

type CreateLocationRequest struct {
	Name    string              `json:"name"`
	Type    models.LocationType `json:"type"`
	Address string              `json:"address"`
}

func (h *LocationsHandler) List(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())

	rows, err := h.DB.Query(`
		SELECT id, user_id, name, type, address, is_primary_home, kommune_code, kommune_name, is_yderkommune, created_at
		FROM locations
		WHERE user_id = ?
		ORDER BY CASE WHEN type = 'home' THEN 1 ELSE 2 END, name ASC
	`, targetUserID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch locations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	locations := make([]models.Location, 0)
	for rows.Next() {
		var loc models.Location
		var kCode, kName sql.NullString
		var isYder sql.NullBool
		if err := rows.Scan(
			&loc.ID, &loc.UserID, &loc.Name, &loc.Type, &loc.Address, &loc.IsPrimaryHome,
			&kCode, &kName, &isYder, &loc.CreatedAt,
		); err != nil {
			http.Error(w, `{"error":"failed to parse location data"}`, http.StatusInternalServerError)
			return
		}
		loc.KommuneCode = kCode.String
		loc.KommuneName = kName.String
		loc.IsYderkommune = isYder.Bool
		locations = append(locations, loc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func (h *LocationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)

	if req.Name == "" || req.Address == "" {
		http.Error(w, `{"error":"name and address are required"}`, http.StatusBadRequest)
		return
	}

	if req.Type != models.LocationTypeHome && req.Type != models.LocationTypeWorkplace && req.Type != models.LocationTypeUniversity {
		http.Error(w, `{"error":"type must be 'home', 'workplace', or 'university'"}`, http.StatusBadRequest)
		return
	}

	isHome := req.Type == models.LocationTypeHome
	var kCode, kName string
	var isYder bool
	if isHome {
		kCode, kName, isYder = h.resolveKommune(r.Context(), req.Address)
	}

	// If it's a home, enforce single home rule: remove or update existing home
	if isHome {
		var existingHomeID string
		err := h.DB.QueryRow("SELECT id FROM locations WHERE user_id = ? AND type = 'home'", targetUserID).Scan(&existingHomeID)
		if err == nil && existingHomeID != "" {
			// Update the existing home instead of creating duplicate home
			_, err = h.DB.Exec(
				"UPDATE locations SET name = ?, address = ?, is_primary_home = 1, kommune_code = ?, kommune_name = ?, is_yderkommune = ? WHERE id = ?",
				req.Name, req.Address, kCode, kName, isYder, existingHomeID,
			)
			if err != nil {
				http.Error(w, `{"error":"failed to update existing home"}`, http.StatusInternalServerError)
				return
			}

			var updatedHome models.Location
			var kCodeNull, kNameNull sql.NullString
			var isYderNull sql.NullBool
			_ = h.DB.QueryRow(
				"SELECT id, user_id, name, type, address, is_primary_home, kommune_code, kommune_name, is_yderkommune, created_at FROM locations WHERE id = ?",
				existingHomeID,
			).Scan(
				&updatedHome.ID, &updatedHome.UserID, &updatedHome.Name, &updatedHome.Type, &updatedHome.Address, &updatedHome.IsPrimaryHome,
				&kCodeNull, &kNameNull, &isYderNull, &updatedHome.CreatedAt,
			)
			updatedHome.KommuneCode = kCodeNull.String
			updatedHome.KommuneName = kNameNull.String
			updatedHome.IsYderkommune = isYderNull.Bool

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedHome)
			return
		}
	}

	loc := models.Location{
		ID:            uuid.New().String(),
		UserID:        targetUserID,
		Name:          req.Name,
		Type:          req.Type,
		Address:       req.Address,
		IsPrimaryHome: isHome,
		KommuneCode:   kCode,
		KommuneName:   kName,
		IsYderkommune: isYder,
		CreatedAt:     time.Now().UTC(),
	}

	_, err := h.DB.Exec(
		"INSERT INTO locations (id, user_id, name, type, address, is_primary_home, kommune_code, kommune_name, is_yderkommune, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		loc.ID, loc.UserID, loc.Name, loc.Type, loc.Address, loc.IsPrimaryHome, loc.KommuneCode, loc.KommuneName, loc.IsYderkommune, loc.CreatedAt,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to create location"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loc)
}

func (h *LocationsHandler) Update(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	locID := chi.URLParam(r, "id")

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)

	if req.Name == "" || req.Address == "" {
		http.Error(w, `{"error":"name and address are required"}`, http.StatusBadRequest)
		return
	}

	var locType models.LocationType
	err := h.DB.QueryRow("SELECT type FROM locations WHERE id = ? AND user_id = ?", locID, targetUserID).Scan(&locType)
	if err != nil {
		http.Error(w, `{"error":"location not found"}`, http.StatusNotFound)
		return
	}

	var kCode, kName string
	var isYder bool
	if locType == models.LocationTypeHome {
		kCode, kName, isYder = h.resolveKommune(r.Context(), req.Address)
	}

	res, err := h.DB.Exec(
		"UPDATE locations SET name = ?, address = ?, kommune_code = ?, kommune_name = ?, is_yderkommune = ? WHERE id = ? AND user_id = ?",
		req.Name, req.Address, kCode, kName, isYder, locID, targetUserID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to update location"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"location not found"}`, http.StatusNotFound)
		return
	}

	var updatedLoc models.Location
	var kCodeNull, kNameNull sql.NullString
	var isYderNull sql.NullBool
	_ = h.DB.QueryRow(
		"SELECT id, user_id, name, type, address, is_primary_home, kommune_code, kommune_name, is_yderkommune, created_at FROM locations WHERE id = ?",
		locID,
	).Scan(
		&updatedLoc.ID, &updatedLoc.UserID, &updatedLoc.Name, &updatedLoc.Type, &updatedLoc.Address, &updatedLoc.IsPrimaryHome,
		&kCodeNull, &kNameNull, &isYderNull, &updatedLoc.CreatedAt,
	)
	updatedLoc.KommuneCode = kCodeNull.String
	updatedLoc.KommuneName = kNameNull.String
	updatedLoc.IsYderkommune = isYderNull.Bool

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedLoc)
}


func (h *LocationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	targetUserID := middleware.GetTargetUserID(r.Context())
	locID := chi.URLParam(r, "id")

	// Check if this location is used in any routes
	var routeCount int
	_ = h.DB.QueryRow(
		"SELECT COUNT(*) FROM routes WHERE origin_location_id = ? OR destination_location_id = ?",
		locID, locID,
	).Scan(&routeCount)

	if routeCount > 0 {
		http.Error(w, `{"error":"cannot delete location because it is used in one or more routes"}`, http.StatusBadRequest)
		return
	}

	res, err := h.DB.Exec("DELETE FROM locations WHERE id = ? AND user_id = ?", locID, targetUserID)
	if err != nil {
		http.Error(w, `{"error":"failed to delete location"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"location not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"location deleted successfully"}`))
}

// AutocompleteAddress handles type-ahead search queries for Danish addresses using Adressevælgeren.
func (h *LocationsHandler) AutocompleteAddress(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 || h.DARClient == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	suggestions, err := h.DARClient.Autocomplete(r.Context(), q)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}
