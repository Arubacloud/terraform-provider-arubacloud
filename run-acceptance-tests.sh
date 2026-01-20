#!/bin/bash

# This script runs acceptance tests with API credentials from terraform.tfvars
# Usage: ./run-acceptance-tests.sh [test-pattern]
# Example: ./run-acceptance-tests.sh TestAccProjectDataSource

# Load credentials from terraform.tfvars
TFVARS_FILE="examples/test/compute/terraform.tfvars"

if [ -f "$TFVARS_FILE" ]; then
    export ARUBACLOUD_API_KEY=$(grep arubacloud_api_key "$TFVARS_FILE" | cut -d'"' -f2)
    export ARUBACLOUD_API_SECRET=$(grep arubacloud_api_secret "$TFVARS_FILE" | cut -d'"' -f2)
    echo "✓ Loaded credentials from $TFVARS_FILE"
else
    echo "⚠ Warning: $TFVARS_FILE not found. Using environment variables if set."
fi

# Enable acceptance tests
export TF_ACC=1

# Run tests
if [ -z "$1" ]; then
    echo "Running all acceptance tests..."
    go test -v -timeout=120m ./internal/provider/...
else
    echo "Running test pattern: $1"
    go test -v -timeout=120m ./internal/provider/... -run "$1"
fi
