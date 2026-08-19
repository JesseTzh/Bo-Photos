#!/usr/bin/env bash
# Run inside the production image (or any host with the same tools).
# Asserts LibRaw can fully decode the Sony ARW fixture, then emit a WebP.
set -euo pipefail

FIXTURE="${1:-${RAW_FIXTURE:-/fixture/test.ARW}}"
WORK="${RAW_WORK:-/tmp/test.ARW}"
WEBP="${RAW_WEBP:-/tmp/raw-ci.webp}"
MIN_WIDTH="${RAW_MIN_WIDTH:-6000}"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null || die "missing required command: $1"
}

first_existing() {
  local path
  for path in "$@"; do
    if [[ -s "$path" ]]; then
      printf '%s\n' "$path"
      return 0
    fi
  done
  return 1
}

webp_fourcc() {
  # od is in coreutils (present on debian slim). hexdump is not.
  od -An -tx1 -j "$1" -N 4 "$2" | tr -d ' \n'
}

[[ -f "$FIXTURE" ]] || die "fixture not found: $FIXTURE"
require_cmd exiftool
require_cmd vipsheader
require_cmd vipsthumbnail
require_cmd od

filetype=$(exiftool -s3 -FileType "$FIXTURE")
[[ "$filetype" == "ARW" ]] || die "FileType=$filetype, want ARW"

model=$(exiftool -s3 -Model "$FIXTURE")
[[ "$model" == "ILCE-6400" ]] || die "Model=$model, want ILCE-6400"

cp "$FIXTURE" "$WORK"

if command -v simple_dcraw >/dev/null; then
  decoder=simple_dcraw
elif command -v dcraw_emu >/dev/null; then
  decoder=dcraw_emu
else
  die "neither simple_dcraw nor dcraw_emu is on PATH"
fi

echo "decoding with $decoder -T $WORK"
"$decoder" -T "$WORK"

ext=".ARW"
base="${WORK%"$ext"}"
decoded=$(first_existing \
  "${WORK}.tiff" \
  "${WORK}.tif" \
  "${base}.tiff" \
  "${base}.tif" \
  "${WORK}.ppm" \
  "${base}.ppm" \
) || die "LibRaw produced no TIFF/PPM next to $WORK"

header=$(vipsheader "$decoded")
echo "$header"

width=$(printf '%s\n' "$header" | sed -n 's/^[^:]*: \([0-9][0-9]*\)x.*/\1/p')
[[ -n "$width" ]] || die "could not parse width from vipsheader"
[[ "$width" =~ ^[0-9]+$ ]] || die "non-numeric width: $width"
(( width >= MIN_WIDTH )) || die "decoded width $width < $MIN_WIDTH ($decoded)"

vipsthumbnail "$decoded" \
  --size '480x480>' \
  --rotate \
  --export-profile srgb \
  --output "${WEBP}[Q=80,strip]"

[[ -s "$WEBP" ]] || die "vipsthumbnail produced empty $WEBP"

riff=$(webp_fourcc 0 "$WEBP")
webp=$(webp_fourcc 8 "$WEBP")
[[ "$riff" == "52494646" ]] || die "WebP missing RIFF header: $riff"
[[ "$webp" == "57454250" ]] || die "WebP missing WEBP fourcc: $webp"

echo "ok: $decoded ${width}px -> $WEBP"
