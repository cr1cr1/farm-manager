#!/usr/bin/env bash
#MISE description="Create a local snapshot release using GoReleaser"
#MISE alias="rs"
set -euo pipefail
goreleaser release --snapshot --clean
