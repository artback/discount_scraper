# ICA Offers Scraper (Go + headless Chrome + PostgreSQL)
A small Go service that scrapes weekly promotions from selected ICA stores, waits for dynamically rendered content, normalizes offer data, calculates discount percentages, and persists the results to PostgreSQL for analysis or downstream use.
Why this exists:
- ICA store pages render many parts of the offers client-side. Traditional HTTP fetching won’t include the final DOM.
- This project uses a headless browser to wait for the page to fully render before extracting the relevant section.
- The offers are parsed and normalized (single price, multi-buy, percentage) and stored in a relational database to enable easy querying.

## Features
- Headless, deterministic scraping of dynamic pages using Chrome DevTools (headless Chrome).
- Site-specific wait strategy to avoid race conditions with dynamic content.
- Normalization of offer text into structured fields (e.g., single price, multi-buy, percent-off).
- Discount percentage calculation based on extracted data.
- Concurrent scraping of multiple stores.
- Persistence via GORM with automatic migrations.
- Dockerized PostgreSQL for local development.

## Architecture at a glance
- Fetching: Headless browser navigates to the store’s offers page and waits until all items are rendered.
- Parsing: Raw HTML is parsed into a lightweight intermediate structure.
- Transformation: Deals are normalized and a discount percentage is computed.
- Persistence: Structured offers are upserted into PostgreSQL with migrations on startup.

## Prerequisites
- Go 1.25+
- Docker and Docker Compose (for PostgreSQL)
- Google Chrome or Chromium installed on the machine (required by the headless browser)

Tip: Ensure Chrome/Chromium is installed and accessible on your PATH. On macOS this typically works out of the box with Google Chrome.
## Quick start
1. Clone the repository
``` bash
git clone <YOUR_REPO_URL>
cd <YOUR_PROJECT_DIR>
```
1. Start PostgreSQL with Docker Compose
``` bash
docker compose up -d db
```
The database will listen on localhost:5432 and persist data in a local Docker volume.
1. Configure the application to connect to PostgreSQL

You can either use a single connection string (recommended) or discrete variables. Use placeholders instead of real secrets.
- Option A: Single DSN (recommended)
``` bash
# Example DSN. Replace placeholders with your own values.
export DB_CONN="host=localhost user=<DB_USER> password=<DB_PASSWORD> dbname=offers_db port=5432 sslmode=disable TimeZone=UTC"
```
- Option B: Discrete variables (supported if your configuration is set up for them)
``` bash
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="<DB_USER>"
export DB_PASSWORD="<DB_PASSWORD>"
export DB_NAME="offers_db"
```
If you prefer a .env file, add the same variables there (ensure your local setup loads it or export them in your shell).
1. Run the scraper
``` bash
go run .
```
What happens:
- The service connects to PostgreSQL and runs auto-migrations.
- It scrapes the configured store pages concurrently.
- Offers are inserted/updated in the database.
- A summary is printed at the end.

1. Build a binary (optional)
``` bash
go build -o bin/ica-scraper .
./bin/ica-scraper
```
## Configuration
- Database connection:
    - Prefer a single DSN in DB_CONN for simplicity (see example above).
    - For local development via the included service, use host=localhost port=5432 dbname=offers_db with your own user/password.

- Target stores:
    - The initial set of stores is defined in the application. To add or change stores, you can update the list in the app for now. A future enhancement could externalize this to configuration.

- Timeouts and waits:
    - The headless browser uses sensible defaults for navigation and extraction. If you’re scraping under slow networks or heavy load, consider increasing timeouts in the codebase.

## Working with the database
- Verify that the container is running:
``` bash
docker ps
```
- Connect with your favorite client using:
    - Host: localhost
    - Port: 5432
    - Database: offers_db
    - User: <DB_USER>
    - Password: <DB_PASSWORD>

- The application will create and migrate tables automatically on startup.

## Common tasks
- Run:
``` bash
go run .
```
- Test:
``` bash
go test ./...
```
- Lint (optional if you use golangci-lint):
``` bash
golangci-lint run
```
## Troubleshooting
- The app can’t find Chrome/Chromium:
    - Ensure Google Chrome or Chromium is installed.
    - On CI or minimal systems, install Chromium.
    - If still failing, confirm that launching Chrome manually works on the machine.

- Database connection refused:
    - Ensure Docker is running and `docker compose up -d db` succeeded.
    - Confirm credentials in your environment match the ones used by the database.
    - Verify port 5432 isn’t blocked or already in use.

- SSL errors locally:
    - Include `sslmode=disable` in your DSN for local development unless you have SSL configured.

## Contributing
Contributions are welcome! Here’s how to help:
1. Fork and create a feature branch
``` bash
git checkout -b feature/<short-description>
```
1. Set up your environment

- Start the database: `docker compose up -d db`
- Export DB variables (or use a .env)
- Run the app: `go run .`

1. Code style and quality

- Format: `gofmt -w .`
- Vet: `go vet ./...`
- Tests: `go test ./...`
- Optional: `golangci-lint run`

1. Commit and push

- Use clear, descriptive commit messages.
- Keep PRs focused and include context, screenshots (if applicable), and testing notes.

1. Open a Pull Request

- Explain the motivation, approach, and any trade-offs.
- Note any follow-ups or areas for future improvement.

Ideas to contribute:
- Add configuration for stores (file/env-driven) instead of hardcoding.
- Extend parsing coverage for more offer formats.
- Add metrics and observability.
- Containerize the app itself for end-to-end Docker runs.
- Improve error handling, retries, and backoff in scraping.

## License
Please add a license that fits your needs (e.g., MIT, Apache-2.0). If you’re unsure, start a discussion in an issue.
## Disclaimer
Use this project responsibly and in compliance with the target website’s terms of use.
