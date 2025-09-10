# Bootstrap a minimal Go 1.25 project with mise, GoFrame, Docker, templ, and GoReleaser

You are generating a minimal, production-ready starter for a Go 1.25 project using the GoFrame (gf) framework, mise-en-place (mise) for tool and environment management, a basic task set (lint, test, build, run, templ generate, release-snapshot), a multi-stage Dockerfile, and a GitHub Actions workflow to publish releases on tag push via GoReleaser. Keep scope minimal: no devcontainer, no databases.

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
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	s := g.Server()
	s.BindHandler("/healthz", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"version": version,
			"commit":  commit,
			"date":    date,
		})
	})
	s.SetAddr(":8080")
	s.Run()
}
```

3) mise.toml
```toml
[env]
CGO_ENABLED = "0"

[tools]
go = "1.25"
golangci-lint = "latest"
templ = "latest"
goreleaser = "latest"

[tasks.lint]
description = "Static analysis and linting"
run = "bash mise-tasks/lint.sh"

[tasks.test]
description = "Unit tests with race detector and coverage"
run = "bash mise-tasks/test.sh"

[tasks.build]
description = "Build binary to bin/farm-manager"
run = "bash mise-tasks/build.sh"

[tasks.run]
description = "Run the application locally"
run = "bash mise-tasks/run.sh"

[tasks.templ]
description = "Generate code from .templ files"
run = "bash mise-tasks/templ.sh"

[tasks.release-snapshot]
description = "Create a local snapshot release using GoReleaser"
run = "bash mise-tasks/release-snapshot.sh"
```

4) mise-tasks/lint.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
golangci-lint version >/dev/null 2>&1 || true
go vet ./...
golangci-lint run ./...
```

5) mise-tasks/test.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
mkdir -p ./.artifacts
go test ./... -race -covermode=atomic -coverprofile=./.artifacts/coverage.out
```

6) mise-tasks/build.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
mkdir -p ./bin
go mod tidy
GOFLAGS="-trimpath" CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION:-dev} -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ./bin/farm-manager ./cmd/farm-manager
```

7) mise-tasks/run.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
go run -ldflags "-X main.version=dev" ./cmd/farm-manager
```

8) mise-tasks/templ.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
templ generate ./...
```

9) mise-tasks/release-snapshot.sh
```bash
#!/usr/bin/env bash
set -euo pipefail
goreleaser release --snapshot --clean
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

14) README.md
```markdown
# farm-manager

Minimal Go 1.25 service using GoFrame, templ, mise task runner, Docker, and GoReleaser.

## Prerequisites
- Install mise: https://mise.jdx.dev
- Ensure Go 1.25 is installed via mise

## Setup

    mise install
    chmod +x mise-tasks/*.sh

## Common tasks

    mise run templ
    mise run lint
    mise run test
    mise run build
    mise run run
    mise run release-snapshot

## Docker

    docker build -t farm-manager:local .
    docker run --rm -p 8080:8080 farm-manager:local

## Health check

    curl -sS http://localhost:8080/healthz

## Releasing

- Tagging with v* (e.g., v0.1.0) triggers GitHub Actions to build and publish a GitHub Release via GoReleaser.
- For local testing, use: mise run release-snapshot
```

Acceptance criteria
- Building locally produces bin/farm-manager without CGO.
- Running locally responds 200 OK on GET /healthz with JSON including status and version metadata.
- mise run templ, lint, test, build, run, release-snapshot succeed on a clean checkout.
- Docker image builds and runs, exposing port 8080.
- Pushing a tag matching v* triggers the GitHub Actions workflow and completes successfully.

Implementation notes
- Keep dependencies minimal; go mod tidy will resolve GoFrame.
- Scripts use set -euo pipefail and should be executable.
- GoReleaser embeds version info via ldflags; main.go exposes it in /healthz.
- No devcontainer, no database.
