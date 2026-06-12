---
description: Implement a new resource in the Python SDK (sdk-python) from the extracted OpenAPI slice, with examples and tests.
argument-hint: <resource/domain name, e.g. widgets>
---

# Implement the Python SDK resource

You are adding a new resource to the **Python** SDK only (`sdk-python/`). Follow
the `3common-sdk-conventions` skill for all layout, naming, codegen, and
quality-gate rules -- it is preloaded into your context.

## Canonical slug: use this, not the raw message
Read `$ARTIFACTS_DIR/resource-spec.json` and take its **`domain`** field as the
resource slug (always the lowercase OpenAPI path segment, the first path
segment after `/v1/`). Below,
**`<slug>`** means that value. Use `<slug>` verbatim for the conformance scenario
directory and the `call.resource` value. For names derived from a hyphenated slug
(e.g. `payment-links`), since Python names cannot contain hyphens: the
package/module directory, dispatcher module, and client attribute use snake_case
(`payment_links`), while classes use PascalCase (`PaymentLinksService`). A
single-word slug is used as-is; mirror the sibling's casing. Do NOT derive
paths or names from the raw invocation message, as it may be capitalized or a whole
sentence; only `domain` is canonical, and it is the exact path the publish step
stages.

## Inputs
- The API contract to wrap: `$ARTIFACTS_DIR/resource-spec.json` contains the `domain`
  slug plus every endpoint (method, path, params, request/response schemas),
  sliced from the canonical spec. This is your source of truth for what to wrap
  and how to call it.

## Scope
- Modify files **only under `sdk-python/`**. Do not touch `sdk-node/`,
  `sdk-go/`, `openapi/`, or any other SDK: those are handled separately and
  must not appear in this change.
- Do **not** run any `git` commands (no branch/add/commit/push). A later step
  owns all git and PR creation. Just leave your changes in the working tree.

## Steps
0. **Install dependencies** (the worktree has no `.venv`): `cd sdk-python && uv venv --python 3.10 .venv && uv pip install -e ".[dev]"`. Run all Python tooling below via `uv run --no-sync <tool>`, as it uses this venv exactly as installed (cross-platform, no `Scripts/`-vs-`bin/` issues; and `--no-sync` avoids a re-sync that could prune the dev extras, and never writes `uv.lock`).
1. **No codegen - hand-write the types.** The Python SDK does NOT use a generated layer at runtime: `src/threecommon/_generated/models.py` is a contract-reference artifact that nothing imports, so you write all request/response models by hand in step 3. Do NOT run `datamodel-codegen` or regenerate `_generated/` (it rewrites the whole file from the full spec, producing a large diff unrelated to this resource, and CI never runs it).
2. **Study a sibling.** Read `sdk-python/src/threecommon/contacts/{service.py,types.py,__init__.py}` (or the closest-shaped sibling) and mirror its structure, docstrings, and idioms, including providing BOTH a sync and an async service.
3. **Write the resource** under `sdk-python/src/threecommon/<slug>/`: `types.py`, `service.py` (sync `Service` + `AsyncService`), `__init__.py` (alphabetized `__all__`) -- one method per endpoint in `resource-spec.json`.
4. **Mount it** in `sdk-python/src/threecommon/client.py` in BOTH `ThreeCommon` and `AsyncThreeCommon`: import, declare the attribute, assign in `__init__`.
5. **Add examples**: one runnable file per endpoint under `sdk-python/examples/<slug>/<op>.py` (snake_case), with `_sync`/`_async` variants where the sibling has them. Use the `"3co_your_api_key_here"` placeholder key.
6. **Write tests** under `sdk-python/tests/` matching the sibling's layout, covering sync + async paths and errors. Meet the **>= 90%** coverage gate (`pytest --cov-fail-under=90`).
7. **Wire conformance.** The shared scenarios in `conformance/scenarios/<slug>/` are already generated and define the canonical method names/args your Python API must match. Create `sdk-python/tests/_conformance/dispatch_<slug>.py` with `dispatch_sync` and `dispatch_async` (mirror `dispatch_contacts.py`), then register it in `sdk-python/tests/test_conformance.py`: add `dispatch_<slug>` to the `from _conformance import (...)` block and add an `if resource == "<slug>"` branch in BOTH `_dispatch_sync` and `_dispatch_async`.
8. **Update the changelog**: add a bullet under `## [Unreleased]` -> `### Added` in `sdk-python/CHANGELOG.md` (create the `### Added` subheading if absent), mentioning both sync and async surfaces. Mirror the most recent resource entry's wording. See the conventions skill.
9. **Get to green** (run from `sdk-python/`, each via `uv run --no-sync`, mirroring `.github/workflows/python.yml` exactly): `uv run --no-sync ruff format .` to auto-format, then iterate until all pass -- `uv run --no-sync ruff check .`, `uv run --no-sync ruff format --check .`, `uv run --no-sync mypy src/threecommon tests`, `uv run --no-sync pyright src/threecommon`, `uv run --no-sync pytest --cov=src/threecommon --cov-fail-under=90` (the conformance suite runs under pytest here too). Strictness is configured in `pyproject.toml` (mypy `strict = true`, pyright `standard`, the curated ruff `select`): do NOT add `--strict` or `--select=ALL`, and do NOT run mypy over `.` (it pulls in `examples/`, whose reused module names like `retrieve.py` make mypy abort with duplicate-module errors). Do not stop while any is red.

## Output
Write a short summary to `$ARTIFACTS_DIR/python-summary.md`: the files you created,
the endpoints wrapped, and the final coverage number.
