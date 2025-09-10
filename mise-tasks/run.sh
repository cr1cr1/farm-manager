#!/usr/bin/env bash
#MISE description="Run the application locally with hot reloading (wgo + templ generate)"
#MISE short="run"
#MISE sources=["**/*.go","**/*.templ","go.mod","go.sum"]
set -euo pipefail
# Watch .templ and .go; regenerate templates, then restart the app
exec wgo -file=.templ -file=.go templ generate :: go run ./cmd/farm-manager
