#!/usr/bin/env bash
#MISE description="Build local container image tagged with project name"
#MISE alias="bc"
#MISE sources=["Dockerfile","go.mod","go.sum","**/*.go","**/*.templ"]
set -euo pipefail
IMAGE="${IMAGE:-farm-manager}"
TAG="${TAG:-local}"
set -x
docker build -t "${IMAGE}:${TAG}" .
docker image ls "${IMAGE}:${TAG}"
