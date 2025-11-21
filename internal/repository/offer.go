package repository

import (
	"context"
	"fmt"
	"grocery_scraper/internal/models"
	"time"

	"gorm.io/gorm"        // GORM library
	"gorm.io/gorm/clause" // Required for Upsert logic (OnConflict)
)

// OfferRepository defines the interface for persisting offer data. (Remains the same)
type OfferRepository interface {
	InsertOffers(ctx context.Context, offers []models.Offer) (int, error)
	CountOffers(ctx context.Context) (int, error)
	GetAllOffers(ctx context.Context) ([]models.Offer, error)
	// Init method for GORM AutoMigrate
	Init(ctx context.Context) error
}

// PostgresOfferRepository implements the OfferRepository interface for PostgreSQL using GORM.
type PostgresOfferRepository struct {
	db *gorm.DB // Use *gorm.DB instead of *sql.DB
}

// NewPostgresOfferRepository creates a new instance.
func NewPostgresOfferRepository(db *gorm.DB) *PostgresOfferRepository {
	return &PostgresOfferRepository{
		db: db,
	}
}

// Init handles GORM's automatic table creation/migration.
func (r *PostgresOfferRepository) Init(ctx context.Context) error {
	// AutoMigrate creates tables/columns based on the struct if they don't exist
	return r.db.WithContext(ctx).AutoMigrate(&models.Offer{})
}

// InsertOffers uses GORM to perform a bulk UPSERT (Insert or Update) operation.
func (r *PostgresOfferRepository) InsertOffers(ctx context.Context, offers []models.Offer) (int, error) {
	if len(offers) == 0 {
		return 0, nil
	}
	// Use CreateInBatches for high performance. GORM manages the transactions.
	// We wrap the operation with OnConflict clause to perform an UPSERT.
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		// Target the unique index we defined on (StoreName, Name)
		Columns: []clause.Column{{Name: "store_name"}, {Name: "name"}, {Name: "product_url"}},
		// If a conflict occurs, update all columns.
		// We use pq.StringArray in the model which handles the array serialization correctly.
		UpdateAll: true,
	}).CreateInBatches(&offers, 100) // Insert in batches of 100

	if result.Error != nil {
		return 0, fmt.Errorf("gorm bulk upsert failed: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// CountOffers returns the total number of offers in the table.
func (r *PostgresOfferRepository) CountOffers(ctx context.Context) (int, error) {
	var count int64
	// Use GORM's Model() and Count() methods
	result := r.db.WithContext(ctx).Model(&models.Offer{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("gorm count failed: %w", result.Error)
	}
	return int(count), nil
}
func (r *PostgresOfferRepository) GetAllOffers(ctx context.Context) ([]models.Offer, error) {
	var offers []models.Offer
	now := time.Now()
	// Fetches all records from the 'offers' table where valid_from <= now <= valid_to
	result := r.db.WithContext(ctx).Where("valid_from <= ? AND valid_to >= ?", now, now).Find(&offers)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve offers: %w", result.Error)
	}
	return offers, nil
}
