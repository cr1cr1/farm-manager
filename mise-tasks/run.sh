#!/usr/bin/env bash
#MISE description="Run the application locally with hot reloading (wgo + templ generate)"
#MISE alias="r"
#MISE sources=["**/*.go","**/*.templ","go.mod","go.sum"]
set -euo pipefail

# Prevent loop: run two watchers independently:
# - watcher A: watches only .templ and runs templ generate
# - watcher B: watches only .go/go.mod/go.sum and restarts the app
# We also generate once before starting.
mise run templ

# Ensure both watchers are terminated when the script exits.
trap 'kill 0' EXIT

# Watch templ files and regenerate on change (no app restart here).
wgo -file=.templ -file=./app.css mise run templ :: mise run build:tailwind &

# Watch Go files and restart the app on change.
wgo -file=.go -file=go.mod -file=go.sum go run ./cmd/farm-manager
