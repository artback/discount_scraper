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
- A simple API to serve the scraped offers.

## Architecture at a glance

- **Fetching**: Headless browser navigates to the store’s offers page and waits until all items are rendered.
- **Parsing**: Raw HTML is parsed into a lightweight intermediate structure.
- **Transformation**: Deals are normalized and a discount percentage is computed.
- **Persistence**: Structured offers are upserted into PostgreSQL with migrations on startup.
- **API**: A simple HTTP server to expose the scraped offers as JSON.

## Prerequisites

- Go 1.25.3+
- Docker and Docker Compose (for PostgreSQL)
- Google Chrome or Chromium installed on the machine (required by the headless browser)

Tip: Ensure Chrome/Chromium is installed and accessible on your PATH. On macOS this typically works out of the box with Google Chrome.

## Quick start

1.  Clone the repository

    ```bash
    git clone <YOUR_REPO_URL>
    cd <YOUR_PROJECT_DIR>
    ```

2.  Start PostgreSQL with Docker Compose

    ```bash
    docker compose up -d db
    ```

    The database will listen on localhost:5432 and persist data in a local Docker volume.

3.  Configure the application

    Copy the `config.yaml.example` to `config.yaml` and update the database credentials and store list.

    ```bash
    cp config.yaml.example config.yaml
    ```

    Update `config.yaml` with your database credentials.

4.  Run the scraper

    ```bash
    go run ./cmd/parser
    ```

    What happens:

    - The service connects to PostgreSQL and runs auto-migrations.
    - It scrapes the configured store pages concurrently.
    - Offers are inserted/updated in the database.
    - A summary is printed at the end.

5.  Run the API server

    ```bash
    go run ./cmd/api
    ```

    The API server will be available at `http://localhost:8080`.

## Usage

### Parser

To run the scraper, use the following command:

```bash
go run ./cmd/parser
```

### API

To run the API server, use the following command:

```bash
go run ./cmd/api
```

The API has the following endpoints:

- `GET /`: Serves the main page.
- `GET /api/offers`: Serves the scraped offers as JSON.

## Configuration

-   **Database connection**:
    -   The database connection is configured in `config.yaml`.
-   **Target stores**:
    -   The list of stores to scrape is defined in `config.yaml`.
-   **Timeouts and waits**:
    -   The headless browser uses sensible defaults for navigation and extraction. If you’re scraping under slow networks or heavy load, consider increasing timeouts in the codebase.

## Working with the database

-   Verify that the container is running:

    ```bash
    docker ps
    ```

-   Connect with your favorite client using:
    -   Host: localhost
    -   Port: 5432
    -   Database: offers_db
    -   User: <DB_USER>
    -   Password: <DB_PASSWORD>
-   The application will create and migrate tables automatically on startup.

## Common tasks

-   Run the scraper:

    ```bash
    go run ./cmd/parser
    ```

-   Run the API server:

    ```bash
    go run ./cmd/api
    ```

-   Test:

    ```bash
    go test ./...
    ```

-   Lint (optional if you use golangci-lint):

    ```bash
    golangci-lint run
    ```

## Troubleshooting

-   **The app can’t find Chrome/Chromium**:
    -   Ensure Google Chrome or Chromium is installed.
    -   On CI or minimal systems, install Chromium.
    -   If still failing, confirm that launching Chrome manually works on the machine.
-   **Database connection refused**:
    -   Ensure Docker is running and `docker compose up -d db` succeeded.
    -   Confirm credentials in your environment match the ones used by the database.
    -   Verify port 5432 isn’t blocked or already in use.
-   **SSL errors locally**:
    -   Include `sslmode=disable` in your DSN for local development unless you have SSL configured.

## Contributing

Contributions are welcome! Here’s how to help:

1.  Fork and create a feature branch

    ```bash
    git checkout -b feature/<short-description>
    ```

2.  Set up your environment

    -   Start the database: `docker compose up -d db`
    -   Copy `config.yaml.example` to `config.yaml` and update the database credentials.
    -   Run the app: `go run ./cmd/parser`

3.  Code style and quality

    -   Format: `gofmt -w .`
    -   Vet: `go vet ./...`
    -   Tests: `go test ./...`
    -   Optional: `golangci-lint run`

4.  Commit and push

    -   Use clear, descriptive commit messages.
    -   Keep PRs focused and include context, screenshots (if applicable), and testing notes.

5.  Open a Pull Request

    -   Explain the motivation, approach, and any trade-offs.
    -   Note any follow-ups or areas for future improvement.

### Ideas to contribute

-   Add configuration for stores (file/env-driven) instead of hardcoding.
-   Extend parsing coverage for more offer formats.
-   Add metrics and observability.
-   Containerize the app itself for end-to-end Docker runs.
-   Improve error handling, retries, and backoff in scraping.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

Use this project responsibly and in compliance with the target website’s terms of use.