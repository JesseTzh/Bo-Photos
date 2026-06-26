#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

: "${BOPHOTOS_ADDR:=:8080}"
: "${BOPHOTOS_DATA_DIR:=$ROOT_DIR/data}"
: "${BOPHOTOS_INITIAL_PASSWORD:=change-this-to-a-strong-password}"
: "${BOPHOTOS_FRONTEND_DIR:=$ROOT_DIR/frontend/dist}"
: "${BOPHOTOS_FRONTEND_PORT:=5173}"

BACKEND_PID=""
FRONTEND_PID=""

backend_url() {
  if [[ "$BOPHOTOS_ADDR" == :* ]]; then
    printf "http://127.0.0.1%s" "$BOPHOTOS_ADDR"
  elif [[ "$BOPHOTOS_ADDR" == 0.0.0.0:* ]]; then
    printf "http://127.0.0.1:%s" "${BOPHOTOS_ADDR##*:}"
  else
    printf "http://%s" "$BOPHOTOS_ADDR"
  fi
}

cleanup() {
  local exit_code=$?

  trap - INT TERM EXIT

  if [[ -n "$FRONTEND_PID" ]] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi

  if [[ -n "$BACKEND_PID" ]] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi

  wait "$FRONTEND_PID" 2>/dev/null || true
  wait "$BACKEND_PID" 2>/dev/null || true

  exit "$exit_code"
}

trap cleanup INT TERM EXIT

mkdir -p "$BOPHOTOS_DATA_DIR"

echo "Starting BoPhoto dev stack"
echo "  Backend:  $(backend_url)"
echo "  Frontend: http://127.0.0.1:$BOPHOTOS_FRONTEND_PORT"
echo "  Data dir: $BOPHOTOS_DATA_DIR"

(
  cd "$ROOT_DIR/backend"
  BOPHOTOS_ADDR="$BOPHOTOS_ADDR" \
    BOPHOTOS_DATA_DIR="$BOPHOTOS_DATA_DIR" \
    BOPHOTOS_INITIAL_PASSWORD="$BOPHOTOS_INITIAL_PASSWORD" \
    BOPHOTOS_FRONTEND_DIR="$BOPHOTOS_FRONTEND_DIR" \
    go run ./cmd/server
) &
BACKEND_PID=$!

(
  cd "$ROOT_DIR/frontend"
  pnpm dev --host 127.0.0.1 --port "$BOPHOTOS_FRONTEND_PORT"
) &
FRONTEND_PID=$!

while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 1
done
