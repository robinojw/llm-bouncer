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

uninstall() {
  echo "==> Uninstalling llm-bouncer..."

  # Remove binary and install dir
  if [[ -d "$INSTALL_DIR" ]]; then
    rm -rf "$INSTALL_DIR"
    echo "  Removed $INSTALL_DIR"
  fi

  # Remove hook entry from Claude Code settings
  local claude_config="$HOME/.claude/settings.json"
  if [[ -f "$claude_config" ]]; then
    python3 - "$claude_config" <<'PYEOF'
import json, sys

config_path = sys.argv[1]

with open(config_path, "r") as f:
    config = json.load(f)

hooks = config.get("hooks", {})
post_tool = hooks.get("PostToolUse", [])

filtered = [
    entry for entry in post_tool
    if not any("llm-bouncer" in h.get("command", "") for h in entry.get("hooks", []))
]

if len(filtered) != len(post_tool):
    if filtered:
        hooks["PostToolUse"] = filtered
    else:
        hooks.pop("PostToolUse", None)
    with open(config_path, "w") as f:
        json.dump(config, f, indent=2)
        f.write("\n")
    print("  Removed hook from Claude Code settings.")
else:
    print("  No hook found in Claude Code settings.")
PYEOF
  fi

  # Remove Codex hooks.json if it's our placeholder
  local codex_config="$HOME/.codex/hooks.json"
  if [[ -f "$codex_config" ]]; then
    local content
    content=$(python3 -c "
import json
with open('$codex_config') as f:
    c = json.load(f)
print('empty' if c == {'hooks': {}} else 'has_content')
")
    if [[ "$content" == "empty" ]]; then
      rm "$codex_config"
      echo "  Removed Codex hooks.json placeholder."
    else
      echo "  Codex hooks.json has custom content, leaving in place."
    fi
  fi

  echo ""
  echo "Uninstall complete."
}

patch_codex() {
  local config="$HOME/.codex/hooks.json"
  echo "==> Patching Codex CLI ($config)..."

  if [[ ! -d "$HOME/.codex" ]]; then
    echo "  Codex CLI config directory not found, skipping."
    return
  fi

  if [[ -f "$config" ]]; then
    echo "  hooks.json already exists, skipping."
    return
  fi

  cat > "$config" <<JSONEOF
{
  "hooks": {}
}
JSONEOF

  echo "  Created hooks.json placeholder."
  echo "  Note: Codex CLI lacks PostToolUse. Use the wrapper instead:"
  echo "    $INSTALL_DIR/codex-lint.sh \"your prompt\""
}

main "$@"
