#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="$HOME/.llm-bouncer"
BIN_DIR="$INSTALL_DIR/bin"
BINARY="$BIN_DIR/llm-bouncer"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

main() {
  if [[ "${1:-}" == "--uninstall" ]]; then
    uninstall
    exit 0
  fi

  echo "==> Checking prerequisites..."
  command -v go >/dev/null 2>&1 || { echo "Error: go is not installed"; exit 1; }
  command -v python3 >/dev/null 2>&1 || { echo "Error: python3 is required for config patching"; exit 1; }

  echo "==> Building llm-bouncer..."
  (cd "$SCRIPT_DIR" && CGO_ENABLED=1 go build -o llm-bouncer .)

  echo "==> Installing binary to $BIN_DIR..."
  mkdir -p "$BIN_DIR"
  cp "$SCRIPT_DIR/llm-bouncer" "$BINARY"
  chmod +x "$BINARY"

  echo "==> Installing codex wrapper to $INSTALL_DIR..."
  cp "$SCRIPT_DIR/codex-lint.sh" "$INSTALL_DIR/codex-lint.sh"
  chmod +x "$INSTALL_DIR/codex-lint.sh"

  patch_claude_code
  patch_codex

  echo ""
  echo "Done! llm-bouncer installed globally."
  echo "  Binary:        $BINARY"
  echo "  Codex wrapper: $INSTALL_DIR/codex-lint.sh"
  echo ""
  echo "To uninstall: $SCRIPT_DIR/install.sh --uninstall"
}

main "$@"
