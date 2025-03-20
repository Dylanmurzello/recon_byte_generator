package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	GeminiAPIKey   string `json:"gemini_api_key"`
	DBPath         string `json:"db_path"`
	PromptPath     string `json:"prompt_path"`
	CategoriesPath string `json:"categories_path"`
	OutputDir      string `json:"output_dir"`
}

// LoadConfig loads configuration from environment and files
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	config := &Config{
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		DBPath:         getEnvWithDefault("DB_PATH", "files.db"),
		PromptPath:     getEnvWithDefault("PROMPT_PATH", "prompt.txt"),
		CategoriesPath: getEnvWithDefault("CATEGORIES_PATH", "Categories.json"),
		OutputDir:      getEnvWithDefault("OUTPUT_DIR", "recon_bytes"),
	}

	// Validate required fields
	if config.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return config, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
