#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/root/bo-photos"

cd "$APP_DIR"

docker compose pull
docker compose up -d
docker image prune -f
