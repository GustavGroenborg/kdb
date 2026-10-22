#!/usr/bin/env bash
set -euo pipefail

GOARCH="${GOARCH:-$(go env GOARCH)}"
GOOS="${GOOS:-linux}"
OUTPUT_BIN="${OUTPUT_BIN:-../bin/kdb-service}"

echo "==> 1. Building UI..."
(cd ui && [ -d node_modules ] || npm ci; npm run build)

echo "==> 2. Copying built UI into Go embed directory..."
mkdir -p service/cmd/server/dist
rm -rf service/cmd/server/dist/*
cp -r ui/dist/* service/cmd/server/dist/

echo "==> 3. Compiling self-contained Go binary for ${GOOS}/${GOARCH}..."
mkdir -p bin
(cd service && CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build -ldflags="-s -w" -o "${OUTPUT_BIN}" ./cmd/server)

echo "==> ✅ Build complete! Standalone binary created at ${OUTPUT_BIN} ($(du -h "${OUTPUT_BIN#../}" | cut -f1))"
echo "==> Run with: ./${OUTPUT_BIN#../}"
