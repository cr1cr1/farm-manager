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
    mise run build:container
    mise run build:program
    mise run run
    mise run release-snapshot

Dev hot reload

- mise run run runs two watchers: one for .templ (templ generate) and one for .go/go.mod/go.sum (go run). This avoids reload loops from templ output.

## Docker

    docker build -t farm-manager:local .
    docker run --rm -e PORT=8080 -p 8080:8080 farm-manager:local

## Health check

    curl -sS http://localhost:8080/healthz

## Releasing

- Tagging with v* (e.g., v0.1.0) triggers GitHub Actions to build and publish a GitHub Release via GoReleaser.
- For local testing, use: mise run release-snapshot

## Application bootstrap (auth + dashboard)

- Base path: /app (override via APP_BASE_PATH)
- Environment:
  - PORT=8080
  - APP_BASE_PATH=/app
  - SQLITE_DSN="file:./data/app.db?cache=shared&amp;mode=rwc"
  - SESSION_SECRET="set-a-32+-byte-secret"
  - RATE_LIMIT_RPS=10
  - RATE_LIMIT_BURST=20
  - CSRF_COOKIE_NAME=csrf_token
  - CSRF_HEADER_NAME=X-CSRF-Token
  - ADMIN_PASSWORD=<required> (initial admin password used for first-run seeding; must be set when no users exist)

### Run locally

1) Install tools and deps
   mise install
   go mod tidy

2) Generate templ code
   mise run templ

3) Start
   mise run run

   # or: go run ./cmd/farm-manager

4) Open:
   <http://localhost:8080/app/login>

### First login

- On first run, if the users table is empty and ADMIN_PASSWORD is set, the app seeds an admin user:
  - username: admin
  - password: value of ADMIN_PASSWORD
- Change this password immediately after logging in.
- If ADMIN_PASSWORD is not set, no user is seeded and an error is logged.

### CSRF & sessions

- CSRF token cookie is issued on safe methods and validated on POST/PUT/PATCH/DELETE via either:
  - Header: X-CSRF-Token, or
  - Hidden field: csrf_token
- Session-based auth using GoFrame sessions; logout at POST {APP_BASE_PATH}/logout.

### Static assets

- Served under /public
- Minimal CSS at /public/css/app.css and JS at /public/js/app.js
- JS enhances hypermedia by attaching CSRF header to fetch and initializing DataStar/DatastarUI when present.

### Database & migration

- SQLite DSN from SQLITE_DSN (default file:./data/app.db?cache=shared&amp;mode=rwc)
- On startup, migrations from db/migrations are applied. Tests discover migrations from common relative paths.
