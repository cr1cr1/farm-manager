#!/usr/bin/env bash
#MISE description="Tidy go.mod and go.sum"
#MISE alias="mt"
#MISE sources=["go.mod","go.sum"]
#MISE outputs=["go.mod","go.sum"]
set -euo pipefail
go mod tidy
go mod vendor
