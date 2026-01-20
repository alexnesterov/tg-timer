.PHONY: build run clean test webhook webhook-local webhook-build

# Build the bot
build:
	@echo "Building telegram timer bot..."
	@mkdir -p bin
	@go build -o bin/tg-timer ./cmd/tg-timer
	@echo "Binary created: bin/tg-timer"

# Build webhook version
webhook-build:
	@echo "Building telegram timer bot (webhook)..."
	@mkdir -p bin
	@go build -o bin/tg-timer-webhook ./cmd/tg-timer-webhook
	@echo "Webhook binary created: bin/tg-timer-webhook"

# Build both versions
build-all:
	@echo "Building both versions..."
	@make build
	@make webhook-build
	@echo "All binaries created in bin/"

# Run the bot (long polling)
run:
	@echo "Starting telegram timer bot (long polling)..."
	@go run ./cmd/tg-timer

# Run webhook version
webhook:
	@echo "Starting telegram timer bot (webhook)..."
	@go run ./cmd/tg-timer-webhook

# Run webhook with ngrok for local development
webhook-local:
	@echo "Starting webhook with ngrok..."
	@if [ -f .env.webhook ]; then \
		export $$(cat .env.webhook | grep -v '^#' | xargs); \
		echo "Starting webhook server on port 8443..."; \
		echo "Make sure ngrok is running: ngrok http 8443"; \
		echo "Update WEBHOOK_URL with ngrok HTTPS URL"; \
		go run ./cmd/tg-timer-webhook; \
	else \
		echo "Please copy .env.webhook.example to .env.webhook and configure it"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Development build with debug info
dev:
	@echo "Building development version..."
	@mkdir -p bin
	@go build -gcflags="all=-N -l" -o bin/tg-timer-dev ./cmd/tg-timer
	@echo "Dev binary created: bin/tg-timer-dev"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Docker commands
docker-build:
	@echo "Building Docker image (webhook)..."
	@docker build -f Dockerfile.webhook -t tg-timer-webhook .

docker-run:
	@echo "Running Docker container..."
	@docker-compose up -d

docker-stop:
	@echo "Stopping Docker containers..."
	@docker-compose down
