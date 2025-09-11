#!/usr/bin/env bash
#MISE description="Static analysis and linting"
#MISE alias="l"
#MISE sources=["go.mod","**/*.go",".golangci.yaml"]
set -euo pipefail
go vet ./...
golangci-lint run ./...
