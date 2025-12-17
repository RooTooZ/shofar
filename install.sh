#!/bin/sh
set -e

# Shofar Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/RooTooZ/shofar/main/install.sh | sh

REPO="RooTooZ/shofar"
INSTALL_DIR="$HOME/.local/bin"
DATA_DIR="$HOME/.local/share/shofar"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) error "macOS: Pre-built binaries not available. Please build from source: make all" ;;
        MINGW*|MSYS*|CYGWIN*) error "Windows: Pre-built binaries not available. Please build from source." ;;
        *) error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/'
}

# Download and install
install_shofar() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected: ${OS}/${ARCH}"

    # Get version
    if [ -n "$VERSION" ]; then
        info "Installing version: $VERSION"
    else
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            error "Failed to get latest version. Set VERSION env var manually."
        fi
        info "Latest version: $VERSION"
    fi

    # Construct download URL
    FILENAME="shofar-${VERSION}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        FILENAME="${FILENAME}.zip"
    else
        FILENAME="${FILENAME}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

    info "Downloading from: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    if command -v curl > /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME"
    elif command -v wget > /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$FILENAME"
    else
        error "curl or wget required"
    fi

    # Extract
    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$FILENAME"
        BINARY="shofar.exe"
    else
        tar -xzf "$FILENAME"
        BINARY="shofar-${OS}-${ARCH}"
    fi

    # Install
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$DATA_DIR/models/whisper"
    mkdir -p "$DATA_DIR/models/vosk"
    mkdir -p "$DATA_DIR/models/llm"

    if [ "$OS" = "windows" ]; then
        mv "$BINARY" "$INSTALL_DIR/shofar.exe"
        info "Installed to: $INSTALL_DIR/shofar.exe"
    else
        mv "$BINARY" "$INSTALL_DIR/shofar"
        chmod +x "$INSTALL_DIR/shofar"
        info "Installed to: $INSTALL_DIR/shofar"
    fi

    # Check if in PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        warn "$INSTALL_DIR is not in PATH"
        echo ""
        echo "Add to your shell config:"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
    fi

    info "Installation complete!"
    echo ""
    echo "Run 'shofar' to start the application."
    echo "Models will be downloaded on first use via Settings."
}

# Check dependencies
check_deps() {
    OS=$(detect_os)

    if [ "$OS" = "linux" ]; then
        MISSING=""

        # Check for portaudio
        if ! ldconfig -p 2>/dev/null | grep -q libportaudio; then
            MISSING="$MISSING portaudio"
        fi

        # Check for gtk3
        if ! ldconfig -p 2>/dev/null | grep -q libgtk-3; then
            MISSING="$MISSING gtk3"
        fi

        # Check for appindicator
        if ! ldconfig -p 2>/dev/null | grep -q libayatana-appindicator; then
            MISSING="$MISSING libayatana-appindicator"
        fi

        if [ -n "$MISSING" ]; then
            warn "Missing dependencies:$MISSING"
            echo ""
            echo "Install with:"
            echo "  Arch:   sudo pacman -S portaudio libayatana-appindicator gtk3 xdotool"
            echo "  Ubuntu: sudo apt install libportaudio2 libayatana-appindicator3-1 libgtk-3-0 xdotool"
            echo ""
        fi
    fi
}

# Main
main() {
    echo ""
    echo "  Shofar Installer"
    echo "  Voice-to-Text for Desktop"
    echo ""

    check_deps
    install_shofar
}

main "$@"
