package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/handlers"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

func setupTestRouter(t *testing.T) (*chi.Mux, *db.DB, []byte) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory db: %v", err)
	}

	jwtSecret := []byte("test-secret-key-32-bytes-long!")

	authHandler := &handlers.AuthHandler{DB: database, JWTSecret: jwtSecret}
	locationsHandler := &handlers.LocationsHandler{DB: database}
	routesHandler := &handlers.RoutesHandler{DB: database}
	drivesHandler := &handlers.DrivesHandler{DB: database}
	statsHandler := &handlers.StatsHandler{DB: database}
	sharesHandler := &handlers.SharesHandler{DB: database}
	ratesHandler := &handlers.RatesHandler{DB: database}

	r := chi.NewRouter()

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)
		api.Get("/rates", ratesHandler.GetRates)

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireAuth(jwtSecret, database))
			protected.Get("/auth/me", authHandler.Me)

			protected.Get("/locations", locationsHandler.List)
			protected.Post("/locations", locationsHandler.Create)
			protected.Put("/locations/{id}", locationsHandler.Update)
			protected.Delete("/locations/{id}", locationsHandler.Delete)

			protected.Get("/routes", routesHandler.List)
			protected.Post("/routes", routesHandler.Create)
			protected.Put("/routes/{id}", routesHandler.Update)
			protected.Delete("/routes/{id}", routesHandler.Delete)

			protected.Get("/drives", drivesHandler.List)
			protected.Post("/drives", drivesHandler.Create)
			protected.Delete("/drives/{id}", drivesHandler.Delete)

			protected.Get("/stats", statsHandler.GetStats)
			protected.Get("/export", statsHandler.ExportCSV)

			protected.Get("/shares", sharesHandler.List)
			protected.Post("/shares", sharesHandler.Create)
			protected.Delete("/shares/{id}", sharesHandler.Delete)
		})
	})

	return r, database, jwtSecret
}

func TestCompleteFlow(t *testing.T) {
	router, database, _ := setupTestRouter(t)
	defer database.Close()

	// 1. Register Alice
	regPayload := `{"username":"alice","email":"alice@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Register failed: status %d, body: %s", rec.Code, rec.Body.String())
	}

	var aliceAuth handlers.AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &aliceAuth); err != nil {
		t.Fatalf("Failed to parse register response: %v", err)
	}

	aliceToken := aliceAuth.Token

	// 2. Register Home location
	homePayload := `{"name":"My Flat","type":"home","address":"Nørrebrogade 10, Copenhagen"}`
	req = httptest.NewRequest("POST", "/api/locations", bytes.NewBufferString(homePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create home failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var home models.Location
	_ = json.Unmarshal(rec.Body.Bytes(), &home)

	// 3. Register Workplace location
	workPayload := `{"name":"Tech Corp","type":"workplace","address":"Lautrupvang 1, Ballerup"}`
	req = httptest.NewRequest("POST", "/api/locations", bytes.NewBufferString(workPayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create workplace failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var workplace models.Location
	_ = json.Unmarshal(rec.Body.Bytes(), &workplace)

	// 4. Create Route (one way 25 km -> round trip 50 km)
	routePayload := `{"name":"Home <-> Work","origin_location_id":"` + home.ID + `","destination_location_id":"` + workplace.ID + `","distance_one_way_km":25,"distance_round_trip_km":50,"is_default":true}`
	req = httptest.NewRequest("POST", "/api/routes", bytes.NewBufferString(routePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create route failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var route models.Route
	_ = json.Unmarshal(rec.Body.Bytes(), &route)

	// 4b. Update Route (change name and distance to 60 km round trip, then revert to 50 for drive test)
	updatePayload := `{"name":"Home <-> Tech Corp HQ","origin_location_id":"` + home.ID + `","destination_location_id":"` + workplace.ID + `","distance_one_way_km":25,"distance_round_trip_km":50,"is_default":true}`
	req = httptest.NewRequest("PUT", "/api/routes/"+route.ID, bytes.NewBufferString(updatePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Update route failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var updatedRoute models.Route
	_ = json.Unmarshal(rec.Body.Bytes(), &updatedRoute)
	if updatedRoute.Name != "Home <-> Tech Corp HQ" || updatedRoute.DistanceRoundTripKm != 50.0 {
		t.Errorf("Unexpected updated route: %+v", updatedRoute)
	}

	// 5. Register Drive (50 km round trip)
	// For 50 km round trip, deduction should be (50 - 24) * 2.23 = 57.98 DKK
	drivePayload := `{"route_id":"` + route.ID + `","date":"2026-10-21","notes":"Work day"}`
	req = httptest.NewRequest("POST", "/api/drives", bytes.NewBufferString(drivePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create drive failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var drive models.Drive
	_ = json.Unmarshal(rec.Body.Bytes(), &drive)

	if drive.DistanceKm != 50.0 {
		t.Errorf("Expected distance 50.0 km, got %f", drive.DistanceKm)
	}
	if drive.CalculatedFradrag != 57.98 {
		t.Errorf("Expected calculated fradrag 57.98 DKK, got %f", drive.CalculatedFradrag)
	}

	// 6. Check Stats endpoint
	req = httptest.NewRequest("GET", "/api/stats?year=2026", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Get stats failed: status %d, body: %s", rec.Code, rec.Body.String())
	}
	var stats models.OverviewStats
	_ = json.Unmarshal(rec.Body.Bytes(), &stats)

	if stats.TotalDrives != 1 {
		t.Errorf("Expected 1 total drive, got %d", stats.TotalDrives)
	}
	if stats.TotalFradrag != 57.98 {
		t.Errorf("Expected total fradrag 57.98, got %f", stats.TotalFradrag)
	}

	// 7. Register Bob and test account sharing
	bobReg := `{"username":"bob","email":"bob@example.com","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(bobReg))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var bobAuth handlers.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &bobAuth)
	bobToken := bobAuth.Token

	// Alice shares account with Bob with write permission
	sharePayload := `{"username_or_email":"bob","permission":"write"}`
	req = httptest.NewRequest("POST", "/api/shares", bytes.NewBufferString(sharePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create share failed: status %d, body: %s", rec.Code, rec.Body.String())
	}

	// Bob uses X-Account-ID to register a drive on Alice's account
	bobDrivePayload := `{"route_id":"` + route.ID + `","date":"2026-10-22","notes":"Drive registered by Bob"}`
	req = httptest.NewRequest("POST", "/api/drives", bytes.NewBufferString(bobDrivePayload))
	req.Header.Set("Authorization", "Bearer "+bobToken)
	req.Header.Set("X-Account-ID", aliceAuth.User.ID)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Bob registering drive on Alice's account failed: status %d, body: %s", rec.Code, rec.Body.String())
	}

	// Verify Alice now has 2 drives
	req = httptest.NewRequest("GET", "/api/stats?year=2026", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	_ = json.Unmarshal(rec.Body.Bytes(), &stats)
	if stats.TotalDrives != 2 {
		t.Errorf("Expected 2 total drives for Alice, got %d", stats.TotalDrives)
	}
}
