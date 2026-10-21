package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type LocationType string

const (
	LocationTypeHome       LocationType = "home"
	LocationTypeWorkplace  LocationType = "workplace"
	LocationTypeUniversity LocationType = "university"
)

type Location struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id"`
	Name          string       `json:"name"`
	Type          LocationType `json:"type"`
	Address       string       `json:"address"`
	IsPrimaryHome bool         `json:"is_primary_home"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Route struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Name                  string    `json:"name"`
	OriginLocationID      string    `json:"origin_location_id"`
	DestinationLocationID string    `json:"destination_location_id"`
	OriginLocationName    string    `json:"origin_location_name,omitempty"`
	DestinationLocationName string  `json:"destination_location_name,omitempty"`
	DistanceOneWayKm      float64   `json:"distance_one_way_km"`
	DistanceRoundTripKm   float64   `json:"distance_round_trip_km"`
	IsDefault             bool      `json:"is_default"`
	CreatedAt             time.Time `json:"created_at"`
}

type Drive struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`               // Target account
	RegisteredByUserID string    `json:"registered_by_user_id"` // User who recorded it
	RegisteredByName   string    `json:"registered_by_name,omitempty"`
	RouteID            string    `json:"route_id"`
	RouteName          string    `json:"route_name,omitempty"`
	Date               string    `json:"date"` // YYYY-MM-DD
	DistanceKm         float64   `json:"distance_km"`
	CalculatedFradrag  float64   `json:"calculated_fradrag"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

type SharePermission string

const (
	SharePermissionRead  SharePermission = "read"
	SharePermissionWrite SharePermission = "write"
)

type AccountShare struct {
	ID                 string          `json:"id"`
	OwnerUserID        string          `json:"owner_user_id"`
	OwnerUsername      string          `json:"owner_username,omitempty"`
	SharedWithUserID   string          `json:"shared_with_user_id"`
	SharedWithUsername string          `json:"shared_with_username,omitempty"`
	Permission         SharePermission `json:"permission"`
	CreatedAt          time.Time       `json:"created_at"`
}

type TaxRate struct {
	Year             int     `json:"year"`
	ThresholdKm      float64 `json:"threshold_km"`       // default: 24 km
	BaseRate         float64 `json:"base_rate"`          // default: 2.23 DKK/km (25-120 km)
	Tier2ThresholdKm float64 `json:"tier2_threshold_km"` // default: 120 km
	Tier2Rate        float64 `json:"tier2_rate"`         // default: 1.12 DKK/km (> 120 km)
}

// Stats & Aggregation models
type OverviewStats struct {
	TotalDrives     int                 `json:"total_drives"`
	TotalDistanceKm float64             `json:"total_distance_km"`
	TotalFradrag    float64             `json:"total_fradrag"`
	AverageKm       float64             `json:"average_km"`
	MonthlyBreakdown []MonthlyStat      `json:"monthly_breakdown"`
	RouteBreakdown   []RouteStat        `json:"route_breakdown"`
}

type MonthlyStat struct {
	MonthName    string  `json:"month_name"` // "2026-10"
	DrivesCount  int     `json:"drives_count"`
	DistanceKm   float64 `json:"distance_km"`
	TotalFradrag float64 `json:"total_fradrag"`
}

type RouteStat struct {
	RouteID         string  `json:"route_id"`
	RouteName       string  `json:"route_name"`
	DestinationType string  `json:"destination_type"` // workplace or university
	DrivesCount     int     `json:"drives_count"`
	DistanceKm      float64 `json:"distance_km"`
	TotalFradrag    float64 `json:"total_fradrag"`
}
