#!/usr/bin/env bash
set -e

REPO="dk-a-dev/termf1"

# Detect OS

OS=$(uname | tr '[:upper:]' '[:lower:]')

# Detect architecture

ARCH=$(uname -m)
case "$ARCH" in
x86_64) ARCH="amd64" ;;
aarch64|arm64) ARCH="arm64" ;;
*) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version tag from GitHub

VERSION=$(curl -fsSL https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name"' | cut -d '"' -f4)

FILE="termf1-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILE"

echo "Installing termf1 $VERSION for $OS-$ARCH"

curl -L "$URL" | tar -xz

sudo mv termf1 /usr/local/bin/

echo "termf1 installed successfully"
