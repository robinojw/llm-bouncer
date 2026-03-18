#!/usr/bin/env bash
# codex-lint.sh — Run codex then lint all modified Go files
#
# Since Codex CLI has no native hook support yet, this wrapper detects
# which Go files changed and runs the linter after codex completes.
# See: https://github.com/openai/codex/issues/7396

set -euo pipefail

LINT_BINARY=".claude/hooks/llm-bouncer/llm-bouncer"
VIOLATIONS_FOUND=false

# Run codex with all passed arguments
codex "$@"

# Find all Go files modified since last commit (includes staged and unstaged)
CHANGED_FILES=$(git diff --name-only HEAD -- '*.go' 2>/dev/null || true)
UNSTAGED=$(git diff --name-only -- '*.go' 2>/dev/null || true)
CHANGED_FILES=$(printf '%s\n%s' "$CHANGED_FILES" "$UNSTAGED" | sort -u)

while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  [[ ! -f "$file" ]] && continue

  VIOLATIONS=$("$LINT_BINARY" "$file" 2>/dev/null || true)
  if [[ -n "$VIOLATIONS" ]]; then
    echo "$VIOLATIONS"
    VIOLATIONS_FOUND=true
  fi
done <<< "$CHANGED_FILES"

if [[ "$VIOLATIONS_FOUND" == "true" ]]; then
  echo ""
  echo "Lint violations found. Review the output above."
  exit 1
fi
