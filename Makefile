# AniFlux Worker & API Makefile

all: build-worker build-migrate test

# Pre-build setup step (DB migrations / seeding)
prebuild:
	@echo "Running pre-build tasks & environment check..."
	@go mod download

# Build the AniFlux Worker binary
build-worker:
	@echo "Building AniFlux Worker binary..."
	@go build -o bin/worker cmd/worker/main.go

# Build the database migration tool
build-migrate:
	@echo "Building AniFlux Database Migration binary..."
	@go build -o bin/migrate cmd/migrate/main.go

# Build the API binary
build-api:
	@echo "Building AniFlux API binary..."
	@go build -o bin/api cmd/api/main.go

# Run the compiled worker executable directly (accepting JSON payload on Stdin)
run-worker:
	@./bin/worker

# Run the database migrations
migrate:
	@go run cmd/migrate/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean binaries
clean:
	@echo "Cleaning..."
	@rm -rf bin/ main

.PHONY: all prebuild build-worker build-migrate build-api run-worker migrate test clean
