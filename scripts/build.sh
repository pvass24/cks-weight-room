#!/usr/bin/env bash
set -e

# CKS Weight Room multi-platform build script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from git tag or use dev version
get_version() {
    if git describe --tags --exact-match >/dev/null 2>&1; then
        git describe --tags --exact-match
    elif git describe --tags >/dev/null 2>&1; then
        git describe --tags
    else
        echo "v0.1.0-dev"
    fi
}

VERSION=$(get_version)

echo ""
echo "====================================="
echo "  CKS Weight Room Build"
echo "  Version: $VERSION"
echo "====================================="
echo ""

# Step 1: Build Next.js frontend
echo -e "${YELLOW}Step 1/3: Building Next.js frontend...${NC}"
cd "$PROJECT_ROOT/web"
npm run build
echo -e "${GREEN}✓ Frontend build complete${NC}"
echo ""

# Step 2: Prepare dist directory
echo -e "${YELLOW}Step 2/3: Preparing distribution directory...${NC}"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
echo -e "${GREEN}✓ Dist directory ready${NC}"
echo ""

# Step 3: Cross-compile Go binaries
echo -e "${YELLOW}Step 3/3: Building binaries for all platforms...${NC}"
cd "$PROJECT_ROOT"

# Build for macOS (amd64 and arm64)
echo "  Building darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build \
    -ldflags "-X main.version=$VERSION" \
    -o "$DIST_DIR/cks-weight-room-darwin-amd64" \
    .

echo "  Building darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build \
    -ldflags "-X main.version=$VERSION" \
    -o "$DIST_DIR/cks-weight-room-darwin-arm64" \
    .

# Build for Linux (amd64 and arm64)
echo "  Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.version=$VERSION" \
    -o "$DIST_DIR/cks-weight-room-linux-amd64" \
    .

echo "  Building linux/arm64..."
GOOS=linux GOARCH=arm64 go build \
    -ldflags "-X main.version=$VERSION" \
    -o "$DIST_DIR/cks-weight-room-linux-arm64" \
    .

echo -e "${GREEN}✓ All binaries built successfully${NC}"
echo ""

# Show results
echo "====================================="
echo "  Build Complete!"
echo "====================================="
echo ""
echo "Binaries created in: $DIST_DIR"
ls -lh "$DIST_DIR"
echo ""

# Generate SHA256 checksums
echo -e "${YELLOW}Generating SHA256 checksums...${NC}"
cd "$DIST_DIR"
shasum -a 256 cks-weight-room-* > checksums.txt
echo -e "${GREEN}✓ Checksums saved to checksums.txt${NC}"
echo ""
cat checksums.txt
