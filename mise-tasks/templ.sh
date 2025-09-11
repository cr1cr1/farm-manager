#!/usr/bin/env bash
#MISE description="Generate code from .templ files"
#MISE alias="t"
#MISE sources=["**/*.templ"]
set -euo pipefail
templ generate ./...
