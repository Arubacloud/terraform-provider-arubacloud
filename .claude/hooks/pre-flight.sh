#!/usr/bin/env bash
# Pre-flight checks: runs before Claude Code finishes a task.
# Exits 2 to re-prompt Claude if any check fails.

FAILED=0

check_go_changes() {
  git diff --name-only HEAD 2>/dev/null | grep -qE '\.go$|go\.mod|go\.sum' ||
    git diff --cached --name-only 2>/dev/null | grep -qE '\.go$|go\.mod|go\.sum'
}

if ! check_go_changes; then
  exit 0
fi

echo "=== Pre-flight checks ==="

# 1. gofmt
printf "\n--- gofmt ---\n"
GOFMT_OUT=$(gofmt -s -l . 2>/dev/null)
if [ -n "$GOFMT_OUT" ]; then
  echo "FAIL: gofmt would reformat these files:"
  echo "$GOFMT_OUT"
  FAILED=1
else
  echo "OK: gofmt"
fi

# 2. go build
printf "\n--- go build ---\n"
if go build ./... 2>&1; then
  echo "OK: go build"
else
  echo "FAIL: go build"
  FAILED=1
fi

# 3. go mod tidy drift
printf "\n--- go mod tidy ---\n"
go mod tidy 2>&1
if git diff --exit-code go.mod go.sum > /dev/null 2>&1; then
  echo "OK: go.mod / go.sum in sync"
else
  echo "FAIL: go mod tidy changed go.mod or go.sum — commit the updated files"
  git diff --stat go.mod go.sum
  FAILED=1
fi

# 4. go vet
printf "\n--- go vet ---\n"
if go vet ./... 2>&1; then
  echo "OK: go vet"
else
  echo "FAIL: go vet"
  FAILED=1
fi

# 5. golangci-lint (skip if not installed)
printf "\n--- golangci-lint ---\n"
if command -v golangci-lint > /dev/null 2>&1; then
  if golangci-lint run 2>&1; then
    echo "OK: golangci-lint"
  else
    echo "FAIL: golangci-lint"
    FAILED=1
  fi
else
  echo "SKIP: golangci-lint not found in PATH"
fi

# 6. make generate — only when provider source or templates changed
printf "\n--- make generate ---\n"
GEN_CHANGED=$(git diff --name-only HEAD 2>/dev/null | grep -cE 'internal/provider/.*\.go$|templates/' || true)
if command -v make > /dev/null 2>&1 && [ "$GEN_CHANGED" -gt 0 ]; then
  if make generate 2>&1; then
    if git diff --compact-summary --exit-code > /dev/null 2>&1; then
      echo "OK: make generate (no drift)"
    else
      echo "FAIL: make generate produced uncommitted changes — commit or stash them"
      git diff --compact-summary
      FAILED=1
    fi
  else
    echo "FAIL: make generate"
    FAILED=1
  fi
elif ! command -v make > /dev/null 2>&1; then
  echo "SKIP: make not found in PATH"
else
  echo "SKIP: no docs-relevant source changes"
fi

# 7. Unit tests (non-acceptance only, short timeout)
printf "\n--- unit tests ---\n"
if go test ./... -count=1 -timeout=120s -run "^Test[^A]" 2>&1; then
  echo "OK: unit tests"
else
  echo "FAIL: unit tests"
  FAILED=1
fi

echo ""
if [ "$FAILED" -eq 1 ]; then
  echo "Pre-flight FAILED — fix the issues above."
  exit 2
fi

echo "All pre-flight checks passed."
