package headless

import (
	"bytes"
	"context"
	"fmt"
	"github.com/DataHenHQ/useragent"
	"io"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// Default settings for headless browser operation.
const (
	DefaultTimeout    = 45 * time.Second
	DefaultWaitBuffer = 2 * time.Second
)

// WaitStrategy is a function that performs the necessary tasks to determine
// when a dynamic page has finished loading all content.
type WaitStrategy func(ctx context.Context, url string) error

// FetchRenderedContent navigates to a URL, uses the provided WaitStrategy to determine
// when dynamic content has finished loading, and extracts the content defined by
// the extractionSelector as an io.Reader.
//
// Arguments:
// - parentCtx: The context inherited from the caller.
// - url: The target URL.
// - strategy: A function encapsulating site-specific logic to pause execution.
// - extractionSelector: The CSS selector identifying the HTML node to extract (e.g., ".offers__container").
func FetchRenderedContent(parentCtx context.Context, url string, strategy WaitStrategy, extractionSelector string) (io.Reader, error) {
	ua, err := useragent.Desktop()
	if err != nil {
		return nil, fmt.Errorf("could not generate random UA: %w", err)
	}
	// 1. Setup Context with Timeout: Derive a new timed context from the parentCtx.
	ctx, cancel := context.WithTimeout(parentCtx, DefaultTimeout)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(ua), // Random User Agent (essential)
		chromedp.Headless,      // Still run headless
		chromedp.WindowSize(1920, 1080),

		// Core Evasion Flags
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),

		// Additional "Stealth" Flags:
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("ignore-certificate-errors", true), // Good for testing, but be mindful
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("no-first-run", true),

		// CRITICAL for local/Docker environments:
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("single-process", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()
	// 2. Create a Chrome instance context derived from the new timed context.
	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer chromeCancel()

	var fullHTML string

	// 3. Run the custom waiting strategy (which includes navigation)
	if err := strategy(chromeCtx, url); err != nil {
		return nil, fmt.Errorf("wait strategy failed for %s: %w", url, err)
	}

	// 4. Extract the final HTML from the specified extractionSelector
	tasks := chromedp.Tasks{
		// Wait a small buffer just to be safe after the custom wait passes
		chromedp.Sleep(DefaultWaitBuffer),

		// Use the generic extractionSelector provided by the caller
		chromedp.OuterHTML(extractionSelector, &fullHTML, chromedp.ByQuery),
	}

	if err := chromedp.Run(chromeCtx, tasks); err != nil {
		// If an error occurs, log the error and the length of the string to help diagnose truncation
		log.Printf("Extraction failed (Length: %d). Error: %v", len(fullHTML), err)
		return nil, fmt.Errorf("failed to extract HTML from selector '%s': %w", extractionSelector, err)
	}

	// 5. Convert the content to an io.Reader
	return bytes.NewReader([]byte(fullHTML)), nil
}
