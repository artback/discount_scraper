package main

import (
	"context"
	"encoding/json"
	"grocery_scraper/internal/config"
	"grocery_scraper/internal/repository"
	"log"
	"net/http"
	"time" // Required for context timeout

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Configuration and Initialization ---

// Constants for config keys (reused from scraper main)
const (
	DBHostKey     = "DB_HOST"
	DBPortKey     = "DB_PORT"
	DBUserKey     = "DB_USER"
	DBPasswordKey = "DB_PASSWORD"
	DBNameKey     = "DB_NAME"
)

// initDatabase establishes a connection and initializes the repository.
func initDatabase(dsn string) repository.OfferRepository {
	var offerRepo repository.OfferRepository
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Fatal Error: Could not connect to the database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL for API server.")

	// Initialize the global repository instance
	offerRepo = repository.NewPostgresOfferRepository(db)

	// Optional: Check if the table is ready (Init() handles migration)
	if err := offerRepo.Init(context.Background()); err != nil {
		log.Fatalf("Fatal Error: Database migration failed: %v", err)
	}
	return offerRepo
}

type OfferApi struct {
	offerRepository repository.OfferRepository
}

// offersHandler fetches data directly from the database repository and serves it as JSON.
func (o OfferApi) offersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	offers, err := o.offerRepository.GetAllOffers(ctx)
	if err != nil {
		http.Error(w, "Could not retrieve data from the database", http.StatusInternalServerError)
		log.Printf("Error fetching offers: %v", err)
		return
	}

	if err := json.NewEncoder(w).Encode(offers); err != nil {
		http.Error(w, "Could not send JSON data", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
	}
}

// indexHandler serves the main page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func main() {
	ctx := context.Background()
	const port = "8080"
	conf := config.Init()
	// 1. Initialize Database Connection and Repository
	database := initDatabase(conf.DBConn)
	api := OfferApi{database}
	// 2. Set up Handlers
	http.HandleFunc("/", indexHandler)                // Serves the homepage
	http.HandleFunc("/api/offers", api.offersHandler) // Serves the JSON data
	count, err := database.CountOffers(ctx)
	if err != nil {
		log.Fatalf("Error counting offers: %v", err)
	}
	log.Printf("Serving %d offers", count)
	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
