#!/usr/bin/env bash
# Run the live smoke test against the host configured in ./.env.
#
# Mirrors `yarn smoke:local` in sdk-node: loads variables from .env and runs
# scripts/livesmoke.py. `uv run --no-sync` reuses the existing .venv without
# touching dependencies or writing a uv.lock (matching this repo's uv
# convention). The plain CI path stays `python scripts/livesmoke.py`.
set -euo pipefail
cd "$(dirname "$0")/.."
exec uv run --no-sync --env-file .env python scripts/livesmoke.py
