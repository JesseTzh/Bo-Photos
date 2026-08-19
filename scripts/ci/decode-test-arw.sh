#!/usr/bin/env bash
# Decode test.ARW with the tools inside the production image.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${CI_IMAGE:-bophotos-ci:local}"
FIXTURE="${RAW_FIXTURE:-$ROOT_DIR/test.ARW}"
INNER="$ROOT_DIR/scripts/ci/decode-test-arw-in-image.sh"

[[ -f "$FIXTURE" ]] || {
  printf 'fixture not found: %s\n' "$FIXTURE" >&2
  exit 1
}
[[ -f "$INNER" ]] || {
  printf 'in-image script not found: %s\n' "$INNER" >&2
  exit 1
}

docker run --rm \
  --volume "$FIXTURE:/fixture/test.ARW:ro" \
  --volume "$INNER:/ci/decode-test-arw-in-image.sh:ro" \
  --entrypoint bash \
  "$IMAGE" \
  /ci/decode-test-arw-in-image.sh
