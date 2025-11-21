package models

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type to handle []string mapping to Postgres text[]
type StringArray []string

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	// Postgres driver might return []byte for text[]
	var strVal string
	switch v := value.(type) {
	case []byte:
		strVal = string(v)
	case string:
		strVal = v
	default:
		return errors.New("failed to scan StringArray: value is not string or []byte")
	}

	// Basic parsing for Postgres array format: "{a,b,c}"
	// This is a simplified parser. For robust parsing of complex strings with commas/quotes,
	// a full parser is needed. However, for simple categories, this might suffice.
	// Alternatively, we can use JSON serialization if we change the column type to jsonb,
	// but the schema is text[].

	// Trim curly braces
	strVal = strings.Trim(strVal, "{}")
	if strVal == "" {
		*a = []string{}
		return nil
	}

	// Split by comma (Note: this fails if categories contain commas)
	// A more robust way without lib/pq is tricky.
	// Let's try to handle quoted strings if possible, but for now simple split.
	parts := strings.Split(strVal, ",")
	for i, part := range parts {
		// Trim quotes if present (Postgres adds quotes if needed)
		parts[i] = strings.Trim(part, "\"")
	}
	*a = StringArray(parts)
	return nil
}

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}

	// Serialize to Postgres array format: "{a,b,c}"
	// We need to escape values if they contain special characters.
	var parts []string
	for _, s := range a {
		// Escape backslashes and double quotes
		escaped := strings.ReplaceAll(s, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		// Wrap in double quotes
		parts = append(parts, "\""+escaped+"\"")
	}

	return "{" + strings.Join(parts, ",") + "}", nil
}

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
	OriginalPrice float64 `json:"originalPrice" gorm:"type:numeric(10, 2)"`
	// the sale price of the product
	//
	// required: true
	SalePrice float64 `json:"salePrice" gorm:"type:numeric(10, 2);not null"`
	// the quantity of the product for the sale price
	SaleQuantity int `json:"saleQuantity"`
	// the total price of the sale
	SalePriceTotal float64 `json:"salePriceTotal" gorm:"type:numeric(10, 2)"`
	// the discount of the product
	Discount int `json:"discount"`
	// the discount percentage of the product
	DiscountPercentage float64 `json:"discountPercentage" gorm:"type:numeric(5, 2)"`

	// Categories for the product
	Categories StringArray `json:"categories" gorm:"type:text[]"`

	// Validity period of the offer
	ValidFrom time.Time `json:"validFrom" gorm:"index"`
	ValidTo   time.Time `json:"validTo" gorm:"index"`
}
