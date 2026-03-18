#!/usr/bin/env bash
# install.sh — Build llm-bouncer and install it as a Claude Code hook
#
# Usage:
#   cd your-project && bash /path/to/llm-bouncer/install.sh
#   # or:
#   bash <(curl -fsSL https://raw.githubusercontent.com/robinojw/llm-bouncer/main/install.sh)

set -euo pipefail

BOUNCER_REPO="https://github.com/robinojw/llm-bouncer.git"
HOOK_DIR=".claude/hooks/llm-bouncer"
SETTINGS_FILE=".claude/settings.json"

# ---------- helpers ----------

die() { echo "ERROR: $1" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required but not installed"
}

# ---------- pre-flight ----------

require_cmd go
require_cmd git
require_cmd python3

# CGO is required for tree-sitter (C library).
export CGO_ENABLED=1

# Determine the project root (where the hook will be installed).
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
echo "Installing llm-bouncer into: $PROJECT_DIR"

# ---------- clone or locate source ----------

if [ -f "go.mod" ] && grep -q "module llm-bouncer" go.mod 2>/dev/null; then
  # Running from inside the llm-bouncer repo itself.
  SRC_DIR="$(pwd)"
else
  # Clone to a temp directory.
  SRC_DIR="$(mktemp -d)"
  trap 'rm -rf "$SRC_DIR"' EXIT
  echo "Cloning llm-bouncer..."
  git clone --depth 1 "$BOUNCER_REPO" "$SRC_DIR"
fi

# ---------- build ----------

echo "Building llm-bouncer (CGO_ENABLED=1)..."
cd "$SRC_DIR"
go build -o llm-bouncer .

# ---------- install binary ----------

mkdir -p "$PROJECT_DIR/$HOOK_DIR"
cp llm-bouncer "$PROJECT_DIR/$HOOK_DIR/llm-bouncer"
chmod +x "$PROJECT_DIR/$HOOK_DIR/llm-bouncer"
echo "Binary installed to $PROJECT_DIR/$HOOK_DIR/llm-bouncer"

# ---------- patch settings.json ----------

HOOK_ENTRY='{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/llm-bouncer/llm-bouncer"
          }
        ]
      }
    ]
  }
}'

mkdir -p "$PROJECT_DIR/.claude"

if [ -f "$PROJECT_DIR/$SETTINGS_FILE" ]; then
  echo "Merging hook into existing $SETTINGS_FILE..."
  python3 -c "
import json, sys

with open('$PROJECT_DIR/$SETTINGS_FILE') as f:
    settings = json.load(f)

hook_entry = json.loads('''$HOOK_ENTRY''')

hooks = settings.setdefault('hooks', {})
post = hooks.setdefault('PostToolUse', [])

# Avoid duplicates.
cmd = '.claude/hooks/llm-bouncer/llm-bouncer'
already = any(
    h.get('command') == cmd
    for group in post
    for h in group.get('hooks', [])
)

if not already:
    post.append(hook_entry['hooks']['PostToolUse'][0])
    with open('$PROJECT_DIR/$SETTINGS_FILE', 'w') as f:
        json.dump(settings, f, indent=2)
    print('Hook added to $SETTINGS_FILE')
else:
    print('Hook already present in $SETTINGS_FILE')
"
else
  echo "$HOOK_ENTRY" | python3 -c "
import json, sys
data = json.load(sys.stdin)
with open('$PROJECT_DIR/$SETTINGS_FILE', 'w') as f:
    json.dump(data, f, indent=2)
"
  echo "Created $SETTINGS_FILE with hook configuration"
fi

echo ""
echo "Done! llm-bouncer is installed."
echo "Supported languages: Go, Python, TypeScript, JavaScript, Rust, Java"
