#!/usr/bin/env bash
set -euo pipefail

echo "==> 1. Building UI..."
(cd ui && npm run build)

echo "==> 2. Copying built UI into Go embed directory..."
mkdir -p service/cmd/server/dist
rm -rf service/cmd/server/dist/*
cp -r ui/dist/* service/cmd/server/dist/

echo "==> 3. Compiling self-contained Go binary..."
mkdir -p bin
(cd service && go build -ldflags="-s -w" -o ../bin/kdb-service ./cmd/server)

echo "==> ✅ Build complete! Standalone binary created at bin/kdb-service ($(du -h bin/kdb-service | cut -f1))"
echo "==> Run with: ./bin/kdb-service"
