package main

import (
	"context"
	"fmt"
	"grocery_scraper/internal/config"
	"grocery_scraper/internal/parser"
	"grocery_scraper/internal/repository"
	"grocery_scraper/internal/service"
	"log"

	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Main Application Logic ---
func main() {
	// 1. Load configuration
	appConfig := config.Init()
	dsn := appConfig.DBConn
	targetStores := appConfig.Stores // Get stores from the config struct

	if len(targetStores) == 0 {
		log.Fatal("No target stores configured. Please add stores to config.yaml or check defaults.")
	}

	// 2. Database Connection (using GORM)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatalf("Error connecting to database with GORM using DSN '%s': %v", dsn, err)
	}
	log.Println("Successfully connected to PostgreSQL using GORM!")

	// 3. Dependency Injection: Initialize components
	icaRepo := repository.NewICARepository()
	offerRepo := repository.NewPostgresOfferRepository(db)

	// 4. Database Migration
	ctx := context.Background()
	if err := offerRepo.Init(ctx); err != nil {
		log.Fatalf("Failed to run database auto-migration: %v", err)
	}
	log.Println("Database structure verified/migrated successfully.")

	// Initialize AI Categorizer
	var categorizer service.Categorizer
	if appConfig.AIAPIKey != "" {
		var err error
		categorizer, err = service.NewAICategorizer(ctx, appConfig.AIAPIKey)
		if err != nil {
			log.Printf("Warning: Failed to initialize AI Categorizer: %v. Categorization will be skipped.", err)
		} else {
			log.Println("AI Categorizer initialized successfully.")
			if aiCat, ok := categorizer.(*service.AICategorizer); ok {
				defer aiCat.Close()
			}
		}
	} else {
		log.Println("No AI API key provided. Categorization will be skipped.")
	}

	par := parser.NewOfferParser()
	offerService := service.NewOfferService(icaRepo, par, categorizer)

	// Initialize the errgroup.Group
	g, gCtx := errgroup.WithContext(ctx)

	// 5. Execution Loop: Scrape and Save in parallel
	for _, store := range targetStores {
		g.Go(func() error {
			log.Printf("Starting scrape for: %s", store.Name)

			// Use the context from the errgroup for scrape calls
			offers, err := offerService.GetStoreOffers(ctx, store)
			if err != nil {
				return fmt.Errorf("error scraping %s: %w", store.Name, err)
			}

			log.Printf("Successfully processed %d offers from %s. Starting insertion...", len(offers), store.Name)
			//Use the context from the errgroup for insertion calls
			insertedOrUpdatedCount, err := offerRepo.InsertOffers(gCtx, offers)
			if err != nil {
				return fmt.Errorf("error inserting offers for %s: %w", store.Name, err)
			}

			log.Printf("Successfully inserted/updated %d offers from %s", insertedOrUpdatedCount, store.Name)
			return nil
		})
	}

	// 6. Wait for all goroutines to complete.
	if err := g.Wait(); err != nil {
		log.Fatalf("One or more scraping/insertion tasks failed: %v", err)
	}

	// 7. Final Output
	totalCount, err := offerRepo.CountOffers(ctx)
	if err != nil {
		log.Printf("Warning: Could not get final offer count from DB: %v", err)
	}

	fmt.Printf("\n--- SCRAPE AND PERSISTENCE COMPLETE (via GORM) ---\n")
	fmt.Printf("Successfully scraped and saved/updated a total of %d offers to PostgreSQL.\n", totalCount)
}
