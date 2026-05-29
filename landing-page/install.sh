#!/bin/bash
# ==============================================================================
# GCAM One-Click Installer (macOS / Linux)
# Usage: curl -sSL https://gcam.dong4j.site/install.sh | bash
# Or specify version: curl -sSL https://gcam.dong4j.site/install.sh | bash -s -- v1.1.0
# ==============================================================================

set -e

# Configuration
REPO="dong4j/gemini-cli-account-manager"
BINARY_NAME="gcam"
INSTALL_DIR="${HOME}/.local/bin"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Detect OS and architecture
detect_os() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Darwin*)
            OS="darwin"
            ;;
        Linux*)
            OS="linux"
            ;;
        *)
            error "Unsupported OS: $OS"
            error "Windows users please use install.bat: https://gcam.dong4j.site/install.bat"
            exit 1
            ;;
    esac

    # Normalize architecture names
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    info "Detected platform: ${OS}-${ARCH}"
}

# Get latest version
get_latest_version() {
    if [ -n "$1" ]; then
        VERSION="$1"
        return
    fi

    info "Fetching latest version..."

    # Use GitHub token if available (raises API rate limit from 60 to 5000 per hour)
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    if [ -n "$GITHUB_TOKEN" ]; then
        VERSION=$(curl -sSL -H "Authorization: token $GITHUB_TOKEN" "$api_url" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    else
        VERSION=$(curl -sSL "$api_url" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version. Please check your network or specify a version."
        error "Tip: export GITHUB_TOKEN to avoid rate limits"
        exit 1
    fi

    info "Latest version: $VERSION"
}

# Download binary (returns file path only)
download_binary() {
    local version="$1"
    local filename="${BINARY_NAME}-${OS}-${ARCH}"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"

    info "Download URL: $download_url"

    # Create temp file
    local tmpfile="/tmp/gcam_install_${filename}"
    rm -f "$tmpfile"

    info "Downloading..."
    if ! curl -sSL -o "$tmpfile" "$download_url"; then
        error "Download failed. Please verify the version: $version"
        rm -f "$tmpfile"
        exit 1
    fi

    # Verify file
    if [ ! -s "$tmpfile" ]; then
        error "Downloaded file is empty or corrupted"
        rm -f "$tmpfile"
        exit 1
    fi

    # Check file size
    local size=$(wc -c < "$tmpfile" 2>/dev/null || echo "0")
    if [ "$size" -lt 1000 ]; then
        if file "$tmpfile" 2>/dev/null | grep -q "HTML"; then
            error "Downloaded file appears to be an HTML error page. Version may not exist: $version"
            rm -f "$tmpfile"
            exit 1
        fi
    fi

    # Return file path to stdout
    echo "$tmpfile"
}

# Install binary to target directory
install_binary() {
    local src="$1"

    # Create install directory
    if [ ! -d "$INSTALL_DIR" ]; then
        info "Creating install directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi

    local target="${INSTALL_DIR}/${BINARY_NAME}"

    # Move file
    info "Installing to: $target"
    mv "$src" "$target"
    chmod +x "$target"

    # Cleanup temp file
    rm -f "$src"

    success "Installation complete!"
    echo
    echo "=========================================="
    echo "  GCAM installed to:"
    echo "  $target"
    echo "=========================================="
    echo
}

# Check and prompt for PATH configuration
check_path() {
    # Check if PATH already contains install directory
    if echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
        info "PATH is correctly configured"
        return 0
    fi

    warn "Need to add $INSTALL_DIR to PATH"
    echo
    echo "Add the following to your shell configuration file:"
    echo
    echo -e "  ${GREEN}# GCAM${NC}"
    echo -e "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo

    # Detect shell type
    local shell_config=""
    case "$(basename "$SHELL")" in
        zsh)
            shell_config="${HOME}/.zshrc"
            ;;
        bash)
            if [ "$(uname -s)" = "Darwin" ]; then
                shell_config="${HOME}/.bash_profile"
            else
                shell_config="${HOME}/.bashrc"
            fi
            ;;
        fish)
            shell_config="${HOME}/.config/fish/config.fish"
            ;;
        *)
            shell_config="${HOME}/.profile"
            ;;
    esac

    echo "Config file: $shell_config"
    echo
    echo "After adding, run the following command to apply:"
    echo
    echo -e "  ${GREEN}source $shell_config${NC}"
    echo
    echo "Or restart your terminal."
}

# Show usage guide
show_usage() {
    echo
    echo "=========================================="
    echo "  Common Commands"
    echo "=========================================="
    echo
    echo -e "  ${GREEN}gcam${NC}               List accounts"
    echo -e "  ${GREEN}gcam 1${NC}            Switch to account #1"
    echo -e "  ${GREEN}gcam next${NC}          Switch to next account"
    echo -e "  ${GREEN}gcam quota${NC}         Check quota usage"
    echo -e "  ${GREEN}gcam pool login${NC}   Add new account"
    echo -e "  ${GREEN}gcam menu${NC}         Open interactive menu"
    echo
    echo "  Install hooks (enable /gcam command):"
    echo -e "  ${GREEN}gcam install${NC}"
    echo
    echo "  View help:"
    echo -e "  ${GREEN}gcam --help${NC}"
    echo
    echo "=========================================="
    echo
    echo -e "  Docs: ${BLUE}https://gcam.dong4j.site${NC}"
    echo -e "  GitHub: ${BLUE}https://github.com/${REPO}${NC}"
    echo
}

# Main function
main() {
    echo
    echo "=========================================="
    echo "  GCAM One-Click Installer"
    echo "  https://github.com/${REPO}"
    echo "=========================================="
    echo

    # Detect OS
    detect_os

    # Get version
    get_latest_version "$1"

    # Download
    BINARY_PATH=$(download_binary "$VERSION")

    # Install
    install_binary "$BINARY_PATH"

    # Check PATH
    check_path

    # Show usage guide
    show_usage
}

# Run
main "$@"
