#!/usr/bin/env bash
# Prompts documentation sync when backend files were edited this session.
set -euo pipefail

LOG=".cursor/state/backend-edits.log"

if [[ ! -f "$LOG" ]] || [[ ! -s "$LOG" ]]; then
  echo '{}'
  exit 0
fi

files=$(sort -u "$LOG" | paste -sd ', ' -)
rm -f "$LOG"

python3 - <<PY
import json

files = """${files}"""

msg = f"""Backend files were edited this session ({files}).

Run the documentation sync workflow before finishing:

1. Task Tracker Agent — update docs/BACKEND_TASKS.md (checkboxes, phase counts, changelog).
2. README Agent — update README.md if routes, setup, env vars, or layout changed.
3. Postman Agent — update docs/postman/brenox.postman_collection.json for any HTTP API changes.

Read AGENTS.md and .cursor/rules/ for checklists. Only mark tasks [x] when implementation is complete."""

print(json.dumps({"followup_message": msg}))
PY

exit 0
