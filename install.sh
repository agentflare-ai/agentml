#!/bin/sh
# agentmlx installer
# Installs the latest version of agentmlx from Cloudflare R2
# Usage: curl -fsSL https://raw.githubusercontent.com/agentflare-ai/agentml/main/install.sh | sh

set -e

# Configuration
RELEASES_BASE_URL="${AGENTMLX_RELEASES_URL:-https://amlx.agentml.dev/agentmlx}"
GITHUB_RELEASES_API="https://api.github.com/repos/agentflare-ai/agentmlx/releases"
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

# Fetch all releases and extract versions
fetch_all_releases() {
    local releases_json

    if command -v curl >/dev/null 2>&1; then
        releases_json=$(curl -sfL "${GITHUB_RELEASES_API}?per_page=100")
    elif command -v wget >/dev/null 2>&1; then
        releases_json=$(wget -qO- "${GITHUB_RELEASES_API}?per_page=100")
    else
        error "Neither curl nor wget found. Please install one and try again."
    fi

    # Extract tag names (removes "v" prefix)
    echo "$releases_json" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'
}

# Get version for specific channel
get_version_for_channel() {
    local channel="$1"
    local versions

    versions=$(fetch_all_releases)

    case "$channel" in
        latest)
            # Stable versions: v1.0.0 (no suffix)
            echo "$versions" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | head -n1
            ;;
        next)
            # RC versions: v1.0.0-rc.1
            echo "$versions" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+$' | head -n1
            ;;
        beta)
            # Beta versions: v1.0.0-beta.1
            echo "$versions" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+-beta\.[0-9]+$' | head -n1
            ;;
        *)
            error "Unknown channel: $channel"
            ;;
    esac
}

# Get version with waterfall fallback
get_version_with_fallback() {
    local channel="$1"
    local version

    info "Fetching ${channel} version from GitHub..."

    # Try requested channel
    version=$(get_version_for_channel "$channel")
    if [ -n "$version" ]; then
        echo "$version"
        return 0
    fi

    # Waterfall logic: latest → next → beta
    case "$channel" in
        latest)
            warn "No stable releases found, trying next (rc) channel..."
            version=$(get_version_for_channel "next")
            if [ -n "$version" ]; then
                warn "Using next channel: v${version}"
                echo "$version"
                return 0
            fi

            warn "No rc releases found, trying beta channel..."
            version=$(get_version_for_channel "beta")
            if [ -n "$version" ]; then
                warn "Using beta channel: v${version}"
                echo "$version"
                return 0
            fi
            ;;
        next)
            warn "No rc releases found, trying beta channel..."
            version=$(get_version_for_channel "beta")
            if [ -n "$version" ]; then
                warn "Using beta channel: v${version}"
                echo "$version"
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

    # Get version
    if [ -z "$version" ]; then
        version=$(get_version_with_fallback "$channel")
        if [ -z "$version" ]; then
            error "Failed to find any releases for channel: $channel"
        fi
    fi
    info "Version: $version"

    # Construct download URLs from R2
    local binary_name="agentmlx_${version}_${platform}"
    local binary_url="${RELEASES_BASE_URL}/v${version}/${binary_name}"
    local checksums_url="${RELEASES_BASE_URL}/v${version}/checksums.txt"

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

    info "${GREEN}✓${RESET} agentmlx v${version} installed successfully!"
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
