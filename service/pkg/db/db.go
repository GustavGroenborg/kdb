package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Configure connection pool for SQLite in WAL mode
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

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
		is_admin BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS locations (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		address TEXT NOT NULL,
		is_primary_home BOOLEAN NOT NULL DEFAULT 0,
		kommune_code TEXT DEFAULT '',
		kommune_name TEXT DEFAULT '',
		is_yderkommune BOOLEAN NOT NULL DEFAULT 0,
		is_dar_approved BOOLEAN NOT NULL DEFAULT 0,
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
		is_hidden BOOLEAN NOT NULL DEFAULT 0,
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
		tier2_rate REAL NOT NULL DEFAULT 1.12,
		yderkommune_rate REAL NOT NULL DEFAULT 2.23
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Idempotent column migrations for existing databases
	var hasIsAdmin bool
	if rows, err := db.Query("PRAGMA table_info(users);"); err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil && name == "is_admin" {
				hasIsAdmin = true
				break
			}
		}
		rows.Close()
	}
	if !hasIsAdmin {
		_, _ = db.Exec("ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT 0;")
	}

	// Bootstrap Admin privileges:
	// 1) Environment-configured admin users
	if adminList := os.Getenv("KDB_ADMIN_USERS"); adminList != "" {
		for _, u := range strings.Split(adminList, ",") {
			trimmed := strings.TrimSpace(u)
			if trimmed != "" {
				_, _ = db.Exec("UPDATE users SET is_admin = 1 WHERE username = ? OR email = ?", trimmed, strings.ToLower(trimmed))
			}
		}
	}
	// 2) Auto-promote earliest registered user if no admin exists yet
	var adminCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1;").Scan(&adminCount)
	if adminCount == 0 {
		_, _ = db.Exec("UPDATE users SET is_admin = 1 WHERE id = (SELECT id FROM users ORDER BY created_at ASC LIMIT 1);")
	}

	var hasIsHidden bool
	if rows, err := db.Query("PRAGMA table_info(routes);"); err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil && name == "is_hidden" {
				hasIsHidden = true
				break
			}
		}
		rows.Close()
	}
	if !hasIsHidden {
		_, _ = db.Exec("ALTER TABLE routes ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT 0;")
	}

	// Idempotent column migrations for locations table
	locationCols := make(map[string]bool)
	if rows, err := db.Query("PRAGMA table_info(locations);"); err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
				locationCols[name] = true
			}
		}
		rows.Close()
	}
	if !locationCols["kommune_code"] {
		_, _ = db.Exec("ALTER TABLE locations ADD COLUMN kommune_code TEXT DEFAULT '';")
	}
	if !locationCols["kommune_name"] {
		_, _ = db.Exec("ALTER TABLE locations ADD COLUMN kommune_name TEXT DEFAULT '';")
	}
	if !locationCols["is_yderkommune"] {
		_, _ = db.Exec("ALTER TABLE locations ADD COLUMN is_yderkommune BOOLEAN NOT NULL DEFAULT 0;")
	}
	if !locationCols["is_dar_approved"] {
		_, _ = db.Exec("ALTER TABLE locations ADD COLUMN is_dar_approved BOOLEAN NOT NULL DEFAULT 0;")
		_, _ = db.Exec("UPDATE locations SET is_dar_approved = 1 WHERE kommune_code IS NOT NULL AND kommune_code != '';")
	}

	// Idempotent column migrations for tax_rates table
	var hasYderkommuneRate bool
	if rows, err := db.Query("PRAGMA table_info(tax_rates);"); err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil && name == "yderkommune_rate" {
				hasYderkommuneRate = true
				break
			}
		}
		rows.Close()
	}
	if !hasYderkommuneRate {
		_, _ = db.Exec("ALTER TABLE tax_rates ADD COLUMN yderkommune_rate REAL NOT NULL DEFAULT 2.23;")
	}

	// Upsert official tax rates
	seedRates := `
	INSERT INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate, yderkommune_rate)
	VALUES (2024, 24.0, 2.23, 120.0, 1.12, 2.23)
	ON CONFLICT(year) DO UPDATE SET base_rate=excluded.base_rate, tier2_rate=excluded.tier2_rate, yderkommune_rate=excluded.yderkommune_rate;

	INSERT INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate, yderkommune_rate)
	VALUES (2025, 24.0, 2.23, 120.0, 1.12, 2.23)
	ON CONFLICT(year) DO UPDATE SET base_rate=excluded.base_rate, tier2_rate=excluded.tier2_rate, yderkommune_rate=excluded.yderkommune_rate;

	INSERT INTO tax_rates (year, threshold_km, base_rate, tier2_threshold_km, tier2_rate, yderkommune_rate)
	VALUES (2026, 24.0, 3.17, 120.0, 1.59, 3.51)
	ON CONFLICT(year) DO UPDATE SET base_rate=excluded.base_rate, tier2_rate=excluded.tier2_rate, yderkommune_rate=excluded.yderkommune_rate;
	`
	if _, err := db.Exec(seedRates); err != nil {
		return nil, fmt.Errorf("failed to seed tax rates: %w", err)
	}


	return &DB{DB: db}, nil
}
