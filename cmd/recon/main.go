package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/dylanmurzello/recon_byte_generator/config"
	"github.com/dylanmurzello/recon_byte_generator/internal/ai"
	"github.com/dylanmurzello/recon_byte_generator/internal/database"
	"github.com/dylanmurzello/recon_byte_generator/internal/scraper"
)

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal. Shutting down...")
		cancel()
	}()

	// Initialize database
	db, err := database.NewClient(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize scraper
	s, err := scraper.NewScraper()
	if err != nil {
		log.Fatalf("Failed to initialize scraper: %v", err)
	}
	defer s.Close()

	// Initialize AI processor
	processor, err := ai.NewProcessor(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize AI processor: %v", err)
	}
	defer processor.Close()

	// Step 1: Initialize Gemini with prompt and categories
	log.Printf("Initializing Gemini AI with prompt and categories...")
	if err := processor.InitializeGemini(ctx, cfg.PromptPath, cfg.CategoriesPath); err != nil {
		log.Fatalf("Failed to initialize Gemini: %v", err)
	}
	log.Printf("✅ Gemini AI initialized successfully!")

	// Get URL from user
	fmt.Print("Enter the news URL: ")
	var url string
	fmt.Scanln(&url)

	// Scrape the URL
	log.Printf("Scraping URL: %s", url)
	reconData, err := s.Scrape(ctx, url)
	if err != nil {
		log.Fatalf("Failed to scrape URL: %v", err)
	}

	// Save scraped data
	filename := fmt.Sprintf("%d.json", time.Now().Unix())
	filePath := filepath.Join(cfg.OutputDir, filename)

	// Convert reconData to JSON and save it
	reconJSON, err := json.MarshalIndent(reconData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal scraped data: %v", err)
	}
	if err := os.WriteFile(filePath, reconJSON, 0644); err != nil {
		log.Fatalf("Failed to save scraped data: %v", err)
	}

	// Step 2: Process with AI
	log.Printf("Processing article with Gemini AI...")
	aiResponse, err := processor.Process(ctx, reconData, cfg.PromptPath, cfg.CategoriesPath)
	if err != nil {
		log.Fatalf("Failed to process with AI: %v", err)
	}

	// Save AI response
	aiFilePath := filepath.Join(cfg.OutputDir, "gemini_response_"+filename)
	if err := os.WriteFile(aiFilePath, []byte(aiResponse), 0644); err != nil {
		log.Fatalf("Failed to save AI response: %v", err)
	}

	// Update database
	if err := db.InsertFile(filename); err != nil {
		log.Printf("Warning: Failed to log file in database: %v", err)
	}

	if err := db.MarkProcessed(filename); err != nil {
		log.Printf("Warning: Failed to mark file as processed: %v", err)
	}

	log.Printf("✅ Processing completed successfully!")
	log.Printf("Original data saved to: %s", filePath)
	log.Printf("AI response saved to: %s", aiFilePath)
}
