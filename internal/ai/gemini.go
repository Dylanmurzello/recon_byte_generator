package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dylanmurzello/recon_byte_generator/internal/models"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// Processor represents an AI processor
type Processor struct {
	apiKey string
	client *http.Client
}

// Categories represents the structure of the categories JSON file
type Categories struct {
	Categories []struct {
		Name          string   `json:"name"`
		Subcategories []string `json:"subcategories"`
	} `json:"categories"`
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// NewProcessor creates a new AI processor
func NewProcessor(apiKey string) (*Processor, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	return &Processor{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

// Close closes the AI client
func (p *Processor) Close() error {
	return nil
}

// formatCategories formats the categories into a readable string
func formatCategories(cats Categories) string {
	var result strings.Builder
	result.WriteString("Available Categories:\n")

	for _, cat := range cats.Categories {
		result.WriteString(fmt.Sprintf("\n%s:\n", cat.Name))
		for _, sub := range cat.Subcategories {
			result.WriteString(fmt.Sprintf("  - %s\n", sub))
		}
	}

	return result.String()
}

// Process processes the given ReconByte with AI
func (p *Processor) Process(ctx context.Context, data *models.ReconByte, promptPath, categoriesPath string) (string, error) {
	if data == nil || data.Content == "" {
		return "", fmt.Errorf("invalid input: empty content")
	}

	// Read prompt template
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template: %w", err)
	}

	// Read and parse categories
	categoriesData, err := os.ReadFile(categoriesPath)
	if err != nil {
		return "", fmt.Errorf("failed to read categories: %w", err)
	}

	var categories Categories
	if err := json.Unmarshal(categoriesData, &categories); err != nil {
		return "", fmt.Errorf("failed to parse categories: %w", err)
	}

	// Format categories into a readable string
	formattedCategories := formatCategories(categories)

	// Format the prompt with the article data and categories
	prompt := fmt.Sprintf(string(promptTemplate),
		data.URL,
		data.Author,
		data.Timestamp,
		data.Content,
		formattedCategories)

	// Prepare request body
	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.7,
			TopK:            40,
			TopP:            0.8,
			MaxOutputTokens: 2048,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// InitializeGemini sends the prompt template and categories to Gemini and waits for acknowledgment
func (p *Processor) InitializeGemini(ctx context.Context, promptPath, categoriesPath string) error {
	// Read prompt template
	promptTemplate, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("failed to read prompt template: %w", err)
	}

	// Read and parse categories
	categoriesData, err := os.ReadFile(categoriesPath)
	if err != nil {
		return fmt.Errorf("failed to read categories: %w", err)
	}

	var categories Categories
	if err := json.Unmarshal(categoriesData, &categories); err != nil {
		return fmt.Errorf("failed to parse categories: %w", err)
	}

	// Format categories into a readable string
	formattedCategories := formatCategories(categories)

	// Create initialization message
	initMessage := fmt.Sprintf("Instructions for generating Recon Bytes:\n\n%s\n\nAvailable Categories:\n%s\n\nPlease acknowledge if you understand these instructions.", string(promptTemplate), formattedCategories)

	// Prepare request body
	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: initMessage},
				},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:     0.7,
			TopK:            40,
			TopP:            0.8,
			MaxOutputTokens: 2048,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("no response generated")
	}

	// Check for acknowledgment
	response := geminiResp.Candidates[0].Content.Parts[0].Text
	if !strings.Contains(strings.ToLower(response), "acknowledge") {
		return fmt.Errorf("Gemini did not acknowledge the instructions properly. Response: %s", response)
	}

	return nil
}
