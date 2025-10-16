#!/bin/sh
# agentmlx installer
# Installs the latest version of agentmlx from Cloudflare R2
# Usage: curl -fsSL sh.agentml.dev | sh

set -e

# Configuration
RELEASES_BASE_URL="${AGENTMLX_RELEASES_URL:-https://amlx.agentml.dev/agentmlx}"
INSTALL_DIR="${AGENTMLX_INSTALL_DIR:-$HOME/.agentmlx}"
BIN_DIR="$INSTALL_DIR/bin"

# Colors for output
if [ -t 1 ]; then
    BOLD='\033[1m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    RED='\033[0;31m'
    RESET='\033[0m'
else
    BOLD=''
    GREEN=''
    YELLOW=''
    RED=''
    RESET=''
fi

# Helper functions
info() {
    printf "${GREEN}==>${RESET} ${BOLD}%s${RESET}\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${RESET} %s\n" "$1" >&2
}

error() {
    printf "${RED}Error:${RESET} %s\n" "$1" >&2
    exit 1
}

# Detect platform
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac

    # Special handling for macOS on Rosetta 2
    if [ "$os" = "darwin" ] && [ "$arch" = "amd64" ]; then
        if [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)" = "1" ]; then
            arch="arm64"
            info "Detected macOS on Apple Silicon (Rosetta 2), using arm64 binary"
        fi
    fi

    echo "${os}_${arch}"
}

# Check if a binary exists at the given URL
check_binary_exists() {
    local url="$1"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --head "$url" >/dev/null 2>&1
    elif command -v wget >/dev/null 2>&1; then
        wget -q --spider "$url" 2>&1
    else
        error "Neither curl nor wget found. Please install one and try again."
    fi
}

# Try to get binary URL for a channel with fallback
get_binary_url_with_fallback() {
    local channel="$1"
    local platform="$2"
    local binary_url
    local checksums_url

    # Try requested channel
    info "Checking ${channel} channel..." >&2
    binary_url="${RELEASES_BASE_URL}/${channel}/agentmlx_${platform}"
    checksums_url="${RELEASES_BASE_URL}/${channel}/checksums.txt"

    if check_binary_exists "$binary_url"; then
        echo "${binary_url}|${checksums_url}"
        return 0
    fi

    # Waterfall logic: latest → next → beta
    case "$channel" in
        latest)
            warn "No binary found in latest channel, trying next (rc) channel..."
            binary_url="${RELEASES_BASE_URL}/next/agentmlx_${platform}"
            checksums_url="${RELEASES_BASE_URL}/next/checksums.txt"
            if check_binary_exists "$binary_url"; then
                warn "Using next channel"
                echo "${binary_url}|${checksums_url}"
                return 0
            fi

            warn "No binary found in next channel, trying beta channel..."
            binary_url="${RELEASES_BASE_URL}/beta/agentmlx_${platform}"
            checksums_url="${RELEASES_BASE_URL}/beta/checksums.txt"
            if check_binary_exists "$binary_url"; then
                warn "Using beta channel"
                echo "${binary_url}|${checksums_url}"
                return 0
            fi
            ;;
        next)
            warn "No binary found in next channel, trying beta channel..."
            binary_url="${RELEASES_BASE_URL}/beta/agentmlx_${platform}"
            checksums_url="${RELEASES_BASE_URL}/beta/checksums.txt"
            if check_binary_exists "$binary_url"; then
                warn "Using beta channel"
                echo "${binary_url}|${checksums_url}"
                return 0
            fi
            ;;
    esac

    # Nothing found
    return 1
}

# Download file
download() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one and try again."
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksums_file="$2"
    local binary_name="$3"

    if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
        warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        return 0
    fi

    local expected_sum
    expected_sum=$(grep "$binary_name" "$checksums_file" | awk '{print $1}')

    if [ -z "$expected_sum" ]; then
        warn "Could not find checksum for $binary_name. Skipping verification."
        return 0
    fi

    local actual_sum
    if command -v sha256sum >/dev/null 2>&1; then
        actual_sum=$(sha256sum "$file" | awk '{print $1}')
    else
        actual_sum=$(shasum -a 256 "$file" | awk '{print $1}')
    fi

    if [ "$expected_sum" != "$actual_sum" ]; then
        error "Checksum verification failed!\nExpected: $expected_sum\nActual:   $actual_sum"
    fi

    info "Checksum verified successfully"
}

# Add to PATH
setup_path() {
    local shell_profile=""

    # Detect shell and corresponding profile
    case "$SHELL" in
        */bash)
            if [ -f "$HOME/.bashrc" ]; then
                shell_profile="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                shell_profile="$HOME/.bash_profile"
            fi
            ;;
        */zsh)
            shell_profile="$HOME/.zshrc"
            ;;
        */fish)
            shell_profile="$HOME/.config/fish/config.fish"
            ;;
    esac

    if [ -z "$shell_profile" ]; then
        warn "Could not detect shell profile. Please add $BIN_DIR to your PATH manually."
        return
    fi

    # Check if already in PATH
    if echo "$PATH" | grep -q "$BIN_DIR"; then
        return
    fi

    # Check if already in profile
    if grep -q "$BIN_DIR" "$shell_profile" 2>/dev/null; then
        return
    fi

    # Add to profile
    info "Adding $BIN_DIR to PATH in $shell_profile"
    echo "" >> "$shell_profile"
    echo "# agentmlx" >> "$shell_profile"
    echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$shell_profile"

    info "Please run: ${BOLD}source $shell_profile${RESET} or restart your shell"
}

