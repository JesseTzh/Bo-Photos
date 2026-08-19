#!/usr/bin/env bash
# Boot the production image on an empty volume, then restart against the same volume.
set -euo pipefail

IMAGE="${CI_IMAGE:-bophotos-ci:local}"
CONTAINER="${CI_CONTAINER:-bophotos-ci}"
VOLUME="${CI_VOLUME:-bophotos-ci-data}"
HOST_PORT="${CI_HOST_PORT:-18080}"
READY_URL="http://127.0.0.1:${HOST_PORT}/health/ready"
INITIAL_PASSWORD="${BOPHOTOS_INITIAL_PASSWORD:-ci-initial-password}"

cleanup() {
  local status=$?
  if [[ "$status" -ne 0 ]]; then
    docker logs "$CONTAINER" 2>/dev/null || true
  fi
  docker rm --force "$CONTAINER" 2>/dev/null || true
  docker volume rm "$VOLUME" 2>/dev/null || true
  exit "$status"
}

wait_ready() {
  local attempts="$1"
  local attempt
  for attempt in $(seq 1 "$attempts"); do
    if curl --fail --silent "$READY_URL"; then
      return 0
    fi
    docker inspect --format '{{.State.Running}}' "$CONTAINER" | grep --quiet true
    sleep 1
  done
  return 1
}

trap cleanup EXIT

docker volume create "$VOLUME"

docker run --detach --name "$CONTAINER" \
  --publish "127.0.0.1:${HOST_PORT}:8080" \
  --env "BOPHOTOS_INITIAL_PASSWORD=$INITIAL_PASSWORD" \
  --volume "$VOLUME:/data" \
  "$IMAGE"
wait_ready 60

docker rm --force "$CONTAINER"
docker run --detach --name "$CONTAINER" \
  --publish "127.0.0.1:${HOST_PORT}:8080" \
  --volume "$VOLUME:/data" \
  "$IMAGE"
wait_ready 30
