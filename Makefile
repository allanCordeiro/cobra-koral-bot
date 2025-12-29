.PHONY: help build test run clean docker-build docker-run deploy

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the worker binary
	@echo "Building worker..."
	@go build -o bin/worker ./cmd/worker
	@echo "Build complete: bin/worker"

test: ## Run all tests
	@echo "Running tests..."
	@go test ./... -v

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run: ## Run the worker locally
	@echo "Running worker..."
	@mkdir -p data
	@go run ./cmd/worker

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t weather-worker:latest .
	@echo "Docker image built: weather-worker:latest"

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --env-file .env weather-worker:latest

lint: ## Run Go linter
	@echo "Running linter..."
	@go vet ./...
	@echo "Lint complete"

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

mod-tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Modules tidied"

# Google Cloud Run deployment targets
GCP_PROJECT ?= your-project-id
GCP_REGION ?= us-central1
IMAGE_NAME = gcr.io/$(GCP_PROJECT)/weather-worker

gcp-build: ## Build and push to Google Container Registry
	@echo "Building and pushing to GCR..."
	@gcloud builds submit --tag $(IMAGE_NAME)

gcp-deploy-job: ## Deploy to Cloud Run Jobs
	@echo "Deploying to Cloud Run Jobs..."
	@gcloud run jobs create weather-worker \
		--image $(IMAGE_NAME) \
		--set-env-vars BOT_TOKEN=$(BOT_TOKEN),USER_ID=$(USER_ID) \
		--region $(GCP_REGION) \
		--max-retries 1

gcp-schedule: ## Create Cloud Scheduler job
	@echo "Creating Cloud Scheduler job..."
	@gcloud scheduler jobs create http weather-check \
		--location $(GCP_REGION) \
		--schedule="*/30 * * * *" \
		--uri="https://$(GCP_REGION)-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/$(GCP_PROJECT)/jobs/weather-worker:run" \
		--http-method POST \
		--oauth-service-account-email $(GCP_PROJECT)@appspot.gserviceaccount.com

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Created .env file - please edit with your credentials"
	@mkdir -p data
	@echo "Created data directory for state persistence"
	@go mod download
	@echo "Setup complete"

.DEFAULT_GOAL := help
