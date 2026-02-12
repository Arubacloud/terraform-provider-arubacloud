TFPLUGINDOCS_PROVIDER_NAME      ?= arubacloud
TFPLUGINDOCS_RENDERED_NAME      ?= ArubaCloud
TFPLUGINDOCS_PROVIDER_DIR       ?= .
TFPLUGINDOCS_RENDERED_WEBSITE   ?= docs
TFPLUGINDOCS_TEMPLATES_DIR      ?= templates
TFPLUGINDOCS_EXAMPLES_DIR       ?= examples

# Default target when you just run `make`
default: fmt lint test build generate

# ---- Build / install / lint / tests ----

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	@GOLANGCI_LINT=$$(command -v golangci-lint 2>/dev/null || echo ""); \
	GOPATH_BIN=$$(go env GOPATH)/bin; \
	if [ -z "$$GOLANGCI_LINT" ]; then \
		echo "golangci-lint not found. Installing latest version..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOPATH_BIN latest; \
		GOLANGCI_LINT=$$GOPATH_BIN/golangci-lint; \
	fi; \
	if ! $$GOLANGCI_LINT version 2>/dev/null | grep -qE "(version 2\.|v2\.)"; then \
		echo "Installed version may not support config v2. Testing..."; \
		if ! $$GOLANGCI_LINT run --help 2>&1 >/dev/null; then \
			echo "Reinstalling latest version..."; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOPATH_BIN latest; \
			GOLANGCI_LINT=$$GOPATH_BIN/golangci-lint; \
		fi; \
	fi; \
	$$GOLANGCI_LINT run

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

# Run acceptance tests (requires TF_ACC=1)
testacc:
	TF_ACC=1 go test -v -cover -timeout=120m ./internal/provider/...

# Run specific acceptance test
testacc-run:
	@echo "Usage: make testacc-run TEST=TestAccBackupResource"
	TF_ACC=1 go test -v ./internal/provider/ -run $(TEST)

# Run tests with coverage report
testcov:
	go test -v -coverprofile=coverage.out -timeout=120s ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# ---- Docs generation ----

docs:
	@echo "Ensuring terraform-plugin-docs is available..."
	@go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest || true
	@echo "Generating documentation..."
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate \
		--provider-dir $(TFPLUGINDOCS_PROVIDER_DIR) \
		--provider-name $(TFPLUGINDOCS_PROVIDER_NAME) \
		--rendered-provider-name "$(TFPLUGINDOCS_RENDERED_NAME)" \
		--rendered-website-dir $(TFPLUGINDOCS_RENDERED_WEBSITE) \
		--website-source-dir $(TFPLUGINDOCS_TEMPLATES_DIR) \
		--examples-dir $(TFPLUGINDOCS_EXAMPLES_DIR)
	@echo "Formatting documentation to separate Arguments and Attributes..."
	@bash scripts/format-docs.sh $(TFPLUGINDOCS_RENDERED_WEBSITE) || echo "Note: format-docs.sh script not executed (bash not available or script missing)"

generate: docs
	cd tools && go generate ./...

# CI test target - runs all CI checks locally
ci-test:
	@echo "=== Running CI Tests Locally ==="
	@echo ""
	@echo "1. Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies downloaded"
	@echo ""
	@echo "2. Building..."
	@go build -v . || (echo "✗ Build failed"; exit 1)
	@echo "✓ Build successful"
	@echo ""
	@echo "3. Running linter..."
	@GOLANGCI_LINT=$$(command -v golangci-lint 2>/dev/null || echo ""); \
	GOPATH_BIN=$$(go env GOPATH)/bin; \
	if [ -z "$$GOLANGCI_LINT" ]; then \
		echo "golangci-lint not found. Installing latest version..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$GOPATH_BIN latest; \
		GOLANGCI_LINT=$$GOPATH_BIN/golangci-lint; \
	fi; \
	$$GOLANGCI_LINT run || (echo "✗ Linter failed"; exit 1); \
	echo "✓ Linter passed"
	@echo ""
	@echo "4. Running make generate..."
	@make generate || (echo "✗ Generate failed"; exit 1)
	@echo "✓ Generate successful"
	@echo ""
	@echo "5. Running go mod tidy..."
	@go mod tidy
	@echo "✓ go mod tidy completed"
	@echo ""
	@echo "6. Checking for uncommitted changes..."
	@git diff --compact-summary --exit-code || \
		(echo ""; echo "✗ Unexpected changes detected after code generation!"; \
		 echo "Run 'make generate' and 'go mod tidy', then commit the changes."; \
		 git diff --compact-summary; exit 1)
	@echo "✓ No unexpected changes after generation"
	@echo ""
	@echo "7. Running unit tests..."
	@go test -v -cover -timeout=120s ./... || (echo "✗ Unit tests failed"; exit 1)
	@echo "✓ Unit tests passed"
	@echo ""
	@echo "8. Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
	@echo ""
	@echo "=== All CI checks passed! ==="

.PHONY: default fmt lint test testacc testacc-run testcov build install docs generate ci-test
