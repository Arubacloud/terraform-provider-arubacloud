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
	@if command -v golangci-lint >/dev/null 2>&1 && \
	    golangci-lint version --format short 2>/dev/null | grep -qE '^2\.'; then \
		golangci-lint run; \
	else \
		echo "golangci-lint v2 not available. Running via go run..."; \
		go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run; \
	fi

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

# Run all acceptance tests with structured logging.
# Credentials via env vars: ARUBACLOUD_CLIENT_ID, ARUBACLOUD_CLIENT_SECRET, ARUBACLOUD_PROJECT_ID
# Optional: ARUBACLOUD_OS_IMAGE_ID, ARUBACLOUD_DBAAS_ID, ARUBACLOUD_VPNTUNNEL_ID, ARUBACLOUD_VPNROUTE_ID
# Pass extra flags with ARGS: make testacc ARGS="--log-level DEBUG --timeout 60m"
testacc:
	@./run-acceptance-tests.sh $(ARGS)

# Run a single acceptance test by name.
# Usage: make testacc-run TEST=TestAccBackupResource
# Optional: make testacc-run TEST=TestAccVpc ARGS="--log-level DEBUG"
testacc-run:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make testacc-run TEST=<TestName>"; \
		echo "  e.g. make testacc-run TEST=TestAccBackupResource"; \
		exit 1; \
	fi
	@./run-acceptance-tests.sh --run '^$(TEST)$$' $(ARGS)

# Show the most recent acceptance test summary.
testacc-summary:
	@LATEST=$$(ls -t artifacts/summary-*.txt 2>/dev/null | head -1); \
	if [ -z "$$LATEST" ]; then \
		echo "No summary found. Run 'make testacc' first."; exit 1; \
	fi; \
	echo "==> $$LATEST"; echo ""; cat "$$LATEST"

# Run tests with coverage report
testcov:
	go test -v -coverprofile=coverage.out -timeout=120s ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# ---- Docs generation ----

docs:
	@echo "Generating documentation..."
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0 generate \
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
	@make lint || (echo "✗ Linter failed"; exit 1); \
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

.PHONY: default fmt lint test testacc testacc-run testacc-summary testcov build install docs generate ci-test
