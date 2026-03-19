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

patch_claude_code() {
  local config="$HOME/.claude/settings.json"
  echo "==> Patching Claude Code ($config)..."

  if [[ ! -f "$config" ]]; then
    mkdir -p "$HOME/.claude"
    echo '{}' > "$config"
  fi

  python3 - "$config" "$BINARY" <<'PYEOF'
import json, sys

config_path = sys.argv[1]
binary_path = sys.argv[2]

with open(config_path, "r") as f:
    config = json.load(f)

hook_entry = {
    "matcher": "Write|Edit|MultiEdit",
    "hooks": [{"type": "command", "command": binary_path}]
}

hooks = config.setdefault("hooks", {})
post_tool = hooks.setdefault("PostToolUse", [])

# Idempotent: skip if already present
for existing in post_tool:
    for h in existing.get("hooks", []):
        if h.get("command") == binary_path:
            print("  Already configured, skipping.")
            sys.exit(0)

post_tool.append(hook_entry)

with open(config_path, "w") as f:
    json.dump(config, f, indent=2)
    f.write("\n")

print("  Added PostToolUse hook.")
PYEOF
}

main "$@"
