package models

import "gorm.io/gorm"

// Store struct holds the display name and the unique URL slug for the store.
type Store struct {
	Name    string `mapstructure:"name"`
	URLSlug string `mapstructure:"url_slug"`
}

// Offer struct maps to the 'offers' database table via GORM tags
type Offer struct {
	// GORM will automatically add ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model

	StoreName  string `json:"storeName" gorm:"type:varchar(100);uniqueIndex:idx_store_name_product_name"` // Primary key for Upsert logic
	Name       string `json:"name" gorm:"type:varchar(255);uniqueIndex:idx_store_name_product_name"`      // Primary key for Upsert logic
	ProductURL string `json:"productURL" gorm:"type:text;uniqueIndex:idx_store_name_product_name"`
	Type       string `json:"type" gorm:"type:varchar(50)"`

	// Use pointers for omitempty/nullable fields in the DB if they can be nil
	OriginalPrice      float64 `json:"originalPrice,omitempty" gorm:"type:numeric(10, 2)"`
	SalePrice          float64 `json:"salePrice,omitempty" gorm:"type:numeric(10, 2);not null"`
	SaleQuantity       int     `json:"saleQuantity,omitempty"`
	SalePriceTotal     float64 `json:"salePriceTotal,omitempty" gorm:"type:numeric(10, 2)"`
	Discount           int     `json:"discount,omitempty"`
	DiscountPercentage float64 `json:"discountPercentage" gorm:"type:numeric(5, 2)"`
}
