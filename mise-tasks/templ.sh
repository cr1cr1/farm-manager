#!/usr/bin/env bash
#MISE description="Generate code from .templ files"
#MISE short="templ"
#MISE sources=["**/*.templ"]
set -euo pipefail
templ generate ./...
