#!/usr/bin/env bash
# scripts/check.sh — run all quality checks locally (mirrors CI)
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> go vet"
go vet ./...

echo "==> go build"
go build ./...

echo "==> go test"
go test -v -count=1 ./...

echo "==> go test -race"
go test -race -count=1 ./...

echo ""
echo "All checks passed."
