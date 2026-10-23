package handlers_test

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/gcrg/kdb/service/pkg/dar"
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
			protected.Put("/auth/profile", authHandler.UpdateProfile)
			protected.Put("/auth/password", authHandler.ChangePassword)

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
	if route.IsHidden {
		t.Errorf("Expected newly created route to have IsHidden=false, got true")
	}

	// 4b. Update Route (change name, test hiding route)
	updatePayload := `{"name":"Home <-> Tech Corp HQ","origin_location_id":"` + home.ID + `","destination_location_id":"` + workplace.ID + `","distance_one_way_km":25,"distance_round_trip_km":50,"is_default":true,"is_hidden":true}`
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
	if updatedRoute.Name != "Home <-> Tech Corp HQ" || updatedRoute.DistanceRoundTripKm != 50.0 || !updatedRoute.IsHidden {
		t.Errorf("Unexpected updated route: %+v", updatedRoute)
	}

	// Unhide route for subsequent tests
	unhidePayload := `{"name":"Home <-> Tech Corp HQ","origin_location_id":"` + home.ID + `","destination_location_id":"` + workplace.ID + `","distance_one_way_km":25,"distance_round_trip_km":50,"is_default":true,"is_hidden":false}`
	req = httptest.NewRequest("PUT", "/api/routes/"+route.ID, bytes.NewBufferString(unhidePayload))
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Unhide route failed: status %d, body: %s", rec.Code, rec.Body.String())
	}

	// 5. Register Drive (50 km round trip)
	// For 50 km round trip in 2026, deduction is (50 - 24) * 3.17 = 82.42 DKK
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
	if drive.CalculatedFradrag != 82.42 {
		t.Errorf("Expected calculated fradrag 82.42 DKK, got %f", drive.CalculatedFradrag)
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
	if stats.TotalFradrag != 82.42 {
		t.Errorf("Expected total fradrag 82.42, got %f", stats.TotalFradrag)
	}
	if len(stats.MonthlyBreakdown) != 1 {
		t.Fatalf("Expected 1 monthly breakdown, got %d", len(stats.MonthlyBreakdown))
	}
	if len(stats.MonthlyBreakdown[0].Routes) != 1 {
		t.Fatalf("Expected 1 route in monthly breakdown, got %d", len(stats.MonthlyBreakdown[0].Routes))
	}
	if stats.MonthlyBreakdown[0].Routes[0].DrivesCount != 1 {
		t.Errorf("Expected 1 drive in route stat, got %d", stats.MonthlyBreakdown[0].Routes[0].DrivesCount)
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

func TestAccountSettings(t *testing.T) {
	router, database, _ := setupTestRouter(t)
	defer database.Close()

	// 1. Register User Charlie
	regPayload := `{"username":"charlie","email":"charlie@example.com","password":"oldpassword123"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Register failed: %d, body: %s", rec.Code, rec.Body.String())
	}
	var authResp handlers.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &authResp)
	token := authResp.Token

	// 2. Register User David (for duplicate checks)
	davidPayload := `{"username":"david","email":"david@example.com","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(davidPayload))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// 3. Test Profile Update - Duplicate Username Check
	dupPayload := `{"username":"david","email":"charlie_new@example.com"}`
	req = httptest.NewRequest("PUT", "/api/auth/profile", bytes.NewBufferString(dupPayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for duplicate username, got %d", rec.Code)
	}

	// 4. Test Profile Update - Valid Update
	validUpdate := `{"username":"charlie_updated","email":"charlie_updated@example.com"}`
	req = httptest.NewRequest("PUT", "/api/auth/profile", bytes.NewBufferString(validUpdate))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid profile update, got %d: %s", rec.Code, rec.Body.String())
	}
	var updatedAuth handlers.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &updatedAuth)
	if updatedAuth.User.Username != "charlie_updated" || updatedAuth.User.Email != "charlie_updated@example.com" {
		t.Errorf("Profile update mismatch: %+v", updatedAuth.User)
	}
	token = updatedAuth.Token // use new token

	// 5. Test Change Password - Wrong Current Password
	wrongPass := `{"current_password":"wrongpassword","new_password":"newsecurepassword"}`
	req = httptest.NewRequest("PUT", "/api/auth/password", bytes.NewBufferString(wrongPass))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for wrong current password, got %d", rec.Code)
	}

	// 6. Test Change Password - Valid Current Password
	correctPass := `{"current_password":"oldpassword123","new_password":"newsecurepassword"}`
	req = httptest.NewRequest("PUT", "/api/auth/password", bytes.NewBufferString(correctPass))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid password change, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Verify login with new password and updated username
	loginPayload := `{"username_or_email":"charlie_updated","password":"newsecurepassword"}`
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for login with new password, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDARResolutionAndYderkommuneFlow(t *testing.T) {
	// 1. Mock DAR Server returning Guldborgsund ("0376-0042")
	darServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"DAR_Husnummer": {
					"nodes": [
						{
							"adgangsadressebetegnelse": "Strandvejen 1, 4800 Nykøbing F",
							"vejmidte": "0376-0042",
							"status": "gældende"
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer darServer.Close()

	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer database.Close()

	jwtSecret := []byte("test-secret-key-32-bytes-long!")
	darClient := dar.NewClient("mock-token")
	darClient.Endpoint = darServer.URL

	authHandler := &handlers.AuthHandler{DB: database, JWTSecret: jwtSecret}
	locationsHandler := &handlers.LocationsHandler{DB: database, DARClient: darClient}
	routesHandler := &handlers.RoutesHandler{DB: database}
	drivesHandler := &handlers.DrivesHandler{DB: database}

	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/register", authHandler.Register)
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireAuth(jwtSecret, database))
			protected.Post("/locations", locationsHandler.Create)
			protected.Post("/routes", routesHandler.Create)
			protected.Post("/drives", drivesHandler.Create)
		})
	})

	// Register user
	regPayload := `{"username":"dansk_pendler","email":"pendler@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var authResp handlers.AuthResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &authResp)
	token := authResp.Token

	// Create home location in Guldborgsund
	homePayload := `{"name":"Hjem","type":"home","address":"Strandvejen 1, 4800 Nykøbing F"}`
	req = httptest.NewRequest("POST", "/api/locations", bytes.NewBufferString(homePayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create home failed: %d %s", rec.Code, rec.Body.String())
	}
	var homeLoc models.Location
	_ = json.Unmarshal(rec.Body.Bytes(), &homeLoc)

	if homeLoc.KommuneCode != "0376" {
		t.Errorf("got KommuneCode %s, want 0376", homeLoc.KommuneCode)
	}
	if homeLoc.KommuneName != "Guldborgsund" {
		t.Errorf("got KommuneName %s, want Guldborgsund", homeLoc.KommuneName)
	}
	if !homeLoc.IsYderkommune {
		t.Errorf("got IsYderkommune false, want true")
	}

	// Create workplace location
	workPayload := `{"name":"Arbejde","type":"workplace","address":"Vordingborgvej 10, 4700 Næstved"}`
	req = httptest.NewRequest("POST", "/api/locations", bytes.NewBufferString(workPayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var workLoc models.Location
	_ = json.Unmarshal(rec.Body.Bytes(), &workLoc)

	// Create 50 km route
	routePayload := `{"name":"Hjem-Arbejde","origin_location_id":"` + homeLoc.ID + `","destination_location_id":"` + workLoc.ID + `","distance_one_way_km":25,"distance_round_trip_km":50,"is_default":true}`
	req = httptest.NewRequest("POST", "/api/routes", bytes.NewBufferString(routePayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var route models.Route
	_ = json.Unmarshal(rec.Body.Bytes(), &route)

	// Log drive in 2026 for 50 km: yderkommune rate is 3.51 -> (50-24)*3.51 = 91.26 DKK
	drivePayload := `{"route_id":"` + route.ID + `","date":"2026-05-10"}`
	req = httptest.NewRequest("POST", "/api/drives", bytes.NewBufferString(drivePayload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var drive50 models.Drive
	_ = json.Unmarshal(rec.Body.Bytes(), &drive50)
	if math.Abs(drive50.CalculatedFradrag-91.26) > 0.001 {
		t.Errorf("got 50km fradrag %f, want 91.26", drive50.CalculatedFradrag)
	}

	// Log drive in 2026 for 150 km: yderkommune rate is 3.51 regardless of distance -> (150-24)*3.51 = 442.26 DKK
	drive150Payload := `{"route_id":"` + route.ID + `","date":"2026-05-11","distance_km":150}`
	req = httptest.NewRequest("POST", "/api/drives", bytes.NewBufferString(drive150Payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var drive150 models.Drive
	_ = json.Unmarshal(rec.Body.Bytes(), &drive150)
	if math.Abs(drive150.CalculatedFradrag-442.26) > 0.001 {
		t.Errorf("got 150km fradrag %f, want 442.26", drive150.CalculatedFradrag)
	}
}

