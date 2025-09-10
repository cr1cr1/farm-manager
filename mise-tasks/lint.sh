#!/usr/bin/env bash
#MISE description="Static analysis and linting"
#MISE short="lint"
#MISE sources=["go.mod","**/*.go",".golangci.yaml"]
set -euo pipefail
golangci-lint version >/dev/null 2>&1 || true
go vet ./...
golangci-lint run ./...
