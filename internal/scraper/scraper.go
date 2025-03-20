package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dylanmurzello/recon_byte_generator/internal/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Scraper represents a web scraper
type Scraper struct {
	browser *rod.Browser
}

// NewScraper creates a new scraper instance
func NewScraper() (*Scraper, error) {
	browser := rod.New().ControlURL(launcher.New().
		Headless(true).
		NoSandbox(true).
		MustLaunch(),
	).MustConnect()

	return &Scraper{
		browser: browser,
	}, nil
}

// Close closes the browser
func (s *Scraper) Close() error {
	return s.browser.Close()
}

// extractText extracts text using a regex pattern
func extractText(pageHTML, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(pageHTML)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// Scrape scrapes the given URL and returns the extracted data
func (s *Scraper) Scrape(ctx context.Context, url string) (*models.ReconByte, error) {
	// Create a new page
	page := s.browser.MustPage(url)
	defer page.Close()

	// Wait for page to load with context
	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to load page: %w", err)
	}

	// Extract the full HTML source
	pageHTML, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("failed to get page HTML: %w", err)
	}

	// Extract article content (all paragraph text)
	texts, err := page.Elements("p")
	if err != nil {
		return nil, fmt.Errorf("failed to find paragraph elements: %w", err)
	}

	var content []string
	for _, t := range texts {
		text, err := t.Text()
		if err != nil {
			continue
		}
		content = append(content, text)
	}

	// Extract author name
	authorPattern := `"@type":"Person","name":"(.*?)"`
	author := extractText(pageHTML, authorPattern)

	// Create ReconByte
	reconData := &models.ReconByte{
		Timestamp: time.Now(),
		URL:       url,
		Author:    author,
		Content:   strings.Join(content, "\n"),
	}

	return reconData, nil
}
