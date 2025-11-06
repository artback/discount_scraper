package parser

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"log"
	"strings"
)

// RawOffer is a Data Transfer Object (DTO) used internally to pass the raw,
// unparsed string data from the scraper to the service's business logic.
type RawOffer struct {
	PromotionID  string
	Name         string
	OriginalText string
	DealText     string
}

// OfferParser defines the contract for scraping and extracting raw offer data
// from the HTML source. It knows how to read the HTML structure.
type OfferParser interface {
	ParseRawOffers(ctx context.Context, reader io.Reader) ([]RawOffer, error)
}

// icaDealParser is the concrete implementation of the scraping logic.
type icaDealParser struct {
}

// NewOfferParser creates a new parser instance with a repository dependency for fetching HTML.
func NewOfferParser() OfferParser {
	return &icaDealParser{}
}

// ParseRawOffers fetches the rendered HTML and extracts only the required string data
// (name, original price text, deal price text) for each offer card.
func (p *icaDealParser) ParseRawOffers(ctx context.Context, reader io.Reader) ([]RawOffer, error) {
	// 1. Fetch the HTML content
	htmlReader := reader

	// 2. Parse the HTML document
	doc, err := html.Parse(htmlReader)
	if err != nil {
		// Ensure the reader is closed if we received one (Repo should handle cleanup)
		if closer, ok := htmlReader.(io.Closer); ok {
			closer.Close()
		}
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var rawOffers []RawOffer
	// 3. Use goquery to traverse and extract raw strings
	goquery.NewDocumentFromNode(doc).Find("article").Each(func(i int, sel *goquery.Selection) {
		promotionID, exists := sel.Attr("data-promotion-id")
		if !exists {
			// Skip this article if it's not an offer card
			return
		}

		name := strings.TrimSpace(sel.Find(".offer-card__title").Text())
		if name == "" {
			log.Printf("Missing name for promotion ID: %s. Skipping.", promotionID)
			return
		}

		rawOffers = append(rawOffers, RawOffer{
			PromotionID:  promotionID,
			Name:         name,
			OriginalText: sel.Find(".offer-card__text").Text(),
			DealText:     strings.ToLower(sel.Find(".price-splash__text").Text()),
		})
	})

	return rawOffers, nil
}
