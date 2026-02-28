#!/usr/bin/env bash
set -e

# GitHub Copilot Session Delete - Installer
# Automatically detects OS/architecture, downloads latest release, verifies checksums,
# and installs to ~/.local/bin

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="salmanabdurrahman"
REPO_NAME="copilot-session-delete"
BINARY_NAME="copilot-session-delete"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Print colored message
info() { echo -e "${BLUE}${1}${NC}"; }
success() { echo -e "${GREEN}✓${NC} ${1}"; }
warn() { echo -e "${YELLOW}Warning: ${1}${NC}"; }
error() { echo -e "${RED}Error: ${1}${NC}" >&2; }

# Main function
main() {
    echo -e "${BLUE}GitHub Copilot Session Delete - Installer${NC}"
    echo ""

    detect_platform
    check_dependencies
    fetch_latest_release
    download_and_verify
    install_binary
    cleanup
    show_completion_message
}

# Detect operating system
detect_platform() {
    local os arch

    os="$(uname -s)"
    case "$os" in
        Linux*)     OS_TYPE="Linux";;
        Darwin*)    OS_TYPE="Darwin";;
        *)
            error "Unsupported operating system: $os"
            echo "Supported: Linux, macOS" >&2
            exit 1
            ;;
    esac

    arch="$(uname -m)"
    case "$arch" in
        x86_64)     ARCH_TYPE="x86_64";;
        aarch64)    ARCH_TYPE="arm64";;
        arm64)      ARCH_TYPE="arm64";;
        *)
            error "Unsupported architecture: $arch"
            echo "Supported: x86_64, arm64, aarch64" >&2
            exit 1
            ;;
    esac

    PLATFORM="${OS_TYPE}_${ARCH_TYPE}"
    success "Detected platform: ${PLATFORM}"
}

# Check for required dependencies
check_dependencies() {
    if command -v curl &> /dev/null; then
        DOWNLOADER="curl"
    elif command -v wget &> /dev/null; then
        DOWNLOADER="wget"
    else
        error "Neither curl nor wget is installed"
        echo "Please install one of them and try again" >&2
        exit 1
    fi

    # Check for checksum tool
    if command -v sha256sum &> /dev/null; then
        CHECKSUM_CMD="sha256sum"
    elif command -v shasum &> /dev/null; then
        CHECKSUM_CMD="shasum -a 256"
    else
        warn "No checksum tool found (sha256sum/shasum)"
        CHECKSUM_CMD=""
    fi
}

# Download file using curl or wget
download() {
    local url=$1
    local output=$2

    if [ "$DOWNLOADER" = "curl" ]; then
        curl -fsSL "$url" -o "$output" 2>&1
    else
        wget -q -O "$output" "$url" 2>&1
    fi
}

# Fetch latest release version from GitHub
fetch_latest_release() {
    info "Fetching latest release..."
    
    LATEST_RELEASE=$(download "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" - | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

    if [ -z "$LATEST_RELEASE" ]; then
        error "Could not fetch latest release"
        echo "Please check your internet connection or try again later" >&2
        exit 1
    fi

    success "Latest version: ${LATEST_RELEASE}"
}

# Download and verify the binary archive
download_and_verify() {
    ARCHIVE_FILENAME="${REPO_NAME}_${PLATFORM}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_RELEASE/$ARCHIVE_FILENAME"
    CHECKSUMS_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_RELEASE/checksums.txt"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR" || exit 1

    # Download archive
    info "Downloading ${ARCHIVE_FILENAME}..."
    if ! download "$DOWNLOAD_URL" "$ARCHIVE_FILENAME"; then
        error "Failed to download archive"
        echo "URL: $DOWNLOAD_URL" >&2
        rm -rf "$TMP_DIR"
        exit 1
    fi
    success "Downloaded"

    # Download checksums file
    info "Verifying checksum..."
    if ! download "$CHECKSUMS_URL" "checksums.txt"; then
        warn "Could not download checksums file"
        echo "Skipping checksum verification"
    else
        verify_checksum
    fi

    # Extract archive
    info "Extracting..."
    if ! tar -xzf "$ARCHIVE_FILENAME" 2>/dev/null; then
        error "Failed to extract archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi
    success "Extracted"

    # Verify binary exists after extraction
    if [ ! -f "$BINARY_NAME" ]; then
        error "Binary not found in archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi
}

# Verify file checksum
verify_checksum() {
    if [ ! -f "checksums.txt" ]; then
        return
    fi

    local expected actual

    expected=$(grep "$ARCHIVE_FILENAME" checksums.txt | awk '{print $1}')

    if [ -z "$expected" ]; then
        warn "Checksum not found for $ARCHIVE_FILENAME"
        return
    fi

    if [ -z "$CHECKSUM_CMD" ]; then
        warn "No checksum tool available"
        return
    fi

    actual=$($CHECKSUM_CMD "$ARCHIVE_FILENAME" | awk '{print $1}')

    if [ "$expected" = "$actual" ]; then
        success "Checksum verified"
    else
        error "Checksum mismatch!"
        echo "Expected: $expected" >&2
        echo "Got:      $actual" >&2
        rm -rf "$TMP_DIR"
        exit 1
    fi
}

# Install binary to target directory
install_binary() {
    info "Installing to ${INSTALL_DIR}..."

    # Create install directory if it doesn't exist
    if ! mkdir -p "$INSTALL_DIR"; then
        error "Could not create directory: $INSTALL_DIR"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Move binary to install directory
    if ! mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"; then
        error "Could not move binary to $INSTALL_DIR"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make binary executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Clean up temporary files
cleanup() {
    cd ~ || exit 1
    rm -rf "$TMP_DIR"
}

# Show completion message with PATH instructions
show_completion_message() {
    echo ""
    success "Installation complete!"
    echo ""

    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add this line to your shell config (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo -e "  ${BLUE}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        echo ""
        echo "Then restart your shell or run:"
        echo ""
        echo -e "  ${BLUE}source ~/.bashrc${NC}  # or ~/.zshrc"
        echo ""
    else
        echo "Run the tool with:"
        echo ""
        echo -e "  ${BLUE}$BINARY_NAME${NC}"
        echo ""
        echo "To see all available commands:"
        echo ""
        echo -e "  ${BLUE}$BINARY_NAME --help${NC}"
        echo ""
    fi
}

# Run main function
main
