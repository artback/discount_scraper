package service

import (
	"context"
	"fmt"
	"grocery_scraper/internal/models"
	"grocery_scraper/internal/parser"
	"grocery_scraper/internal/repository"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const ICA_BASE_URL = "https://www.ica.se/erbjudanden"

// OfferService defines the business logic contract.
type OfferService interface {
	GetStoreOffers(ctx context.Context, store models.Store) ([]models.Offer, error)
}

// offerService is the concrete service implementation
// It now depends on the Repository (for low-level data access/fetching)
// and the OfferExtractor (for raw HTML parsing).
type offerService struct {
	Repo   repository.ICARepository
	Parser parser.OfferParser // <-- Updated dependency name and type
}

// NewOfferService creates a new service instance with both dependencies.
func NewOfferService(repo repository.ICARepository, extractor parser.OfferParser) OfferService {
	return &offerService{
		Repo:   repo,
		Parser: extractor,
	}
}

// --- Regular Expressions (Business Logic/Transformation) ---
// These are now clearly part of the business transformation layer.
var (
	// Matches 'Ord.pris X kr' - Flexible spacing, captures price part
	originalPriceRegex = regexp.MustCompile(`Ord\.pris\s*([\d:,-]+)`)

	// Matches single prices for calculation (e.g., "25:-/st" or "7990/kg"). Captures the numerical part.
	singlePriceRegex = regexp.MustCompile(`(\d+[\.,]?\d*)\s*(?:[-:\/]|kr|st|kg)*`)

	// Matches 'X för Y kr'. Captures quantity (Group 1) and total price (Group 2).
	multibuyRegex = regexp.MustCompile(`(\d+)\s*för\s*([\d\s:,\.]+)`)

	// Matches percentage deals like '20%'
	percentageRegex = regexp.MustCompile(`(\d+)%`)
)

// --- Utility Functions (Data Transformation) ---

// cleanAndParse removes currency separators and whitespace to prepare for float conversion.
func cleanAndParse(priceStr string) float64 {
	cleanStr := strings.ReplaceAll(priceStr, ":", ".")
	cleanStr = strings.ReplaceAll(cleanStr, ",", ".")
	reStrip := regexp.MustCompile(`[^\d\.\s]`)
	cleanStr = reStrip.ReplaceAllString(cleanStr, "")
	cleanStr = strings.ReplaceAll(cleanStr, " ", "")

	price, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		return 0
	}
	// Assuming prices > 1000 might be in öre (e.g., 7990 for 79.90), common in older scraping targets
	if price > 1000 && !strings.Contains(priceStr, ":") {
		return price / 100
	}
	return price
}

// parsePrice handles single prices, price ranges (X-Y), and returns the average for a range.
func parsePrice(priceString string) float64 {
	if priceString == "" {
		return 0
	}
	if strings.Contains(priceString, "-") {
		// Attempt to extract the numeric price parts in case of a range
		reRange := regexp.MustCompile(`[\d:,\.]+`)
		matches := reRange.FindAllString(priceString, 2)

		if len(matches) == 2 {
			priceMin := cleanAndParse(matches[0])
			priceMax := cleanAndParse(matches[1])

			if priceMin > 0 && priceMax > 0 {
				return (priceMin + priceMax) / 2
			}
		}
	}
	return cleanAndParse(priceString)
}

// calculateDiscount computes the discount percentage based on the offer type.
func calculateDiscount(offer models.Offer) float64 {
	if offer.Type == "percentage" {
		// This already returns a percentage, which can be negative/positive
		return float64(offer.Discount)
	}

	var originalTotal float64
	var saleTotal float64

	if offer.Type == "single" {
		originalTotal = offer.OriginalPrice
		saleTotal = offer.SalePrice
	} else if offer.Type == "multibuy" {
		originalTotal = offer.OriginalPrice * float64(offer.SaleQuantity)
		saleTotal = offer.SalePriceTotal
	} else {
		return 0
	}

	if originalTotal == 0 {
		return 0.0
	}

	// Calculate the discount percentage (will be negative if saleTotal > originalTotal)
	discount := ((originalTotal - saleTotal) / originalTotal) * 100

	// Use math.Round() for robust and mathematically correct rounding
	// for both positive and negative numbers.
	roundedDiscount := math.Round(discount*100) / 100

	return roundedDiscount
}

// GetStoreOffers orchestrates the fetching, parsing, transformation, and calculation steps.
func (s *offerService) GetStoreOffers(ctx context.Context, store models.Store) ([]models.Offer, error) {
	// 1. Fetch HTML content (Repository responsibility)
	storeURLStr := fmt.Sprintf("%s/%s", ICA_BASE_URL, store.URLSlug)
	htmlReader, err := s.Repo.Fetch(ctx, storeURLStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rendered HTML for %s: %w", store.Name, err)
	}

	// Ensure the reader is closed if it's an io.Closer
	if closer, ok := htmlReader.(io.Closer); ok {
		defer closer.Close()
	}

	// 2. Parse Raw Data (Parser responsibility)
	rawDeals, err := s.Parser.ParseRawOffers(ctx, htmlReader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract raw offers for %s: %w", store.Name, err)
	}

	var offers []models.Offer
	// 3. Transform Raw Data into structured Offers (Service Business Logic)
	for _, rawDeal := range rawDeals {
		// Extract Original Price
		originalPrice := 0.0
		if match := originalPriceRegex.FindStringSubmatch(rawDeal.OriginalText); len(match) > 1 {
			originalPrice = parsePrice(match[1])
		}

		deal := models.Offer{
			StoreName:     store.Name,
			Name:          rawDeal.Name,
			OriginalPrice: originalPrice,
			Type:          "unknown",
		}
		// Construct the final, usable URL
		deal.ProductURL = fmt.Sprintf("%s/%s?id=%s&action=details", ICA_BASE_URL, store.URLSlug, rawDeal.PromotionID)

		// Determine Offer Type and Extract Sale Details
		if percentageMatch := percentageRegex.FindStringSubmatch(rawDeal.DealText); len(percentageMatch) > 1 {
			deal.Type = "percentage"
			deal.Discount, _ = strconv.Atoi(percentageMatch[1])
		} else if multibuyMatch := multibuyRegex.FindStringSubmatch(rawDeal.DealText); len(multibuyMatch) > 2 {
			deal.Type = "multibuy"
			deal.SaleQuantity, _ = strconv.Atoi(multibuyMatch[1])
			deal.SalePriceTotal = parsePrice(multibuyMatch[2])
		} else if singlePriceMatch := singlePriceRegex.FindStringSubmatch(rawDeal.DealText); len(singlePriceMatch) > 1 {
			deal.Type = "single"
			deal.SalePrice = parsePrice(singlePriceMatch[1])
		}

		// Calculate Final Discount Percentage
		deal.DiscountPercentage = calculateDiscount(deal)

		offers = append(offers, deal)
	}

	return offers, nil
}
