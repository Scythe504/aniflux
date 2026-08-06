# AniFlux Worker & API Makefile

all: build-worker test

# Pre-build setup step (DB migrations / seeding)
prebuild:
	@echo "Running pre-build tasks & environment check..."
	@go mod download

# Build the AniFlux Worker binary
build-worker:
	@echo "Building AniFlux Worker binary..."
	@go build -o bin/worker cmd/worker/main.go

# Build the API binary
build-api:
	@echo "Building AniFlux API binary..."
	@go build -o bin/api cmd/api/main.go

# Run the compiled worker executable directly (accepting JSON payload on Stdin)
run-worker:
	@./bin/worker

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean binaries
clean:
	@echo "Cleaning..."
	@rm -rf bin/ main

.PHONY: all prebuild build-worker build-api run-worker test clean
