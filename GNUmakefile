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

# ---- Docs generation ----

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
		--provider-dir $(TFPLUGINDOCS_PROVIDER_DIR) \
		--provider-name $(TFPLUGINDOCS_PROVIDER_NAME) \
		--rendered-provider-name "$(TFPLUGINDOCS_RENDERED_NAME)" \
		--rendered-website-dir $(TFPLUGINDOCS_RENDERED_WEBSITE) \
		--website-source-dir $(TFPLUGINDOCS_TEMPLATES_DIR) \
		--examples-dir $(TFPLUGINDOCS_EXAMPLES_DIR)
	@echo "Formatting documentation to separate Arguments and Attributes..."
	@bash scripts/format-docs.sh $(TFPLUGINDOCS_RENDERED_WEBSITE) || echo "Note: format-docs.sh script not executed (bash not available or script missing)"

generate: docs
	cd tools && go generate ./... && \
	git add docs/ && \
	git commit -m "Update generated documentation" || echo "No changes to commit."

.PHONY: default fmt lint test build install docs generate
