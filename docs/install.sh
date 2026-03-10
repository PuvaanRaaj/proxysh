#!/usr/bin/env sh
set -e

REPO="PuvaanRaaj/devtun"
BIN="devtun"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

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
  echo "Error: No releases found for ${REPO}."
  echo "Visit https://github.com/${REPO}/releases to check availability."
  exit 1
fi

TARBALL="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${TARBALL}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading devtun ${LATEST} (${OS}/${ARCH})..."
curl -sL "$URL" -o "${TMP}/${TARBALL}"

echo "Extracting..."
tar -xzf "${TMP}/${TARBALL}" -C "$TMP"

mkdir -p "$INSTALL_DIR"
echo "Installing to ${INSTALL_DIR}/${BIN}..."
mv "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"
chmod +x "${INSTALL_DIR}/${BIN}"

# Warn if INSTALL_DIR is not in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "  Note: add ${INSTALL_DIR} to your PATH:"
     echo "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;;
esac

echo ""
echo "devtun installed successfully!"
echo ""
echo "Get started:"
echo "  devtun start              # set up certificates and start daemon"
echo "  devtun up example 3000      # https://example.test → localhost:3000"
echo "  devtun list               # list active domains"
echo ""
