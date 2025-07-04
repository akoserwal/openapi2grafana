# Makefile for Grafana Dashboard Generator

# Variables
BINARY_NAME=openapi2grafana
OPENAPI_FILE=openapi.yaml
DASHBOARD_FILE=output_dashboard.json
DOCKER_COMPOSE_FILE=docker-compose.yaml

# Default target
.PHONY: all
all: build test

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go
	@echo "Build completed!"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Generate dashboard
.PHONY: generate
generate: build
	@echo "Generating dashboard..."
	./$(BINARY_NAME) $(OPENAPI_FILE) $(DASHBOARD_FILE)

# Generate dashboard from sample API spec
.PHONY: generate-sample
generate-sample: build
	@echo "Generating dashboard from sample API spec..."
	./$(BINARY_NAME) sample-openapi.yaml $(DASHBOARD_FILE) --title "Sample API Dashboard"

# Update existing dashboard
.PHONY: update
update: build
	@echo "Updating dashboard..."
	./$(BINARY_NAME) $(OPENAPI_FILE) $(DASHBOARD_FILE) --update

# Build sample API
.PHONY: build-sample-api
build-sample-api:
	@echo "Building sample API..."
	cd sample-api && go build -o sample-api .

# Run sample API locally
.PHONY: run-sample-api
run-sample-api: build-sample-api
	@echo "Running sample API on :8080..."
	cd sample-api && ./sample-api

# Start monitoring stack
.PHONY: start
start:
	@echo "Starting monitoring stack..."
	docker-compose up -d

# Start with build (rebuilds sample API)
.PHONY: start-fresh
start-fresh:
	@echo "Starting monitoring stack with fresh build..."
	docker-compose up -d --build

# Stop monitoring stack
.PHONY: stop
stop:
	@echo "Stopping monitoring stack..."
	docker-compose down

# Full test with docker
.PHONY: test-full
test-full:
	@echo "Running full test..."
	chmod +x test-dashboard.sh
	./test-dashboard.sh

# Show logs
.PHONY: logs
logs:
	@echo "Showing logs..."
	docker-compose logs -f

# Clean up
.PHONY: clean
clean:
	@echo "Cleaning up..."
	docker-compose down --remove-orphans -v
	rm -f $(BINARY_NAME)
	rm -f $(DASHBOARD_FILE)
	rm -rf backups/

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go mod download
	mkdir -p backups
	chmod +x test-dashboard.sh

# Validate dashboard
.PHONY: validate
validate:
	@echo "Validating dashboard..."
	@if [ -f $(DASHBOARD_FILE) ]; then \
		jq empty $(DASHBOARD_FILE) && echo "✓ Dashboard JSON is valid" || echo "✗ Dashboard JSON is invalid"; \
		jq -e '.title' $(DASHBOARD_FILE) > /dev/null && echo "✓ Dashboard has title" || echo "✗ Dashboard missing title"; \
		jq -e '.panels | length > 0' $(DASHBOARD_FILE) > /dev/null && echo "✓ Dashboard has panels" || echo "✗ Dashboard has no panels"; \
		echo "Panel count: $$(jq '.panels | length' $(DASHBOARD_FILE))"; \
	else \
		echo "✗ Dashboard file not found"; \
	fi

# Check prerequisites
.PHONY: check
check:
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "✗ Go is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "✗ Docker is not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "✗ Docker Compose is not installed"; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "✗ jq is not installed"; exit 1; }
	@echo "✓ All prerequisites met!"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  test             - Run Go tests"
	@echo "  generate         - Generate dashboard from OpenAPI spec"
	@echo "  generate-sample  - Generate dashboard from sample API spec"
	@echo "  update           - Update existing dashboard"
	@echo "  build-sample-api - Build sample API locally"
	@echo "  run-sample-api   - Run sample API locally on :8080"
	@echo "  start            - Start monitoring stack"
	@echo "  start-fresh      - Start monitoring stack with fresh build"
	@echo "  stop             - Stop monitoring stack"
	@echo "  test-full        - Run full integration test"
	@echo "  logs             - Show container logs"
	@echo "  clean            - Clean up all resources"
	@echo "  dev-setup        - Set up development environment"
	@echo "  validate         - Validate generated dashboard"
	@echo "  check            - Check prerequisites"
	@echo "  demo             - Run complete demo"
	@echo "  help             - Show this help message"

# Run complete demo
.PHONY: demo
demo:
	@echo "Running complete demo..."
	./demo.sh

# Monitor and regenerate on changes
.PHONY: watch
watch:
	@echo "Watching for changes..."
	@while true; do \
		if [ $(OPENAPI_FILE) -nt $(DASHBOARD_FILE) ]; then \
			echo "OpenAPI file changed, regenerating dashboard..."; \
			make generate; \
		fi; \
		sleep 2; \
	done 