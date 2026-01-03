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
	golangci-lint run

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

.PHONY: default fmt lint test testacc testacc-run testcov build install docs generate
