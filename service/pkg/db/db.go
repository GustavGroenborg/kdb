package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// InitDB initializes SQLite database connection, enables WAL mode, creates tables, and seeds initial data.
func InitDB(dataSourceName string) (*DB, error) {
	// If dataSourceName is not in-memory, ensure the directory exists
	if dataSourceName != ":memory:" && dataSourceName != "file::memory:?cache=shared" {
		dir := filepath.Dir(dataSourceName)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return nil, fmt.Errorf("failed to create database directory: %w", err)
			}
		}
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1) // recommended for SQLite to avoid lock contention

	// Execute initial pragmas
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			// Some in-memory SQLite modes might not support WAL, ignore error if so
			if dataSourceName != ":memory:" {
				return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
			}
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS locations (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		address TEXT NOT NULL,
		is_primary_home BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_locations_user_id ON locations(user_id);

	CREATE TABLE IF NOT EXISTS routes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		origin_location_id TEXT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
		destination_location_id TEXT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
		distance_one_way_km REAL NOT NULL,
		distance_round_trip_km REAL NOT NULL,
		is_default BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_routes_user_id ON routes(user_id);

	CREATE TABLE IF NOT EXISTS drives (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		registered_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		route_id TEXT NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
		date TEXT NOT NULL,
		distance_km REAL NOT NULL,
		calculated_fradrag REAL NOT NULL,
		notes TEXT DEFAULT '',
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_drives_user_date ON drives(user_id, date);
	CREATE INDEX IF NOT EXISTS idx_drives_route_id ON drives(route_id);

	CREATE TABLE IF NOT EXISTS account_shares (
		id TEXT PRIMARY KEY,
		owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		shared_with_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		permission TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(owner_user_id, shared_with_user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_account_shares_owner ON account_shares(owner_user_id);
	CREATE INDEX IF NOT EXISTS idx_account_shares_shared ON account_shares(shared_with_user_id);

	CREATE TABLE IF NOT EXISTS tax_rates (
		year INTEGER PRIMARY KEY,
		threshold_km REAL NOT NULL DEFAULT 24.0,
		base_rate REAL NOT NULL DEFAULT 2.23,
		tier2_threshold_km REAL NOT NULL DEFAULT 120.0,
		tier2_rate REAL NOT NULL DEFAULT 1.12
	);

	INSERT OR IGNORE INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate)
	VALUES (2024, 24.0, 2.23, 120.0, 1.12);

	INSERT OR IGNORE INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate)
	VALUES (2025, 24.0, 2.23, 120.0, 1.12);

	INSERT OR IGNORE INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate)
	VALUES (2026, 24.0, 2.23, 120.0, 1.12);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &DB{DB: db}, nil
}
