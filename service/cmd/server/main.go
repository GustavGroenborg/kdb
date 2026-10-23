package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/gcrg/kdb/service/pkg/dar"
	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/handlers"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/static"
)

func main() {
	var (
		portFlag      = flag.String("port", "", "Server HTTP port")
		dbPathFlag    = flag.String("db", "", "Path to SQLite database file")
		staticDirFlag = flag.String("static", "", "Directory containing static frontend files")
	)
	flag.Parse()

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	dbPath := *dbPathFlag
	if dbPath == "" {
		dbPath = os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			dbPath = "database/kdb.db"
		}
	}

	staticDir := *staticDirFlag
	if staticDir == "" {
		staticDir = os.Getenv("STATIC_DIR")
		if staticDir == "" {
			staticDir = "ui/dist"
		}
	}

	jwtSecretStr := os.Getenv("JWT_SECRET")
	var jwtSecret []byte
	if jwtSecretStr != "" {
		jwtSecret = []byte(jwtSecretStr)
	} else {
		// Generate random ephemeral secret if not set
		randBytes := make([]byte, 32)
		_, _ = rand.Read(randBytes)
		jwtSecret = []byte(hex.EncodeToString(randBytes))
	}

	log.Printf("Connecting to database at %s ...", dbPath)
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.Close()
	log.Printf("Database initialized successfully.")

	darToken := dar.LoadDARToken()
	if darToken != "" {
		log.Printf("DAR API token loaded successfully.")
	} else {
		log.Printf("DAR API token not configured (address municipality resolution disabled).")
	}
	darClient := dar.NewClient(darToken)

	// Handlers
	authHandler := &handlers.AuthHandler{DB: database, JWTSecret: jwtSecret}
	locationsHandler := &handlers.LocationsHandler{DB: database, DARClient: darClient}
	routesHandler := &handlers.RoutesHandler{DB: database}
	drivesHandler := &handlers.DrivesHandler{DB: database}
	statsHandler := &handlers.StatsHandler{DB: database}
	sharesHandler := &handlers.SharesHandler{DB: database}
	ratesHandler := &handlers.RatesHandler{DB: database}
	adminHandler := &handlers.AdminHandler{DB: database}

	r := chi.NewRouter()

	// Global middlewares
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// CORS configuration for UI development
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Account-ID"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Routes
	r.Route("/api", func(api chi.Router) {
		// Public Auth
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)
		api.Post("/auth/logout", authHandler.Logout)

		// Public Tax Rates
		api.Get("/rates", ratesHandler.GetRates)
		api.Get("/rates/{year}", ratesHandler.GetRates)

		// Protected API routes
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireAuth(jwtSecret, database))

			// User info & settings
			protected.Get("/auth/me", authHandler.Me)
			protected.Put("/auth/profile", authHandler.UpdateProfile)
			protected.Put("/auth/password", authHandler.ChangePassword)

			// Locations (Home, Workplaces, Universities)
			protected.Get("/locations", locationsHandler.List)
			protected.Post("/locations", locationsHandler.Create)
			protected.Put("/locations/{id}", locationsHandler.Update)
			protected.Delete("/locations/{id}", locationsHandler.Delete)
			protected.Get("/addresses/autocomplete", locationsHandler.AutocompleteAddress)

			// Routes
			protected.Get("/routes", routesHandler.List)
			protected.Post("/routes", routesHandler.Create)
			protected.Put("/routes/{id}", routesHandler.Update)
			protected.Delete("/routes/{id}", routesHandler.Delete)

			// Drives
			protected.Get("/drives", drivesHandler.List)
			protected.Post("/drives", drivesHandler.Create)
			protected.Delete("/drives/{id}", drivesHandler.Delete)
			protected.Post("/drives/recalculate", drivesHandler.Recalculate)

			// Stats & Export
			protected.Get("/stats", statsHandler.GetStats)
			protected.Get("/export", statsHandler.ExportCSV)

			// Account Sharing
			protected.Get("/shares", sharesHandler.List)
			protected.Post("/shares", sharesHandler.Create)
			protected.Delete("/shares/{id}", sharesHandler.Delete)

			// Tax Rates adjustment
			protected.Post("/rates", ratesHandler.UpdateRates)

			// Admin Telemetry (GDPR-compliant, restricted to administrators)
			protected.Group(func(admin chi.Router) {
				admin.Use(middleware.RequireAdmin)
				admin.Get("/admin/metrics", adminHandler.GetMetrics)
			})
		})
	})

	// Static UI serving (embedded assets prioritized, fallback to filesystem directory)
	r.Handle("/*", static.Handler(getEmbeddedFS(), staticDir))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚗 KDB Befordringsfradrag Service running on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	<-shutdown
	log.Println("Shutting down KDB server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("KDB server stopped.")
}
