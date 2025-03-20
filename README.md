# Recon Byte Generator

A tool for scraping news articles and processing them with Google's Gemini AI.

## Features

- Web scraping using Rod browser automation
- AI processing with Google's Gemini AI
- SQLite database for tracking processed files
- Configurable through environment variables
- Graceful shutdown handling
- Structured logging

## Prerequisites

- Go 1.21 or later
- Google Gemini API key
- Chrome/Chromium browser (for web scraping)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/dylanmurzello/recon_byte_generator.git
cd recon_byte_generator
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file with your Gemini API key:
```bash
GEMINI_API_KEY=your_api_key_here
```

## Configuration

The following environment variables can be set:

- `GEMINI_API_KEY` (required): Your Google Gemini API key
- `DB_PATH` (optional): Path to SQLite database (default: "files.db")
- `PROMPT_PATH` (optional): Path to prompt template (default: "prompt.txt")
- `CATEGORIES_PATH` (optional): Path to categories file (default: "Categories.json")
- `OUTPUT_DIR` (optional): Directory for output files (default: "recon_bytes")

## Usage

1. Run the application:
```bash
go run cmd/recon/main.go
```

2. Enter a news URL when prompted.

3. The application will:
   - Scrape the article
   - Process it with Gemini AI
   - Save both the original data and AI response
   - Track processing in the database

## Project Structure

```
recon_byte_generator/
├── cmd/
│   └── recon/
│       └── main.go           # Application entry point
├── config/
│   └── config.go            # Configuration management
├── internal/
│   ├── ai/
│   │   └── gemini.go        # AI processing
│   ├── database/
│   │   └── db.go           # Database operations
│   ├── models/
│   │   └── types.go        # Data structures
│   └── scraper/
│       └── scraper.go      # Web scraping
├── recon_bytes/            # Output directory
├── .env                    # Environment variables
├── Categories.json         # Classification categories
├── go.mod                 # Go module file
├── prompt.txt             # AI prompt template
└── README.md             # This file
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 