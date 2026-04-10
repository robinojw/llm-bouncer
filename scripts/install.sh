#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/robinojw/llm-bouncer.git"
TMP_DIR=$(mktemp -d)

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "==> Cloning llm-bouncer..."
git clone --depth=1 "$REPO" "$TMP_DIR"

echo "==> Running installer..."
bash "$TMP_DIR/install.sh"
