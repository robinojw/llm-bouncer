# Global Install Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create an `install.sh` that builds llm-bouncer and installs it globally, patching Claude Code and Codex CLI configs so the linter runs on every project automatically.

**Architecture:** Bash installer builds the Go binary, copies it to `~/.llm-bouncer/bin/`, and merges hook entries into each tool's global config using Python3 for JSON manipulation. Supports `--uninstall` to cleanly remove everything.

**Tech Stack:** Bash, Python3 (for JSON merge), Go (build)

---

### Task 1: Create install.sh with build and binary install

**Files:**
- Create: `install.sh`

**Step 1: Write the install script skeleton**

```bash
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
  (cd "$SCRIPT_DIR" && go build -o llm-bouncer .)

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
```

**Step 2: Verify the skeleton runs**

Run: `cd /Users/robin.white/dev/llm-bouncer && bash install.sh 2>&1 | head -5`
Expected: "==> Checking prerequisites..." then build output (will fail at patch functions since they don't exist yet — that's fine)

**Step 3: Commit**

```bash
git add install.sh
git commit -m "feat: add install.sh skeleton with build and binary install"
```

---

### Task 2: Add Claude Code config patching

**Files:**
- Modify: `install.sh`

**Step 1: Add the `patch_claude_code` function**

This function merges the PostToolUse hook entry into `~/.claude/settings.json`. It uses Python3 to safely parse, merge, and write JSON. It is idempotent — running twice does not duplicate the entry.

```bash
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
```

**Step 2: Test idempotency**

Run install twice, check that `~/.claude/settings.json` only has one llm-bouncer entry:

```bash
bash install.sh
bash install.sh
python3 -c "
import json
with open('$HOME/.claude/settings.json') as f:
    c = json.load(f)
entries = [e for e in c['hooks']['PostToolUse'] for h in e['hooks'] if 'llm-bouncer' in h.get('command','')]
assert len(entries) == 1, f'Expected 1 entry, got {len(entries)}'
print('OK: idempotent')
"
```

**Step 3: Verify existing hooks preserved**

Check that the existing `SessionStart` hook in `~/.claude/settings.json` is still present after install.

**Step 4: Commit**

```bash
git add install.sh
git commit -m "feat: add Claude Code global config patching to installer"
```

---

### Task 3: Add Codex config patching

**Files:**
- Modify: `install.sh`

**Step 1: Add the `patch_codex` function**

Codex CLI doesn't have PostToolUse yet, so we create a minimal `hooks.json` placeholder and print a note about the wrapper script.

```bash
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
```

**Step 2: Test**

Run install, verify `~/.codex/hooks.json` was created:

```bash
bash install.sh
cat ~/.codex/hooks.json
```

Expected: `{"hooks": {}}`

**Step 3: Commit**

```bash
git add install.sh
git commit -m "feat: add Codex CLI config patching to installer"
```

---

### Task 4: Add --uninstall support

**Files:**
- Modify: `install.sh`

**Step 1: Add the `uninstall` function**

```bash
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
```

**Step 2: Test install then uninstall**

```bash
bash install.sh
ls ~/.llm-bouncer/bin/llm-bouncer  # should exist
bash install.sh --uninstall
ls ~/.llm-bouncer/ 2>&1  # should error (not found)
```

Verify `~/.claude/settings.json` still has the SessionStart hook but no PostToolUse/llm-bouncer entry.

**Step 3: Commit**

```bash
git add install.sh
git commit -m "feat: add --uninstall support to installer"
```

---

### Task 5: Update codex-lint.sh for global binary path

**Files:**
- Modify: `codex-lint.sh`

**Step 1: Change LINT_BINARY to use global path**

Change line 10 from:
```bash
LINT_BINARY=".claude/hooks/llm-bouncer/llm-bouncer"
```
to:
```bash
LINT_BINARY="${LLM_BOUNCER_BIN:-$HOME/.llm-bouncer/bin/llm-bouncer}"
```

This uses the global path by default but allows override via environment variable.

**Step 2: Add binary existence check**

After the `LINT_BINARY` line, add:
```bash
if [[ ! -x "$LINT_BINARY" ]]; then
  echo "Error: llm-bouncer not found at $LINT_BINARY"
  echo "Run install.sh first, or set LLM_BOUNCER_BIN to the binary path."
  exit 1
fi
```

**Step 3: Commit**

```bash
git add codex-lint.sh
git commit -m "feat: update codex wrapper to use global binary path"
```

---

### Task 6: Update README.md

**Files:**
- Modify: `README.md`

**Step 1: Replace the Quick Start section**

Replace the current "Quick start" section (### 1. Build, ### 2. Install for Claude Code, ### 3. Install for Codex CLI) with:

```markdown
## Quick start

### One-command install

```bash
git clone https://github.com/robinojw/llm-bouncer.git
cd llm-bouncer
bash install.sh
```

This builds the binary, installs it to `~/.llm-bouncer/bin/`, and patches global configs for:
- **Claude Code** — adds a `PostToolUse` hook to `~/.claude/settings.json`
- **Codex CLI** — installs a wrapper script at `~/.llm-bouncer/codex-lint.sh`

### Codex CLI usage

Codex CLI doesn't support PostToolUse hooks yet. Use the wrapper script:

```bash
~/.llm-bouncer/codex-lint.sh "add a new endpoint"
```

### Uninstall

```bash
bash install.sh --uninstall
```

Removes the binary, wrapper, and all hook entries from global configs.
```

Keep the "Standalone usage", "Tuning", "Rules", "Project structure", and "License" sections unchanged. Update the "Project structure" tree to include `install.sh` and `docs/`.

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README for global install"
```

---

### Task 7: End-to-end verification

**Step 1: Clean slate**

```bash
bash install.sh --uninstall 2>/dev/null || true
```

**Step 2: Run install**

```bash
bash install.sh
```

Expected output includes all "==>" steps completing without error.

**Step 3: Verify binary**

```bash
~/.llm-bouncer/bin/llm-bouncer --help 2>&1 || echo "binary runs"
```

**Step 4: Verify Claude Code config**

```bash
python3 -c "
import json
with open('$HOME/.claude/settings.json') as f:
    c = json.load(f)
# Check PostToolUse hook exists
pt = c['hooks']['PostToolUse']
found = any('llm-bouncer' in h['command'] for e in pt for h in e['hooks'])
assert found, 'llm-bouncer hook not found'
# Check existing hooks preserved
assert 'SessionStart' in c['hooks'], 'SessionStart hook lost'
print('Claude Code config OK')
"
```

**Step 5: Verify idempotency**

```bash
bash install.sh
python3 -c "
import json
with open('$HOME/.claude/settings.json') as f:
    c = json.load(f)
entries = [e for e in c['hooks']['PostToolUse'] for h in e['hooks'] if 'llm-bouncer' in h.get('command','')]
assert len(entries) == 1, f'Duplicate entries: {len(entries)}'
print('Idempotency OK')
"
```

**Step 6: Verify uninstall**

```bash
bash install.sh --uninstall
python3 -c "
import json
with open('$HOME/.claude/settings.json') as f:
    c = json.load(f)
pt = c.get('hooks', {}).get('PostToolUse', [])
found = any('llm-bouncer' in h.get('command','') for e in pt for h in e.get('hooks', []))
assert not found, 'Hook not removed'
assert 'SessionStart' in c['hooks'], 'SessionStart hook lost during uninstall'
print('Uninstall OK')
"
```

**Step 7: Reinstall for real use**

```bash
bash install.sh
```

**Step 8: Final commit**

```bash
git add -A
git commit -m "feat: global installer for llm-bouncer"
```
