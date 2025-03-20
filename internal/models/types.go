package models

import "time"

// ReconByte represents the structure of the scraped data
type ReconByte struct {
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    string    `json:"author,omitempty"`
	Content   string    `json:"content"`
}

// ProcessedFile represents a file in the database
type ProcessedFile struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	Processed bool      `json:"processed"`
}

// Category represents a classification category
type Category struct {
	Name        string   `json:"name"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}
