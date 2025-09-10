#!/usr/bin/env bash
#MISE description="Unit tests with race and coverage"
#MISE short="test"
#MISE sources=["go.mod","**/*.go"]
set -euo pipefail
mkdir -p ./.artifacts
CGO_ENABLED=1 go test ./... -race -covermode=atomic -coverprofile=./.artifacts/coverage.out
