package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Categorizer defines the interface for categorizing products.
type Categorizer interface {
	Categorize(ctx context.Context, products []string) (map[string][]string, error)
}

// AICategorizer implements Categorizer using Google Generative AI.
type AICategorizer struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewAICategorizer creates a new AICategorizer.
func NewAICategorizer(ctx context.Context, apiKey string) (*AICategorizer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for AICategorizer")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash-lite")
	model.ResponseMIMEType = "application/json"

	return &AICategorizer{
		client: client,
		model:  model,
	}, nil
}

// Close closes the underlying client.
func (c *AICategorizer) Close() {
	c.client.Close()
}

// Categorize categorizes a list of products into Swedish categories.
func (c *AICategorizer) Categorize(ctx context.Context, products []string) (map[string][]string, error) {
	if len(products) == 0 {
		return nil, nil
	}

	// Batch processing could be done here if the list is huge, but for now we assume
	// the caller handles reasonable batch sizes or we do simple batching.
	// Let's do simple batching of 50 products to avoid token limits.
	batchSize := 50
	allCategories := make(map[string][]string)

	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		cats, err := c.categorizeBatch(ctx, batch)
		if err != nil {
			log.Printf("Error categorizing batch %d-%d: %v", i, end, err)
			// Continue with other batches or return error?
			// Let's return partial results if possible, but for now logging is safer.
			continue
		}

		for k, v := range cats {
			allCategories[k] = v
		}
	}

	return allCategories, nil
}

func (c *AICategorizer) categorizeBatch(ctx context.Context, products []string) (map[string][]string, error) {
	prompt := fmt.Sprintf(`You are a grocery product categorizer for a Swedish store.
Categorize the following products into standard Swedish grocery categories (e.g., Frukt & Grönt, Mejeri, Kött, Chark, Skafferi, Dryck, Bröd & Kakor, Frys, Hem & Hushåll, Hälsa & Skönhet, Barn, Husdjur).
A product can belong to multiple categories.
Return ONLY a JSON object where keys are product names and values are arrays of category strings.
Do not include markdown formatting like `+"```json"+`.

Products:
%s`, strings.Join(products, "\n"))

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no content generated")
	}

	var result map[string][]string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			// Clean up potential markdown code blocks if the model ignores instructions
			cleanTxt := strings.TrimSpace(string(txt))
			cleanTxt = strings.TrimPrefix(cleanTxt, "```json")
			cleanTxt = strings.TrimPrefix(cleanTxt, "```")
			cleanTxt = strings.TrimSuffix(cleanTxt, "```")

			if err := json.Unmarshal([]byte(cleanTxt), &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON response: %w. Response: %s", err, cleanTxt)
			}
			break // Only process the first text part
		}
	}

	return result, nil
}
