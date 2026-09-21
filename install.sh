#!/usr/bin/env bash
#
# Knowblazer installer script
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash
#   or specify a version:
#   curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash -s -- v0.1.0
#
set -euo pipefail

REPO="KiBlazer/knowblazer"
BINARY_NAME="knowblazer"

# Colors for output
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    BLUE='\033[0;34m'
    YELLOW='\033[1;33m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    BLUE=''
    YELLOW=''
    BOLD=''
    NC=''
fi

info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}==>${NC} ${BOLD}%s${NC}\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
OS_RAW="$(uname -s)"
case "$OS_RAW" in
    Linux*)   OS="linux" ;;
    Darwin*)  OS="darwin" ;;
    CYGWIN*|MINGW*|MSYS*) OS="windows" ;;
    *) error "Unsupported operating system: $OS_RAW" ;;
esac

# Detect Architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH_RAW" ;;
esac

# Determine requested or latest version
VERSION="${VERSION:-${1:-}}"
if [ -z "$VERSION" ]; then
    info "Checking latest release of knowblazer from GitHub..."
    # Try GitHub API first
    TAG="$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)"
    
    # Fallback to HTTP redirect if API is rate-limited
    if [ -z "$TAG" ]; then
        TAG="$(curl -sSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null | sed 's#.*/##' || true)"
    fi

    if [ -z "$TAG" ] || [ "$TAG" = "latest" ]; then
        error "Could not determine latest release. Please specify version, e.g.: bash install.sh v0.1.0"
    fi
else
    TAG="$VERSION"
fi

# Ensure tag has 'v' prefix for release download, clean version without 'v' for archive name
if [[ "$TAG" != v* ]]; then
    TAG="v$TAG"
fi
CLEAN_VERSION="${TAG#v}"

# Determine archive format and asset name
if [ "$OS" = "windows" ]; then
    EXT="zip"
    EXE="${BINARY_NAME}.exe"
else
    EXT="tar.gz"
    EXE="${BINARY_NAME}"
fi
ASSET_NAME="knowblazer_${CLEAN_VERSION}_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET_NAME}"

info "Detected platform: ${OS}/${ARCH}"
info "Downloading knowblazer ${TAG} (${ASSET_NAME})..."

TMP_DIR="$(mktemp -d -t knowblazer-install.XXXXXX)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

ARCHIVE_PATH="${TMP_DIR}/${ASSET_NAME}"
if ! curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_PATH"; then
    error "Download failed from $DOWNLOAD_URL. Please verify version or network connection."
fi

# Extract binary
info "Extracting ${BINARY_NAME}..."
if [ "$EXT" = "zip" ]; then
    unzip -q "$ARCHIVE_PATH" -d "$TMP_DIR"
else
    tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
fi

if [ ! -f "${TMP_DIR}/${EXE}" ]; then
    error "Extracted archive did not contain expected binary: ${EXE}"
fi
chmod +x "${TMP_DIR}/${EXE}"

# Determine installation directory
if [ -n "${INSTALL_DIR:-}" ]; then
    TARGET_DIR="$INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
    TARGET_DIR="/usr/local/bin"
elif [ -d "$HOME/.local/bin" ] || [ "$(id -u)" != "0" ]; then
    TARGET_DIR="$HOME/.local/bin"
else
    TARGET_DIR="/usr/local/bin"
fi

mkdir -p "$TARGET_DIR"

TARGET_BIN="${TARGET_DIR}/${EXE}"
info "Installing to ${TARGET_BIN}..."

if [ -w "$TARGET_DIR" ]; then
    mv "${TMP_DIR}/${EXE}" "$TARGET_BIN"
else
    if command -v sudo >/dev/null 2>&1; then
        warn "Elevated permissions required to write to ${TARGET_DIR}."
        sudo mv "${TMP_DIR}/${EXE}" "$TARGET_BIN"
    else
        error "Directory ${TARGET_DIR} is not writable and sudo is not available."
    fi
fi

chmod 755 "$TARGET_BIN"

success "knowblazer ${TAG} installed successfully!"

# Check if TARGET_DIR is in PATH
if [[ ":$PATH:" != *":$TARGET_DIR:"* ]]; then
    warn "${TARGET_DIR} is not in your PATH."
    printf "\nTo make knowblazer accessible everywhere, add this to your ~/.bashrc or ~/.zshrc:\n"
    printf "  ${BOLD}export PATH=\"%s:\$PATH\"${NC}\n\n" "$TARGET_DIR"
fi

printf "\nRun '${BOLD}knowblazer start${NC}' to automatically connect installed AI tools (Claude, Codex, Gemini).\n"
