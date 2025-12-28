#!/bin/bash
# Build the secure terminal Docker image

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Building secure terminal image..."

cd "$PROJECT_ROOT/docker/terminal"

docker build -t cks-weight-room/terminal:latest .

echo "✅ Terminal image built successfully"
echo ""
echo "Image: cks-weight-room/terminal:latest"
echo ""
echo "To enable secure terminals, set SECURE_TERMINAL=true environment variable"
