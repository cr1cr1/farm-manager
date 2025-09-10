#!/usr/bin/env bash
#MISE description="Unit tests with race detector and coverage"
#MISE short="test"
#MISE sources=["go.mod","**/*.go"]
set -euo pipefail
mkdir -p ./.artifacts
go test ./... -race -covermode=atomic -coverprofile=./.artifacts/coverage.out
