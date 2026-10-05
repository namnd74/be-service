#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
MIN_COVERAGE="${MIN_COVERAGE:-60}"
[[ "$MIN_COVERAGE" =~ ^[0-9]+([.][0-9]+)?$ ]] || { echo "MIN_COVERAGE must be a number" >&2; exit 1; }

echo "== go test =="
go test -coverprofile=coverage.out ./...
coverage=$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/[()%]/, "", $3); print $3}')
awk -v actual="$coverage" -v minimum="$MIN_COVERAGE" 'BEGIN { if (actual == "" || actual + 0 < minimum + 0) { printf "coverage %.1f%% is below minimum %.1f%%\n", actual, minimum > "/dev/stderr"; exit 1 } }'
echo "coverage ${coverage}% (minimum ${MIN_COVERAGE}%)"
echo "== go test -race =="
go test -race ./...
echo "== go vet =="
go vet ./...
echo "== gofmt =="
if files=$(gofmt -l .) && [[ -z "$files" ]]; then
  echo "gofmt clean"
else
  printf 'gofmt required for:\n%s\n' "$files" >&2
  exit 1
fi
