package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dylanmurzello/recon_byte_generator/internal/models"
	_ "modernc.org/sqlite"
)

// Client represents a database client
type Client struct {
	db *sql.DB
}

// NewClient creates a new database client
func NewClient(dbPath string) (*Client, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Client{db: db}, nil
}

// initSchema creates the necessary tables if they don't exist
func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed INTEGER DEFAULT 0
	);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// InsertFile logs a new file in the database
func (c *Client) InsertFile(filename string) error {
	_, err := c.db.Exec("INSERT INTO files (filename, created_at) VALUES (?, ?)",
		filename, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert file: %w", err)
	}
	return nil
}

// MarkProcessed marks a file as processed
func (c *Client) MarkProcessed(filename string) error {
	result, err := c.db.Exec("UPDATE files SET processed = 1 WHERE filename = ?", filename)
	if err != nil {
		return fmt.Errorf("failed to mark file as processed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no file found with filename: %s", filename)
	}

	return nil
}

// GetUnprocessedFile retrieves the next unprocessed file
func (c *Client) GetUnprocessedFile() (*models.ProcessedFile, error) {
	var file models.ProcessedFile
	err := c.db.QueryRow(`
		SELECT id, filename, created_at, processed 
		FROM files 
		WHERE processed = 0 
		ORDER BY created_at 
		LIMIT 1
	`).Scan(&file.ID, &file.Filename, &file.CreatedAt, &file.Processed)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed file: %w", err)
	}

	return &file, nil
}
