#!/usr/bin/env bash
# scripts/check-local.sh — run all quality checks locally (mirrors CI)
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

echo "==> coverage threshold (≥70%)"
go test -coverprofile=coverage.out -covermode=atomic ./... > /dev/null
TOTAL=$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $3}' | tr -d '%')
echo "    Total coverage: ${TOTAL}%"
awk -v t="$TOTAL" 'BEGIN { if (t+0 < 70) { print "Coverage " t "% is below 70% threshold"; exit 1 } }'
rm -f coverage.out

echo ""
echo "All checks passed."
