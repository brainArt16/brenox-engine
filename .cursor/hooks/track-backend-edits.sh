#!/usr/bin/env bash
# Records backend file edits for doc-sync reminder on agent stop.
set -euo pipefail

input=$(cat)
path=""

if command -v python3 >/dev/null 2>&1; then
  path=$(echo "$input" | python3 -c "
import json, sys
try:
    d = json.load(sys.stdin)
    print(d.get('file_path') or d.get('path') or d.get('file') or '')
except Exception:
    print('')
" 2>/dev/null || true)
fi

if [[ -z "$path" ]]; then
  exit 0
fi

case "$path" in
  internal/*|cmd/*|sql/*|pkg/*)
    mkdir -p .cursor/state
    echo "$path" >> .cursor/state/backend-edits.log
    ;;
esac

exit 0
