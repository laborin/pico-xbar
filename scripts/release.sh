#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Building ==="
"$SCRIPT_DIR/build.sh"

echo ""
echo "=== Signing & Notarizing ==="
"$SCRIPT_DIR/notarize.sh"

echo ""
echo "=== Done ==="
echo "Ready for distribution: build/pico-xbar-*.zip"
