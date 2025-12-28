#!/usr/bin/env bash
set -e

# CKS Weight Room installation script
# Supports macOS and Linux (curl method)

REPO="patrickvassell/cks-weight-room"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="cks-weight-room"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Map architecture names
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    # Check OS support
    case "$OS" in
        darwin)
            PLATFORM="darwin-$ARCH"
            ;;
        linux)
            PLATFORM="linux-$ARCH"
            ;;
        mingw*|msys*|cygwin*)
            echo -e "${RED}Error: CKS Weight Room requires macOS or Linux. Windows is not supported.${NC}"
            exit 1
            ;;
        *)
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}Detected platform: $PLATFORM${NC}"
}

# Download and install binary
install_binary() {
    # For now, use GitHub releases URL pattern
    # TODO: Replace with actual release URL when published
    DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY_NAME}-${PLATFORM}"

    echo "Downloading CKS Weight Room from $DOWNLOAD_URL..."

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_NAME"

    # Download binary
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    else
        echo -e "${RED}Error: Neither curl nor wget found. Please install one and try again.${NC}"
        exit 1
    fi

    # Make binary executable
    chmod +x "$TMP_FILE"

    # Install to system directory
    echo "Installing to $INSTALL_DIR..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        echo -e "${YELLOW}Note: Installing to $INSTALL_DIR requires sudo permissions${NC}"
        sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Clean up
    rm -rf "$TMP_DIR"

    echo -e "${GREEN}Installation complete!${NC}"
}

# Display version
show_version() {
    VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>&1 || echo "unknown")
    echo ""
    echo "$VERSION"
    echo ""
    echo -e "${GREEN}To start CKS Weight Room, run: $BINARY_NAME${NC}"
}

# Main installation flow
main() {
    echo ""
    echo "====================================="
    echo "  CKS Weight Room Installer"
    echo "====================================="
    echo ""

    detect_platform
    install_binary
    show_version
}

main
