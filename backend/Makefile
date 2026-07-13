.PHONY: help test test-unit test-integration test-coverage mocks lint fmt vet clean swagger

# Default target
help:
	@echo "Available targets:"
	@echo "  make test              - Run unit tests only (safe for local, used in CI)"
	@echo "  make test-all          - Run ALL tests including testcontainers (requires Colima)"
	@echo "  make test-coverage     - Run tests with coverage report (CI configuration)"
	@echo "  make mocks             - Generate mocks using mockery"
	@echo "  make swagger           - Regenerate OpenAPI/Swagger docs from swag annotations"
	@echo "  make lint              - Run golangci-lint"
	@echo "  make fmt               - Format code with gofmt"
	@echo "  make vet               - Run go vet"
	@echo "  make clean             - Remove build artifacts"
	@echo ""
	@echo "Test Info:"
	@echo "  - 'test' (default): Unit tests only, skip testcontainers"
	@echo "  - 'test-all': Requires Colima running (colima start --cpu 2 --memory 4)"

# Run unit tests (default for local dev and CI)
# Skips testcontainer tests via -short flag
test:
	go test -tags=test -v ./... -short

# Run ALL tests including testcontainers
# ⚠️ Requires Colima: colima start --cpu 2 --memory 4
test-all:
	@echo "⚠️  This requires Colima running. Check with: colima status"
	go test -tags=test -v ./...

# Legacy target - integration tests removed (use test-all instead)
# Integration tests were duplicate of testcontainer tests
test-integration:
	@echo "⚠️  test-integration target deprecated"
	@echo "   Use 'make test-all' instead (runs testcontainer tests with Colima)"
	@echo ""
	@echo "   Testcontainer tests provide same coverage with better isolation"
	@exit 1

# Run tests with coverage (CI configuration)
# Uses same config as CI: unit tests + race detector
test-coverage:
	go test -tags=test -v ./... -short -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate mocks using mockery
mocks:
	@echo "Generating mocks with mockery..."
	@if command -v mockery >/dev/null 2>&1; then \
		mockery --config .mockery.yaml --all; \
		echo "✅ Mocks generated successfully"; \
	else \
		echo "❌ mockery not found. Install with: go install github.com/vektra/mockery/v2@latest"; \
		exit 1; \
	fi

# Regenerate the OpenAPI/Swagger contract from swag doc-comment annotations
# See specs/009-api-documentation/ for the design behind this target.
swagger:
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/docs.go -o docs && \
		echo "✅ Swagger docs regenerated in backend/docs/"; \
	else \
		echo "❌ swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "❌ golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Format code
fmt:
	gofmt -w -s .
	go mod tidy

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -cache -testcache

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/vektra/mockery/v2@latest
	@echo "✅ Tools installed successfully"
	@echo "Note: You may need to install golangci-lint separately:"
	@echo "  https://golangci-lint.run/usage/install/"
