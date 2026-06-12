#!/bin/sh
set -e

# Knowblazer Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | sh

OWNER="KiBlazer"
REPO="knowblazer"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
    linux*)   OS="linux" ;;
    darwin*)  OS="darwin" ;;
    msys*|mingw*|cygwin*) OS="windows" ;;
    *)        echo "Unsupported OS: ${OS}"; exit 1 ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)            echo "Unsupported architecture: ${ARCH}"; exit 1 ;;
esac

# Fetch latest version from GitHub Release
echo "Checking latest release from GitHub..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${LATEST_TAG}" ]; then
    echo "Failed to fetch latest release version. Defaulting to v0.1.0..."
    LATEST_TAG="v0.1.0"
fi

# Clean tag name for version string (remove 'v' prefix if needed)
VERSION="${LATEST_TAG#v}"

# Define download URL
# Example format matching Goreleaser template: knowblazer_0.1.0_linux_amd64.tar.gz
TARBALL_NAME="knowblazer_${VERSION}_${OS}_${ARCH}.tar.gz"
if [ "${OS}" = "windows" ]; then
    TARBALL_NAME="knowblazer_${VERSION}_${OS}_${ARCH}.zip"
fi

DOWNLOAD_URL="https://github.com/${OWNER}/${REPO}/releases/download/${LATEST_TAG}/${TARBALL_NAME}"

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
    # Fallback to local user bin if /usr/local/bin is not writable without sudo
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${INSTALL_DIR}"
fi

# Create a temporary directory
TMP_DIR=$(mktemp -d)
clean_up() {
    rm -rf "${TMP_DIR}"
}
trap clean_up EXIT

echo "Downloading ${LATEST_TAG} for ${OS}/${ARCH}..."
curl -fsSL -o "${TMP_DIR}/${TARBALL_NAME}" "${DOWNLOAD_URL}"

echo "Extracting binary..."
if [ "${OS}" = "windows" ]; then
    unzip -q -o "${TMP_DIR}/${TARBALL_NAME}" -d "${TMP_DIR}"
else
    tar -xzf "${TMP_DIR}/${TARBALL_NAME}" -C "${TMP_DIR}"
fi

echo "Installing to ${INSTALL_DIR}..."
mv "${TMP_DIR}/knowblazer" "${INSTALL_DIR}/knowblazer"
chmod +x "${INSTALL_DIR}/knowblazer"

echo ""
echo "========================================="
echo " Knowblazer installed successfully!"
echo " Location: ${INSTALL_DIR}/knowblazer"
echo "========================================="
echo ""
echo "To get started, run this command in your project directory:"
echo "    knowblazer start"
echo ""
