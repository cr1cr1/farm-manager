#!/usr/bin/env bash
#MISE description="Build binary to bin/farm-manager"
#MISE short="build"
#MISE env={CGO_ENABLED="0"}
#MISE sources=["go.mod","**/*.go"]
#MISE outputs=["bin/farm-manager"]
set -euo pipefail
mkdir -p ./bin
go mod tidy
GOFLAGS="-trimpath" CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION:-dev} -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ./bin/farm-manager ./cmd/farm-manager
