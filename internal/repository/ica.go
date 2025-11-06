package repository

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"grocery_scraper/pkg/headless"
	"io"
	"log"
	"net/http"
	"strconv"
)

const (
	ICA_OFFER_CARD_SELECTOR       = "article"
	ICA_LIST_LENGTH_ATTR          = "data-promotion-list-length"
	ICA_OFFERS_CONTAINER_SELECTOR = ".offers__container"
)

// ICARepository defines the contract for fetching ICA data.
// This is the interface you would mock for testing.
type ICARepository interface {
	Fetch(ctx context.Context, url string) (io.Reader, error)
}

// icaRepositoryImpl is the concrete implementation that performs HTTP requests.
type icaRepositoryImpl struct {
	Client *http.Client
}

// NewICARepository creates and returns a new repository instance.
func NewICARepository() ICARepository {
	return &icaRepositoryImpl{
		Client: &http.Client{},
	}
}

func (r *icaRepositoryImpl) Fetch(ctx context.Context, url string) (io.Reader, error) {
	return headless.FetchRenderedContent(ctx, url, ICAOfferWaitStrategy, ICA_OFFERS_CONTAINER_SELECTOR)
}

// ICAOfferWaitStrategy implements the specific logic for the ICA site.
func ICAOfferWaitStrategy(ctx context.Context, url string) error {
	var listLengthStr string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {get: () => false, configurable: true});`, nil),
		chromedp.WaitVisible(ICA_OFFER_CARD_SELECTOR, chromedp.ByQuery),
		chromedp.AttributeValue(ICA_OFFER_CARD_SELECTOR, ICA_LIST_LENGTH_ATTR, &listLengthStr, nil),
	)

	if err != nil {
		return fmt.Errorf("could not navigate or read list length from '%s': %w", url, err)
	}

	// Convert the string length to an integer.
	listLength, err := strconv.Atoi(listLengthStr)
	if err != nil || listLength <= 0 {
		return fmt.Errorf("could not parse valid list length from attribute value '%s'", listLengthStr)
	}

	log.Printf("ICA strategy: Waiting for %d offers to render.", listLength)

	// --- Phase 2: Wait for all offers to render ---
	lastItemSelector := fmt.Sprintf("%s:nth-child(%d)", ICA_OFFER_CARD_SELECTOR, listLength)

	// This is the custom wait task specific to the ICA page structure.
	waitTask := chromedp.Tasks{
		chromedp.WaitVisible(lastItemSelector, chromedp.ByQuery),
	}

	if err = chromedp.Run(ctx, waitTask); err != nil {
		return fmt.Errorf("timed out waiting for the last offer (selector: %s): %w", lastItemSelector, err)
	}

	return nil
}
