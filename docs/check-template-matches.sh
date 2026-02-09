#!/usr/bin/env bash
# Check for mismatches or missing templates between resources and data-sources

set -e
cd "$(dirname "$0")"

RESOURCE_DIR="resources"
DATA_SOURCE_DIR="data-sources"

# Strip .md.tmpl and sort
resources=$(ls "$RESOURCE_DIR" | sed 's/\.md\.tmpl$//' | sort)
datasources=$(ls "$DATA_SOURCE_DIR" | sed 's/\.md\.tmpl$//' | sort)

# Find in resources but not in data-sources
missing_in_datasources=$(comm -23 <(echo "$resources") <(echo "$datasources"))
# Find in data-sources but not in resources
missing_in_resources=$(comm -13 <(echo "$resources") <(echo "$datasources"))

if [[ -n "$missing_in_datasources" ]]; then
  echo "Present in resources but missing in data-sources:" >&2
  echo "$missing_in_datasources"
else
  echo "No mismatches: all resources have a data-source template." >&2
fi

if [[ -n "$missing_in_resources" ]]; then
  echo "Present in data-sources but missing in resources:" >&2
  echo "$missing_in_resources"
else
  echo "No mismatches: all data-sources have a resource template." >&2
fi
