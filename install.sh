#!/usr/bin/env sh
set -e

REPO="PuvaanRaaj/proxysh"
BIN="proxysh"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Get latest release tag
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest release. Using 'latest'."
  LATEST="latest"
fi

TARBALL="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${TARBALL}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading proxysh ${LATEST} (${OS}/${ARCH})..."
curl -sL "$URL" -o "${TMP}/${TARBALL}"

echo "Extracting..."
tar -xzf "${TMP}/${TARBALL}" -C "$TMP"

echo "Installing to ${INSTALL_DIR}/${BIN}..."
sudo mv "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"
sudo chmod +x "${INSTALL_DIR}/${BIN}"

echo ""
echo "proxysh installed successfully!"
echo ""
echo "Get started:"
echo "  proxysh start              # set up certificates and start daemon"
echo "  proxysh up myapp 3000      # https://myapp.test → localhost:3000"
echo "  proxysh list               # list active domains"
echo ""
