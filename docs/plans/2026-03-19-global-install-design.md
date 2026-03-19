# Global Install for llm-bouncer

## Problem

llm-bouncer installs hooks per-project (`.claude/hooks/llm-bouncer/` and `.claude/settings.json` in the working directory). Users must repeat the setup for every project. The installer should configure global hooks so the linter runs automatically in all projects.

## Target Tools

| Tool | Global Config | Hook Event | Protocol |
|---|---|---|---|
| Claude Code | `~/.claude/settings.json` | `PostToolUse` + `Write\|Edit\|MultiEdit` | JSON stdin/stdout |
| Codex CLI | `~/.codex/hooks.json` | No PostToolUse yet | JSON stdin/stdout (wrapper script for now) |

## Design

### Directory layout

```
~/.llm-bouncer/
├── bin/
│   └── llm-bouncer          # compiled binary
└── codex-lint.sh             # wrapper referencing global binary
```

### install.sh flow

1. Check `go` is on PATH
2. Build binary: `go build -o llm-bouncer .`
3. Install binary to `~/.llm-bouncer/bin/`
4. Merge PostToolUse hook into `~/.claude/settings.json` (non-destructive)
5. Create `~/.codex/hooks.json` placeholder
6. Copy `codex-lint.sh` to `~/.llm-bouncer/` with global binary path
7. Print summary

### Claude Code hook entry

Merged into `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [{
          "type": "command",
          "command": "$HOME/.llm-bouncer/bin/llm-bouncer"
        }]
      }
    ]
  }
}
```

Uses absolute path expanded from `$HOME` at install time.

### Config merge strategy

Read existing JSON, append llm-bouncer entry to `hooks.PostToolUse[]` array. If the array already contains an entry with the same command path, skip (idempotent). If no hooks key exists, create it.

### Uninstall (`--uninstall`)

1. Remove `~/.llm-bouncer/` directory
2. Remove llm-bouncer entry from `~/.claude/settings.json`
3. Remove `~/.codex/hooks.json` if empty

### File changes

- New: `install.sh` (the installer)
- Modified: `codex-lint.sh` (update `LINT_BINARY` to `$HOME/.llm-bouncer/bin/llm-bouncer`)
- Modified: `README.md` (update quick start to show global install)
