#!/usr/bin/env bash
#
# Claude Switch Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/caeser1996/claude-switch/main/scripts/install.sh | bash
#

set -euo pipefail

REPO="caeser1996/claude-switch"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="claude-switch"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info()    { echo -e "${BLUE}ℹ${NC}  $*"; }
success() { echo -e "${GREEN}✓${NC}  $*"; }
warn()    { echo -e "${YELLOW}⚠${NC}  $*"; }
error()   { echo -e "${RED}✗${NC}  $*" >&2; }

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
    for tool in curl tar; do
        if ! command -v "$tool" &> /dev/null; then
            error "$tool is required but not installed"
            exit 1
        fi
    done

    local platform version tmp_dir
    local base_url archive_url binary_url

    platform="$(detect_platform)"
    info "Detected platform: ${platform}"

    version="$(get_latest_version)"
    info "Latest version: ${version}"

    # Strip leading "v" for archive filename (GoReleaser uses version without v)
    local version_no_v="${version#v}"
    base_url="https://github.com/${REPO}/releases/download/${version}"

    # GoReleaser produces two asset types:
    #   1. tar.gz archive:  claude-switch_0.3.0_darwin_arm64.tar.gz
    #   2. raw binary:      claude-switch_darwin_arm64
    archive_url="${base_url}/${BINARY_NAME}_${version_no_v}_${platform}.tar.gz"
    binary_url="${base_url}/${BINARY_NAME}_${platform}"

    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' EXIT

    # Try tar.gz archive first, fall back to raw binary
    if curl -fsSL "$archive_url" -o "${tmp_dir}/archive.tar.gz" 2>/dev/null; then
        info "Downloading archive: ${archive_url}"
        tar -xzf "${tmp_dir}/archive.tar.gz" -C "$tmp_dir"
    elif curl -fsSL "$binary_url" -o "${tmp_dir}/${BINARY_NAME}" 2>/dev/null; then
        info "Downloading binary: ${binary_url}"
    else
        error "Failed to download claude-switch ${version} for ${platform}"
        error "Tried:"
        error "  ${archive_url}"
        error "  ${binary_url}"
        error ""
        error "Check available assets at:"
        error "  https://github.com/${REPO}/releases/tag/${version}"
        exit 1
    fi

    # Install binary
    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        info "Requesting sudo access to install to ${INSTALL_DIR}"
        sudo mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    success "Installed ${BINARY_NAME} ${version} to ${INSTALL_DIR}/${BINARY_NAME}"

    # Create cs symlink
    if [ -w "$INSTALL_DIR" ]; then
        ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/cs"
    else
        sudo ln -sf "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/cs"
    fi
    success "Created shortcut: cs → ${BINARY_NAME}"

    echo ""

    # Verify
    if command -v "$BINARY_NAME" &> /dev/null; then
        "${BINARY_NAME}" version
    fi

    echo ""
    success "Done! Run 'cs --help' to get started."
}

main
