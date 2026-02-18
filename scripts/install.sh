#!/usr/bin/env bash
#
# Claude Switch Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/sumanta-mukhopadhyay/claude-switch/main/scripts/install.sh | bash
#

set -euo pipefail

REPO="sumanta-mukhopadhyay/claude-switch"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="claude-switch"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info()  { echo -e "${BLUE}ℹ${NC}  $*"; }
success() { echo -e "${GREEN}✓${NC}  $*"; }
warn()  { echo -e "${YELLOW}⚠${NC}  $*"; }
error() { echo -e "${RED}✗${NC}  $*" >&2; }

# Detect OS and architecture
detect_platform() {
    local os arch

    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        linux)  os="linux" ;;
        darwin) os="darwin" ;;
        *)      error "Unsupported OS: $os"; exit 1 ;;
    esac

    case "$arch" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $arch"; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Could not determine latest version"
        exit 1
    fi
    echo "$version"
}

main() {
    echo ""
    echo "  Claude Switch Installer"
    echo "  ─────────────────────────"
    echo ""

    # Check for required tools
    for tool in curl; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is required but not installed"
            exit 1
        fi
    done

    local platform version download_url tmp_dir

    platform="$(detect_platform)"
    info "Detected platform: ${platform}"

    version="$(get_latest_version)"
    info "Latest version: ${version}"

    # Construct download URL
    local version_no_v="${version#v}"
    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}_${version_no_v}_${platform}.tar.gz"
    info "Downloading: ${download_url}"

    # Download and extract
    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' EXIT

    curl -fsSL "$download_url" -o "${tmp_dir}/archive.tar.gz"
    tar -xzf "${tmp_dir}/archive.tar.gz" -C "$tmp_dir"

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        info "Requesting sudo access to install to ${INSTALL_DIR}"
        sudo mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    success "Installed ${BINARY_NAME} ${version} to ${INSTALL_DIR}/${BINARY_NAME}"
    echo ""

    # Create alias hint
    if ! command -v cs &> /dev/null; then
        warn "Consider adding an alias for convenience:"
        echo "    echo 'alias cs=claude-switch' >> ~/.bashrc"
        echo ""
    fi

    # Verify
    if command -v "$BINARY_NAME" &> /dev/null; then
        "${BINARY_NAME}" version
    fi

    echo ""
    success "Installation complete! Run 'claude-switch --help' to get started."
}

main
