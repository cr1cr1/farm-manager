#!/usr/bin/env bash
#MISE description="Build Tailwind CSS"
#MISE alias="bt"
#MISE sources=["app.css"]
set -euo pipefail

[[ -f pnpm-lock.yaml ]] || pnpm install
set -x
pnpm exec tailwindcss -i app.css -o public/css/app.css --content "./internal/web/**/*" --content "./cmd/**/*"
