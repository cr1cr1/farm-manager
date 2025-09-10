#!/usr/bin/env bash
#MISE description="Create a local snapshot release using GoReleaser"
#MISE short="release-snapshot"
set -euo pipefail
goreleaser release --snapshot --clean
