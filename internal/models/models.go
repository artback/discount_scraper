package models

import "gorm.io/gorm"

// Store struct holds the display name and the unique URL slug for the store.
type Store struct {
	Name    string `mapstructure:"name"`
	URLSlug string `mapstructure:"url_slug"`
}

// Offer represents an offer for a product.
//
// swagger:model Offer
type Offer struct {
	// GORM will automatically add ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model

	// the name of the store
	//
	// required: true
	StoreName string `json:"storeName" gorm:"type:varchar(100);uniqueIndex:idx_store_name_product_name"`
	// the name of the product
	//
	// required: true
	Name string `json:"name" gorm:"type:varchar(255);uniqueIndex:idx_store_name_product_name"`
	// the url of the product
	//
	// required: true
	ProductURL string `json:"productURL" gorm:"type:varchar(2048);uniqueIndex:idx_store_name_product_name"`
	// the type of the offer
	//
	// required: true
	Type string `json:"type" gorm:"type:varchar(50)"`

	// Use pointers for omitempty/nullable fields in the DB if they can be nil
	// the original price of the product
	OriginalPrice float64 `json:"originalPrice,omitempty" gorm:"type:numeric(10, 2)"`
	// the sale price of the product
	//
	// required: true
	SalePrice float64 `json:"salePrice,omitempty" gorm:"type:numeric(10, 2);not null"`
	// the quantity of the product for the sale price
	SaleQuantity int `json:"saleQuantity,omitempty"`
	// the total price of the sale
	SalePriceTotal float64 `json:"salePriceTotal,omitempty" gorm:"type:numeric(10, 2)"`
	// the discount of the product
	Discount int `json:"discount,omitempty"`
	// the discount percentage of the product
	DiscountPercentage float64 `json:"discountPercentage" gorm:"type:numeric(5, 2)"`
}
