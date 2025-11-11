.PHONY: build install clean test deps fmt lint gen-sdk

VERSION ?= 0.1.0
PROVIDER_NAME := web-components
PROVIDER_BINARY := pulumi-resource-$(PROVIDER_NAME)

# Build the provider binary
build:
	@echo "Building provider..."
	 go build -o ./bin/$(PROVIDER_BINARY) .

# Install the provider locally for testing
install: build
	@echo "Installing provider..."
	mkdir -p ~/.pulumi/plugins/resource-$(PROVIDER_NAME)-v$(VERSION)
	cp bin/$(PROVIDER_BINARY) ~/.pulumi/plugins/resource-$(PROVIDER_NAME)-v$(VERSION)/
	cp PulumiPlugin.yaml ~/.pulumi/plugins/resource-$(PROVIDER_NAME)-v$(VERSION)/

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf sdk/
	rm -rf ~/.pulumi/plugins/resource-$(PROVIDER_NAME)-v$(VERSION)

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

gen-sdk: build
	@echo "Generating SDK..."
	pulumi package gen-sdk ./bin/$(PROVIDER_BINARY)

.DEFAULT_GOAL := build