# Main installation
main() {
    local version="${AGENTMLX_VERSION:-}"
    local channel="${AGENTMLX_CHANNEL:-latest}"
    local no_modify_path=0

    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                version="$2"
                shift 2
                ;;
            --channel)
                channel="$2"
                if [ "$channel" != "latest" ] && [ "$channel" != "next" ] && [ "$channel" != "beta" ]; then
                    error "Invalid channel: $channel\nValid channels: latest, next, beta"
                fi
                shift 2
                ;;
            --no-modify-path)
                no_modify_path=1
                shift
                ;;
            --help)
                echo "agentmlx installer"
                echo ""
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --version <version>    Install specific version (without 'v' prefix)"
                echo "  --channel <channel>    Install from channel: latest, next, beta (default: latest)"
                echo "  --no-modify-path       Don't modify shell profile"
                echo "  --help                 Show this help message"
                echo ""
                echo "Channels:"
                echo "  latest                 Stable releases (v1.0.0)"
                echo "  next                   Release candidates (v1.0.0-rc.1)"
                echo "  beta                   Beta releases (v1.0.0-beta.1)"
                echo ""
                echo "Channel Fallback:"
                echo "  If a channel has no releases, it automatically falls back:"
                echo "  latest → next → beta"
                echo ""
                echo "Environment variables:"
                echo "  AGENTMLX_VERSION        Version to install (without 'v' prefix)"
                echo "  AGENTMLX_CHANNEL        Channel to install from (latest, next, beta)"
                echo "  AGENTMLX_INSTALL_DIR    Installation directory (default: \$HOME/.agentmlx)"
                echo "  AGENTMLX_RELEASES_URL   Base URL for releases (default: https://amlx.agentml.dev/agentmlx)"
                echo ""
                echo "Examples:"
                echo "  $0                           # Install latest stable"
                echo "  $0 --channel next            # Install latest RC"
                echo "  $0 --version 1.0.0-rc.1      # Install specific version"
                exit 0
                ;;
            *)
                error "Unknown option: $1\nRun with --help for usage information."
                ;;
        esac
    done

    info "Installing agentmlx..."

    # Detect platform
    local platform
    platform=$(detect_platform)
    info "Detected platform: $platform"

    # Get binary URLs (with channel fallback logic)
    local binary_url checksums_url
    if [ -n "$version" ]; then
        # Specific version requested
        info "Version: $version"
        local binary_name="agentmlx_${version}_${platform}"
        binary_url="${RELEASES_BASE_URL}/v${version}/${binary_name}"
        checksums_url="${RELEASES_BASE_URL}/v${version}/checksums.txt"
    else
        # Use channel
        local urls
        urls=$(get_binary_url_with_fallback "$channel" "$platform")
        if [ -z "$urls" ]; then
            error "Failed to find any releases for channel: $channel"
        fi
        binary_url=$(echo "$urls" | cut -d'|' -f1)
        checksums_url=$(echo "$urls" | cut -d'|' -f2)
    fi

    # Create temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf '$tmp_dir'" EXIT

    # Download binary
    info "Downloading agentmlx..."
    download "$binary_url" "$tmp_dir/agentmlx" || error "Failed to download binary from $binary_url"

    # Download checksums
    info "Downloading checksums..."
    download "$checksums_url" "$tmp_dir/checksums.txt" || warn "Failed to download checksums"

    # Verify checksum if checksums file was downloaded
    if [ -f "$tmp_dir/checksums.txt" ]; then
        info "Verifying checksum..."
        local binary_name
        binary_name=$(basename "$binary_url")
        verify_checksum "$tmp_dir/agentmlx" "$tmp_dir/checksums.txt" "$binary_name"
    fi

    # Create installation directory
    mkdir -p "$BIN_DIR"

    # Install binary
    info "Installing to $BIN_DIR/agentmlx..."
    mv "$tmp_dir/agentmlx" "$BIN_DIR/agentmlx"
    chmod +x "$BIN_DIR/agentmlx"

    # Create amlx symlink
    info "Creating amlx alias..."
    ln -sf "$BIN_DIR/agentmlx" "$BIN_DIR/amlx"

    # Setup PATH
    if [ $no_modify_path -eq 0 ]; then
        setup_path
    fi

    # Show installed version
    local installed_version
    if [ -n "$version" ]; then
        info "${GREEN}✓${RESET} agentmlx v${version} installed successfully!"
    else
        # Try to get version from binary
        installed_version=$("$BIN_DIR/agentmlx" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+(-[a-z]+\.[0-9]+)?' || echo "")
        if [ -n "$installed_version" ]; then
            info "${GREEN}✓${RESET} agentmlx v${installed_version} installed successfully!"
        else
            info "${GREEN}✓${RESET} agentmlx installed successfully!"
        fi
    fi

    echo ""
    echo "To get started, run:"
    echo "  ${BOLD}agentmlx --help${RESET}"
    echo "  ${BOLD}amlx --help${RESET}"
    echo ""

    if [ $no_modify_path -eq 0 ] && ! echo "$PATH" | grep -q "$BIN_DIR"; then
        echo "Note: You may need to restart your shell or run:"
        echo "  ${BOLD}export PATH=\"\$PATH:$BIN_DIR\"${RESET}"
    fi
}

main "$@"
