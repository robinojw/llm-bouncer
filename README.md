# llm-bouncer

A **language-agnostic** code quality gate for AI coding agents. Hooks into [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [Codex CLI](https://github.com/openai/codex) to catch style violations **before** they land, forcing the agent to fix its own output.

## Supported languages

| Language | File extensions | Naming convention |
|---|---|---|
| Go | `.go` | `snake_case.go` |
| Python | `.py` | `snake_case.py` |
| TypeScript | `.ts`, `.tsx` | `camelCase.ts`, `PascalCase.tsx`, `kebab-case.ts` |
| JavaScript | `.js`, `.jsx` | `camelCase.js`, `PascalCase.jsx`, `kebab-case.js` |
| Rust | `.rs` | `snake_case.rs` |
| Java | `.java` | `PascalCase.java` |

Powered by [tree-sitter](https://tree-sitter.github.io/) for universal AST parsing.

## How it works

When Claude Code writes or edits a supported file, the `PostToolUse` hook pipes the event to the `llm-bouncer` binary. If violations are found, it returns `{"decision": "block", "reason": "..."}` and Claude must fix the issues before continuing.

For Codex CLI (which lacks native hook support), a wrapper script diffs changed files after each run.

## Rules

| Rule | What it catches | Language notes |
|---|---|---|
| `naming` | Single-letter variables (except `i`, `j`, `k`, `_`) | Python also allows `self`, `cls`; Go excludes method receivers |
| `no-nested-ifs` | `if` inside `if` — use early returns or extract a helper | Rust uses `if_expression` |
| `no-inline-booleans` | `&&`/`\|\|`/`and`/`or` directly in `if` conditions | Python uses `and`/`or` keywords |
| `no-inline-comments` | Comments on the same line as code | All comment syntaxes: `//`, `#`, `/* */` |
| `no-repeated-strings` | Same string literal used more than once | Includes template strings for JS/TS |
| `no-magic-numbers` | Numeric literals (except `0` and `1`) outside constants | Go: `const` block, JS/TS: `const` keyword, Python: `UPPER_SNAKE_CASE`, Java: `final` |
| `cyclomatic-complexity` | Functions exceeding complexity threshold (default 10) | Counts `if`, loops, cases, `&&`/`\|\|` |
| `file-size` | Files exceeding line limit (default 300) | |
| `file-naming` | Filenames violating language convention | See table above |

## Quick start

### One-liner install

From your project directory:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/robinojw/llm-bouncer/main/install.sh)
```

This builds the binary, copies it to `.claude/hooks/llm-bouncer/`, and patches your `.claude/settings.json`.

**Requires:** Go 1.21+, git, python3, and a C compiler (CGO is required for tree-sitter).

### Manual install

#### 1. Build

```bash
git clone https://github.com/robinojw/llm-bouncer.git
cd llm-bouncer
CGO_ENABLED=1 go build -o llm-bouncer .
```

#### 2. Install for Claude Code

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

#### 3. Install for Codex CLI

Use the wrapper script instead of calling `codex` directly:

```bash
cp codex-lint.sh your-project/
chmod +x your-project/codex-lint.sh

# Then use it in place of codex:
./codex-lint.sh "add a new endpoint"
```

See [openai/codex#7396](https://github.com/openai/codex/issues/7396) for native hook support progress.

## Standalone usage

Run directly against any supported file:

```bash
./llm-bouncer path/to/file.py
./llm-bouncer path/to/component.tsx
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
├── language/
│   ├── language.go      # LanguageConfig type, Detect(), registry
│   ├── go.go            # Go config
│   ├── python.go        # Python config
│   ├── typescript.go    # TypeScript config
│   ├── javascript.go    # JavaScript config
│   ├── rust.go          # Rust config
│   └── java.go          # Java config
├── checker/
│   ├── checker.go       # Violation type, tree-sitter parsing, walk helpers
│   ├── complexity.go    # Cyclomatic complexity check
│   ├── files.go         # File size and naming checks
│   ├── naming.go        # Variable and parameter naming checks
│   └── style.go         # Nested ifs, inline booleans, comments, strings, magic numbers
├── install.sh           # One-command installer
├── codex-lint.sh        # Codex CLI wrapper script
└── go.mod
```

## License

MIT
