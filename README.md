# farm-manager

Minimal Go 1.25 service using GoFrame, templ, mise task runner, Docker, and GoReleaser. Follows 12-factor principles: configuration via environment variables.

## Prerequisites

- Install mise: <https://mise.jdx.dev>
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
