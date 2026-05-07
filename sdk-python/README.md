# `threecommon`

[![PyPI](https://img.shields.io/pypi/v/threecommon.svg)](https://pypi.org/project/threecommon/)
[![Python](https://img.shields.io/pypi/pyversions/threecommon.svg)](https://pypi.org/project/threecommon/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Python client for the 3Common Public API. Sync **and** async, fully type-checked, Pydantic v2 models.

## Install

```bash
pip install threecommon
```

Requires **Python ≥ 3.10**.

## Quick start (sync)

```python
from threecommon import ThreeCommon
from threecommon.events import ListParams, UpdateBody

with ThreeCommon(api_key="3co_...") as client:
    # List
    result = client.events.list(ListParams(status="open", page_size=50))

    # Retrieve
    ev = client.events.retrieve("evt_123")

    # Update
    updated = client.events.update("evt_123", UpdateBody(name="New name"))

    # Auto-paginate
    for ev in client.events.list_auto_paginate(ListParams(status="open")):
        print(ev.name)
```

## Quick start (async)

```python
import asyncio
from threecommon import AsyncThreeCommon
from threecommon.events import ListParams

async def main() -> None:
    async with AsyncThreeCommon(api_key="3co_...") as client:
        result = await client.events.list(ListParams(status="open"))
        async for ev in client.events.list_auto_paginate(ListParams(status="open")):
            print(ev.name)

asyncio.run(main())
```

The API key may also be supplied via the `THREECOMMON_API_KEY` environment variable.

## Configuration

```python
from threecommon import ThreeCommon, RetryDelay

client = ThreeCommon(
    api_key="3co_...",                        # required (or via env var)
    base_url="https://api.3common.com",       # default
    api_version="2026-04-29",                 # pinned API version
    timeout_seconds=30.0,                     # per-request deadline
    max_retries=3,                            # automatic retries on 408/425/429/5xx
    retry_delay=RetryDelay(
        initial_seconds=0.5,
        max_seconds=8.0,
        jitter=True,
    ),
    telemetry=True,                           # opt-out of anonymous telemetry
)
```

## Error handling

Every error raised by the SDK inherits from `threecommon.APIError`. Catch the typed subclass you care about:

```python
from threecommon import (
    NotFoundError,
    AuthError,
    RateLimitError,
    ConnectionError,
)

try:
    client.events.retrieve("evt_missing")
except NotFoundError as e:
    # 404 — e.request_id, e.code, e.details
    ...
except AuthError as e:
    # 401 — bad or expired API key
    ...
except RateLimitError as e:
    # 429 — e.retry_after_seconds tells you when to retry
    ...
except ConnectionError as e:
    # network error; original cause via e.__cause__
    ...
```

Every error carries `code`, `message`, `http_status`, `request_id`, `details`, and `raw_response`. The default `str(e)` format includes the request ID for log correlation:

```
[not_found] Event evt_missing not found (request_id=req-dfx-abc)
```

## Pagination

Two flavors:

```python
# One page at a time
result = client.events.list(ListParams(page_size=50))

# All pages, lazy
for ev in client.events.list_auto_paginate(ListParams(status="open")):
    print(ev.name)

# Async
async for ev in async_client.events.list_auto_paginate(ListParams(status="open")):
    print(ev.name)
```

## Filters

The `filters` subpackage provides a typed builder for the API's `filters` query parameter — never write the JSON by hand:

```python
from threecommon import filters
from threecommon.events import ListParams

f = filters.and_(
    filters.field("status").is_any_of(["open"]),
    filters.field("ticket_sum").is_greater_than(10),
)

result = client.events.list(ListParams(filters=f.serialize()))
```

The full operator set is enumerated in `threecommon.filters.types`.

## Retries

Idempotent methods (`GET`, `PATCH`, `PUT`) retry automatically on `408`, `425`, `429`, `500`, `502`, `503`, `504` and on network errors. Backoff is exponential with full jitter, capped at `RetryDelay.max_seconds`. The SDK honors a server-provided `Retry-After` header on `429`.

`POST` and `DELETE` do not retry by default; pass an `idempotency_key` via per-request options to opt in (forward-compat — no v1 endpoints currently use this).

## Telemetry

The SDK sends a small, anonymized `Threecommon-Client-Telemetry` header on every request (SDK version, language, last-request latency). This helps debug performance reports from real customers without instrumenting their code. Disable globally:

```python
client = ThreeCommon(api_key="...", telemetry=False)
```

Or at runtime:

```python
client.disable_telemetry()
```

The header never contains your API key, request bodies, or response bodies.

## Repository layout

The package layout mirrors the Node and Go SDKs so behavior changes can be paired across languages:

```
sdk-python/
├── pyproject.toml
├── src/threecommon/
│   ├── __init__.py             # public surface re-exports
│   ├── py.typed                # PEP 561 marker
│   ├── client.py               # ThreeCommon + AsyncThreeCommon
│   ├── config.py               # ClientConfig, RetryDelay, defaults
│   ├── api_version.py          # pinned API version + path
│   ├── version.py              # SDK package version
│   ├── helpers.py              # small utility helpers
│   ├── errors/                 # exception tree
│   │   ├── base.py             # APIError
│   │   └── classes.py          # AuthError, NotFoundError, RateLimitError, ...
│   ├── pagination/             # auto-paginating iterators
│   │   └── auto_paginator.py   # Iter[T] + AsyncIter[T]
│   ├── filters/                # typed filter builder (shared across resources)
│   ├── events/                 # events resource (sync + async + Pydantic types)
│   ├── _core/                  # private HTTP machinery (decomposed)
│   └── _generated/             # datamodel-code-generator output (re-run via `make gen`)
├── examples/events/
└── tests/
```

## Development

The project uses [`uv`](https://github.com/astral-sh/uv) for venv + dependency management, [`ruff`](https://docs.astral.sh/ruff/) for lint and format, [`mypy`](https://mypy-lang.org/) and [`pyright`](https://github.com/microsoft/pyright) for type checking, and [`pytest`](https://docs.pytest.org/) for testing.

### Set up the dev environment

macOS and Linux:

```bash
# One-time: create a venv, activate, install runtime deps + dev tools
uv venv --python 3.10 .venv
source .venv/bin/activate
uv pip install -e ".[dev]"

# Verify:
uv pip list | grep threecommon       # should print: threecommon  0.0.0.dev0  /path/to/sdk-python
pytest -q                            # all tests pass
```

Windows:

```powershell
# One-time: create a venv, activate, install runtime deps + dev tools
uv venv --python 3.10 .venv
.\.venv\Scripts\activate.ps1
uv pip install -e ".[dev]"

# Verify:
uv pip list | Select-String threecommon # should print: threecommon  0.0.0.dev0  \path\to\sdk-python
pytest -q                               # all tests pass
```

If activation fails with an execution-policy error, run `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` once and retry.

Note: with the virtual environment active, all further bash snippets should work as-is in PowerShell on Windows, except where a separate PowerShell snippet is provided.

### Run tests

```bash
pytest                                              # all tests
pytest tests/test_events.py                         # one file
pytest tests/test_events.py::test_list_decodes_response   # one test
pytest -k "conformance"                             # match by name
pytest -q                                           # quiet output
```

The conformance harness (`tests/test_conformance.py`) parametrizes over the shared YAML scenarios at `../conformance/scenarios/*.yaml` and runs each one against both the sync and async clients (26 cases total).

### Coverage

```bash
pytest --cov=src/threecommon --cov-report=term      # term summary
pytest --cov=src/threecommon --cov-report=html      # HTML report at htmlcov/
pytest --cov=src/threecommon --cov-fail-under=90    # CI gate (≥ 90% line + branch)
```

### Lint and format

```bash
ruff check .                       # lint
ruff check --fix .                 # auto-fix
ruff format .                      # format
ruff format --check .              # CI-style check (no changes)
```

### Type check

Both run in CI; either failing blocks the PR.

```bash
mypy src/threecommon tests scripts        # mypy --strict via pyproject
pyright src/threecommon scripts           # pyright with project config
```

### Regenerate OpenAPI models

`src/threecommon/_generated/models.py` is produced from `../openapi/spec.yaml`. Re-run after every spec update:

macOS and Linux:

```bash
datamodel-codegen \
  --input ../openapi/spec.yaml \
  --input-file-type openapi \
  --output src/threecommon/_generated/models.py \
  --output-model-type pydantic_v2.BaseModel \
  --target-python-version 3.10 \
  --use-standard-collections --use-union-operator --use-double-quotes \
  --field-constraints --use-schema-description --capitalise-enum-members \
  --reuse-model --openapi-scopes paths schemas parameters
```

Windows:

```powershell
datamodel-codegen `
  --input ..\openapi\spec.yaml `
  --input-file-type openapi `
  --output .\src\threecommon\_generated\models.py `
  --output-model-type pydantic_v2.BaseModel `
  --target-python-version 3.10 `
  --use-standard-collections --use-union-operator --use-double-quotes `
  --field-constraints --use-schema-description --capitalise-enum-members `
  --reuse-model --openapi-scopes paths schemas parameters
```

The generated package is treated as a contract reference; customer-facing types are hand-curated under `src/threecommon/<resource>/types.py`.

### Live smoke (maintainer-only)

macOS and Linux:

```bash
THREECOMMON_API_KEY=3co_real_key \
SMOKE_EVENT_ID=evt_known \
python scripts/livesmoke.py
```

Windows:

```powershell
$env:THREECOMMON_API_KEY = "3co_real_key"
$env:SMOKE_EVENT_ID = "evt_known"
python .\scripts\livesmoke.py
```

Runs ≤ 10 real API calls and verifies the happy path + 401/404 error paths. Set `THREECOMMON_BASE_URL` to override the default `https://api.3common.com`.

### Build a wheel locally

```bash
uv build            # produces sdist + wheel under dist/
```

## Versioning

The SDK follows SemVer. The pinned **API version** (sent as `Threecommon-Version`) is independent — the API can evolve without breaking already-deployed SDKs. Bump `api_version` to opt into newer server behavior.

PyPI distribution: `threecommon`. Tags use the path-prefixed form `sdk-python/vX.Y.Z` to share the monorepo with the Node and Go SDKs.

## Examples

End-to-end runnable examples live under [`examples/events/`](./examples/events/):

```bash
python examples/events/list_sync.py
python examples/events/list_async.py
python examples/events/retrieve.py
python examples/events/update.py
python examples/events/auto_paginate.py
python examples/events/error_handling.py
python examples/events/filters_demo.py
```

Replace `3co_your_api_key_here` and `evt_replace_with_real_id` with real values before running.

## Contributing

See the [repository CONTRIBUTING guide](https://github.com/3-Common/sdk/blob/main/CONTRIBUTING.md). Issues and PRs welcome.

## License

[MIT](./LICENSE)
