# Bootstrap a minimal Go 1.25 project with mise, GoFrame, Docker, templ, GoReleaser, and 12-factor env config

You are generating a minimal, production-ready starter for a Go 1.25 project using the GoFrame (gf) framework, mise-en-place (mise) for tool and environment management, a basic task set (lint, test, build, run, templ generate, release-snapshot), a multi-stage Dockerfile, and a GitHub Actions workflow to publish releases on tag push via GoReleaser.

Follow 12-factor app principles:
- Configuration must be provided via environment variables (no config files required for defaults).
- Do not commit secrets; provide a .env.example for local development and ensure .env is git-ignored.

Project metadata
- Module path: github.com/cr1cr1/farm-manager
- Go version: 1.25
- Default HTTP port: 8080

Deliverables
Generate the following files with exact contents, creating directories as needed.

1) go.mod
```toml
module github.com/cr1cr1/farm-manager
go 1.25
```

2) cmd/farm-manager/main.go
```go
package main

import (
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func addrFromEnv() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}

func main() {
	s := g.Server()
	s.BindHandler("/healthz", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"version": version,
			"commit":  commit,
			"date":    date,
			"addr":    addrFromEnv(),
		})
	})
	s.SetAddr(addrFromEnv())
	s.Run()
}
```

3) mise.toml
```toml
[env]
CGO_ENABLED = "0"

[tools]
go = "1.25"
goreleaser = "latest"
"aqua:golangci-lint" = "latest"
ubi = "latest"
"ubi:bokwoon95/wgo" = "latest"
"aqua:a-h/templ" = "latest"
```

4) mise-tasks/lint.sh
```bash
#!/usr/bin/env bash
#MISE description="Static analysis and linting"
#MISE short="lint"
#MISE sources=["go.mod","**/*.go",".golangci.yaml"]
set -euo pipefail
golangci-lint version >/dev/null 2>&1 || true
go vet ./...
golangci-lint run ./...
```

5) mise-tasks/test.sh
```bash
#!/usr/bin/env bash
#MISE description="Unit tests with race detector and coverage"
#MISE short="test"
#MISE sources=["go.mod","**/*.go"]
set -euo pipefail
mkdir -p ./.artifacts
go test ./... -race -covermode=atomic -coverprofile=./.artifacts/coverage.out
```

6) mise-tasks/build.sh
```bash
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
```

7) mise-tasks/run.sh
```bash
#!/usr/bin/env bash
#MISE description="Run the application locally with hot reloading (wgo + templ generate)"
#MISE short="run"
#MISE sources=["**/*.go","**/*.templ","go.mod","go.sum"]
set -euo pipefail
# Watch .templ and .go; regenerate templates, then restart the app
exec wgo -file=.templ -file=.go templ generate :: go run ./cmd/farm-manager
```

8) mise-tasks/templ.sh
```bash
#!/usr/bin/env bash
#MISE description="Generate code from .templ files"
#MISE short="templ"
#MISE sources=["**/*.templ"]
set -euo pipefail
templ generate ./...
```

9) mise-tasks/release-snapshot.sh
```bash
#!/usr/bin/env bash
#MISE description="Create a local snapshot release using GoReleaser"
#MISE short="release-snapshot"
set -euo pipefail
goreleaser release --snapshot --clean
```

10) mise-tasks/mod-tidy.sh
```bash
#!/usr/bin/env bash
#MISE description="Tidy go.mod and go.sum"
#MISE short="mod-tidy"
#MISE sources=["go.mod","go.sum"]
#MISE outputs=["go.mod","go.sum"]
set -euo pipefail
go mod tidy
# go mod vendor # Uncomment if you want to vendor dependencies
```

10) .goreleaser.yaml
```yaml
version: 2

project_name: farm-manager

builds:
  - id: farm-manager
    main: ./cmd/farm-manager
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: archives
    builds:
      - farm-manager
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  use: github

release:
  github:
    owner: cr1cr1
    name: farm-manager
  draft: false
  prerelease: auto
```

11) .github/workflows/release.yml
```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.25.x"

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

12) Dockerfile
```Dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/farm-manager ./cmd/farm-manager

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /out/farm-manager /farm-manager
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/farm-manager"]
```

13) .dockerignore
```gitignore
.git
bin
.mise
.mise-local.*
Dockerfile
README.md
*.log
**/node_modules
**/.DS_Store
```

14) .env.example
```gitignore
# Example environment variables for local development
# Copy to .env and adjust values as needed (do not commit .env)
PORT=8080
```

15) .gitignore
```gitignore
# Binaries and build artifacts
bin/
.artifacts/
coverage*
*.log

# Local env and tools
.env
.mise
.mise-local.*
.DS_Store
```

16) README.md
```markdown
# farm-manager

Minimal Go 1.25 service using GoFrame, templ, mise task runner, Docker, and GoReleaser. Follows 12-factor principles: configuration via environment variables.

## Prerequisites
- Install mise: https://mise.jdx.dev
- Ensure Go 1.25 is installed via mise

## Setup

    mise install
    chmod +x mise-tasks/*.sh

## Configuration (12-factor)

- Configuration is provided via environment variables.
- PORT controls the HTTP listen port (default 8080 when unset).
- Do not commit secrets. Use a local .env (not committed) or export in your shell.
- Example:

      export PORT=9090
      mise run run
      # or copy .env.example to .env and load via your shell tooling (e.g., direnv)

## Common tasks

    mise run templ
    mise run lint
    mise run test
    mise run mod-tidy
    mise run build
    mise run run
    mise run release-snapshot

Dev hot reload
- mise run run uses wgo to watch .templ and .go files, runs templ generate, then restarts the app.

## Docker

    docker build -t farm-manager:local .
    docker run --rm -e PORT=8080 -p 8080:8080 farm-manager:local

## Health check

    curl -sS http://localhost:8080/healthz

## Releasing

- Tagging with v* (e.g., v0.1.0) triggers GitHub Actions to build and publish a GitHub Release via GoReleaser.
- For local testing, use: mise run release-snapshot
```

Acceptance criteria
- Building locally produces bin/farm-manager without CGO.
- Running locally responds 200 OK on GET /healthz with JSON including status and version metadata.
- Setting PORT environment variable changes the listen port accordingly.
- mise run templ, lint, test, build, run, release-snapshot succeed on a clean checkout.
- Docker image builds and runs, exposing port 8080.
- Pushing a tag matching v* triggers the GitHub Actions workflow and completes successfully.

Implementation notes
- Respect 12-factor: configuration exclusively via environment variables; do not commit secrets.
- Provide .env.example; ensure .env is ignored by git.
- Keep dependencies minimal; go mod tidy will resolve GoFrame.
- Scripts use set -euo pipefail and should be executable.
- Define mise tasks via #MISE headers in scripts under mise-tasks/; do not define tasks in mise.toml.
- GoReleaser embeds version info via ldflags; main.go exposes it in /healthz.
- No devcontainer, no database.
