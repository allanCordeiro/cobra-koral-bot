# Weather Alert Worker - CGE São Paulo

A Go worker application that monitors weather forecasts from CGE São Paulo and sends alerts via Telegram when specific conditions are met.

## Architecture

This application follows the **Hexagonal Architecture** (Ports and Adapters) pattern:

```
.
├── cmd/worker/              # Application entry point
├── internal/
│   ├── domain/             # Domain entities and interfaces (ports)
│   └── usecases/           # Business logic
└── infra/                  # Infrastructure implementations (adapters)
    ├── logger/             # JSON structured logging
    ├── storage/            # State persistence
    ├── scraper/            # Web scraping
    └── telegram/           # Telegram bot integration
```

## Features

- **Web Scraping**: Monitors CGE São Paulo weather forecast page
- **Keyword Detection**: Analyzes news for weather alert keywords (chuva, alagamento, temporal, etc.)
- **Flooding Information**: Fetches and includes active flooding points
- **Telegram Notifications**: Sends alerts with weather map images
- **State Management**: Tracks last execution to avoid duplicate alerts
- **Structured Logging**: JSON-formatted logs for easy parsing

## Requirements

- Go 1.24+
- Telegram Bot Token
- Telegram User ID

## Environment Variables

```bash
BOT_TOKEN=your_telegram_bot_token
USER_ID=your_telegram_user_id
STATE_FILE_PATH=/app/data/state.json  # Optional
```

## Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd cobra-coral
   ```

2. **Copy environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your credentials
   ```

3. **Create data directory** (for state persistence)
   ```bash
   mkdir -p data
   ```

4. **Run tests**
   ```bash
   go test ./... -v
   ```

5. **Build the application**
   ```bash
   go build -o bin/worker ./cmd/worker
   ```

6. **Run locally**
   ```bash
   ./bin/worker
   ```

**Note**: The application will create `./data/state.json` to track execution state. This directory is ignored by git.

## Docker Build

```bash
docker build -t weather-worker .
docker run --env-file .env weather-worker
```

## Google Cloud Run Deployment

### Using Cloud Run Jobs (Recommended)

1. **Build and push to Container Registry**
   ```bash
   gcloud builds submit --tag gcr.io/PROJECT_ID/weather-worker
   ```

2. **Create Cloud Run Job**
   ```bash
   gcloud run jobs create weather-worker \
     --image gcr.io/PROJECT_ID/weather-worker \
     --set-env-vars BOT_TOKEN=your_token,USER_ID=your_id \
     --region us-central1
   ```

3. **Schedule with Cloud Scheduler**
   ```bash
   gcloud scheduler jobs create http weather-check \
     --location us-central1 \
     --schedule="*/30 * * * *" \
     --uri="https://us-central1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/PROJECT_ID/jobs/weather-worker:run" \
     --http-method POST \
     --oauth-service-account-email PROJECT_ID@appspot.gserviceaccount.com
   ```

### State Persistence Options

1. **Volume Mount** (for Cloud Run services)
   - Mount a GCS bucket using FUSE

2. **Firestore/Cloud SQL** (recommended for production)
   - Implement alternative `StateRepository` using cloud database

3. **File-based** (current implementation)
   - Works with persistent volumes
   - Simple for development and testing

## How It Works

1. **Fetch Latest News**: Scrapes CGE São Paulo homepage for the most recent weather forecast
   - Searches for links containing `noticias.jsp?id=` (actual news articles)
   - Extracts title from `<h1>` tag
   - Extracts datetime from `<h2>` tag (format: DD/MM/YYYY HH:MM)
2. **Check Timestamp**: Compares news publication time with last execution
3. **Keyword Analysis**: Searches for alert keywords in Portuguese
4. **Fetch Flooding Data**: If active flooding points exist, fetches details
   - Scrapes "Pontos de Alagamento: X ativos" from the page
   - If count > 0, fetches detailed flooding information by zone
5. **Download Weather Map**: Gets the current weather map image
6. **Send Telegram Alert**: Sends formatted message with image to configured user
7. **Update State**: Saves current execution time for next run

### Scraping Details

**News Structure** (from https://www.cgesp.org/v3/index.jsp):
```html
<a href="noticias.jsp?id=53676">
  <h1>Title of the news</h1>
  <h2>28/12/2025 21:00 - Domingo</h2>
</a>
```

**Flooding Info**:
```html
Pontos de Alagamento: 0 ativos
```

## Alert Keywords

The following Portuguese keywords trigger alerts:
- atenção
- alerta
- chuva
- alagamento
- temporal
- enchente
- inundação

## Testing

Run all unit tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

## Project Structure Details

### Domain Layer (`internal/domain/`)
- Pure business entities and interfaces
- No external dependencies
- Defines contracts for all infrastructure

### Use Cases Layer (`internal/usecases/`)
- Application business logic
- Orchestrates domain entities
- Independent of infrastructure details

### Infrastructure Layer (`infra/`)
- Concrete implementations of domain interfaces
- External dependencies (HTTP, Telegram API, file I/O)
- Can be swapped without affecting business logic

## Logging

All operations are logged in structured JSON format:

```json
{
  "type": "INFO",
  "message": "Successfully fetched news",
  "domain": "CGESPScraper",
  "timestamp": "2025-12-28T21:00:00Z"
}
```

Log types: `INFO`, `WARN`, `ERROR`

## Error Handling

- Scraping structure errors trigger a Telegram alert
- Failed image downloads don't block notifications
- Failed flooding data doesn't prevent alerts
- All errors are logged with context

## Troubleshooting

### Permission Denied Error on Startup

If you see `mkdir /app: permission denied` when running locally:

1. **Check your .env file** - Make sure `STATE_FILE_PATH` is either:
   - Commented out (to use default `./data/state.json`)
   - Set to a relative path: `STATE_FILE_PATH=./data/state.json`

2. **Create data directory**:
   ```bash
   mkdir -p data
   ```

3. **Use Makefile helper**:
   ```bash
   make dev-setup
   ```

### Tests Failing

Run tests with verbose output:
```bash
go test ./... -v
```

Clean and rebuild:
```bash
make clean
make build
make test
```

## License

MIT

## Contributing

Contributions are welcome! Please ensure:
- Code follows hexagonal architecture principles
- All tests pass
- New features include unit tests
- Variables and functions use clear English names
- Log messages use English (except scraped content)
