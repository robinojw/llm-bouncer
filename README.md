
<img width="928" height="1138" alt="be87e3fa-a874-49cf-ac4b-0a2a111804d5" src="https://github.com/user-attachments/assets/2a6a5043-1057-4ecd-88e3-a0fdc57ccfcf" />

# llm-bouncer

A code quality gate for AI coding agents. Hooks into [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [Codex CLI](https://github.com/openai/codex) to catch Go style violations **before** they land, forcing the agent to fix its own output.

## How it works

When Claude Code writes or edits a `.go` file, the `PostToolUse` hook pipes the event to the `llm-bouncer` binary. If violations are found, it returns `{"decision": "block", "reason": "..."}` and Claude must fix the issues before continuing.

For Codex CLI (which lacks native hook support), a wrapper script diffs changed files after each run.

## Rules

| Rule | What it catches |
|---|---|
| `naming` | Single-letter variables (except `i`, `j`, `k`, `_`, and method receivers) |
| `no-nested-ifs` | `if` inside `if` — use early returns or extract a helper |
| `no-inline-booleans` | `&&` / `\|\|` directly in `if` conditions — assign to a named variable |
| `no-inline-comments` | Comments on the same line as code — write self-documenting code |
| `no-repeated-strings` | Same string literal used more than once — extract to a constant |
| `no-magic-numbers` | Numeric literals (except `0` and `1`) outside `const` blocks |
| `cyclomatic-complexity` | Functions exceeding complexity threshold (default 10) |
| `file-size` | Files exceeding line limit (default 300) |
| `file-naming` | Filenames that aren't `snake_case.go` |

## Quick start

### 1. Build

```bash
git clone https://github.com/robinojw/llm-bouncer.git
cd llm-bouncer
go build -o llm-bouncer .
```

### 2. Install for Claude Code

Copy the binary into your project's hook directory:

```bash
mkdir -p your-project/.claude/hooks/llm-bouncer
cp llm-bouncer your-project/.claude/hooks/llm-bouncer/
```

Add the hook to your project's `.claude/settings.json`:

```json
{
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
}
```

### 3. Install for Codex CLI

Use the wrapper script instead of calling `codex` directly:

```bash
cp codex-lint.sh your-project/
chmod +x your-project/codex-lint.sh

# Then use it in place of codex:
./codex-lint.sh "add a new endpoint"
```

See [openai/codex#7396](https://github.com/openai/codex/issues/7396) for native hook support progress.

## Standalone usage

Run directly against any Go file:

```bash
./llm-bouncer path/to/file.go
```

Returns JSON on stdout when violations are found, exits silently when clean.

## Tuning

The defaults are conservative. Adjust these constants in the source to match your codebase:

| Constant | File | Default | Description |
|---|---|---|---|
| `MaxFileLines` | `checker/files.go` | 300 | Maximum lines per file |
| `maxCyclomaticComplexity` | `checker/complexity.go` | 10 | Maximum cyclomatic complexity per function |

## Project structure

```
.
├── main.go              # Hook entry point — reads stdin JSON or CLI args
├── checker/
│   ├── checker.go       # Violation type and CheckFile orchestrator
│   ├── complexity.go    # Cyclomatic complexity check
│   ├── files.go         # File size and naming checks
│   ├── naming.go        # Variable and parameter naming checks
│   └── style.go         # Nested ifs, inline booleans, comments, strings, magic numbers
├── codex-lint.sh        # Codex CLI wrapper script
└── go.mod
```

## License

MIT
