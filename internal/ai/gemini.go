package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dylanmurzello/recon_byte_generator/internal/models"
	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Processor represents an AI processor
type Processor struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewProcessor creates a new AI processor
func NewProcessor(apiKey string) (*Processor, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.0-pro")

	return &Processor{
		client: client,
		model:  model,
	}, nil
}

// Close closes the AI client
func (p *Processor) Close() error {
	p.client.Close()
	return nil
}

// Process processes the given ReconByte with AI
func (p *Processor) Process(ctx context.Context, data *models.ReconByte, promptPath, categoriesPath string) (string, error) {
	// Read categories
	categoriesBytes, err := os.ReadFile(categoriesPath)
	if err != nil {
		return "", fmt.Errorf("failed to read categories file: %w", err)
	}

	var categories struct {
		Categories []struct {
			Name          string   `json:"name"`
			Subcategories []string `json:"subcategories"`
		} `json:"categories"`
	}

	if err := json.Unmarshal(categoriesBytes, &categories); err != nil {
		return "", fmt.Errorf("failed to parse categories: %w", err)
	}

	// Construct a more focused prompt
	prompt := fmt.Sprintf(`Analyze the following news article and create a Recon Byte report. Focus on:

1. Title: Create a clear, concise title for this event
2. Description: Summarize the key points (who, what, when, where, why, how)
3. Threat Assessment:
   - Select the most appropriate category from: %s
   - Determine severity (None, Low, Medium, High, Critical)
   - Assess priority (Low, Medium, High)
   - Evaluate confidence level (None, Low, Medium, High, Critical)

4. Impact Analysis:
   - Estimate initial impact (0.1 to 1.0)
   - Identify affected areas/populations
   - Project potential escalation

Article Content:
%s

Format your response as a structured report with clear sections and justifications for your assessments.`,
		strings.Join(getCategoryNames(categories.Categories), ", "),
		data.Content)

	// Generate content with safety settings
	safetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
	}

	p.model.SafetySettings = safetySettings

	// Generate content
	resp, err := p.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate AI content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates received")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(*genai.Text); ok {
			result.WriteString(string(*text))
			result.WriteString("\n")
		}
	}

	if result.Len() == 0 {
		return "", fmt.Errorf("empty response from AI")
	}

	return result.String(), nil
}

// getCategoryNames extracts just the category names from the categories structure
func getCategoryNames(categories []struct {
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories"`
}) []string {
	names := make([]string, len(categories))
	for i, cat := range categories {
		names[i] = cat.Name
	}
	return names
}
