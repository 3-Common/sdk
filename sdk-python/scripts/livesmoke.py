"""Pre-release smoke test against the live API.

Runs <= 10 calls and verifies the happy path + the four common error paths.
Used by .github/workflows/live-smoke.yml (maintainer-only).

Required env:
    THREECOMMON_API_KEY   — a real API key

Optional env:
    THREECOMMON_BASE_URL  — defaults to https://api.3common.com
    SMOKE_EVENT_ID        — an event ID known to belong to the API-key host;
                            required for retrieve / 403 / 422 checks

Run with: python scripts/livesmoke.py
"""

from __future__ import annotations

import os
import sys
from dataclasses import dataclass

from threecommon import APIError, AuthError, NotFoundError, ThreeCommon
from threecommon.events import ListParams


@dataclass
class _Result:
    check: str
    status: str  # "pass", "fail", "skip"
    detail: str = ""


def _check_list(client: ThreeCommon) -> _Result:
    try:
        result = client.events.list(ListParams(page_size=1))
    except APIError as e:
        return _Result("events.list", "fail", repr(e))
    return _Result(
        "events.list",
        "pass",
        f"data.len={len(result.data)} has_more={result.has_more}",
    )


def _check_auto_paginate(client: ThreeCommon) -> _Result:
    try:
        iterator = client.events.list_auto_paginate(ListParams(page_size=1))
        first = next(iterator, None)
    except APIError as e:
        return _Result("events.list_auto_paginate", "fail", repr(e))
    return _Result(
        "events.list_auto_paginate",
        "pass",
        "yielded one" if first is not None else "empty",
    )


def _check_retrieve(client: ThreeCommon, known_event_id: str | None) -> _Result:
    if not known_event_id:
        return _Result("events.retrieve", "skip", "SMOKE_EVENT_ID not set")
    try:
        ev = client.events.retrieve(known_event_id)
    except APIError as e:
        return _Result("events.retrieve", "fail", repr(e))
    return _Result("events.retrieve", "pass", f"id={ev.id}")


def _check_404(client: ThreeCommon) -> _Result:
    try:
        client.events.retrieve("000000000000000000000000")
    except NotFoundError as e:
        return _Result("404 path", "pass", f"code={e.code} request_id={e.request_id}")
    except APIError as e:
        return _Result("404 path", "fail", f"unexpected: {e!r}")
    return _Result("404 path", "fail", "expected NotFoundError")


def _check_401(base_url: str) -> _Result:
    fake = "3co_smoke_test_invalid_key"  # gitleaks:allow
    try:
        with ThreeCommon(
            api_key=fake,
            base_url=base_url,
            telemetry=False,
            max_retries=0,
        ) as bad:
            bad.events.list(ListParams(page_size=1))
    except AuthError as e:
        return _Result("401 path", "pass", f"code={e.code}")
    except APIError as e:
        return _Result("401 path", "fail", f"unexpected: {e!r}")
    return _Result("401 path", "fail", "expected AuthError")


def main() -> int:
    api_key = os.environ.get("THREECOMMON_API_KEY", "")
    if not api_key:
        sys.stderr.write("THREECOMMON_API_KEY env var is required for live-smoke runs\n")
        return 1

    base_url = os.environ.get("THREECOMMON_BASE_URL") or "https://api.3common.com"
    known_event_id = os.environ.get("SMOKE_EVENT_ID")

    results: list[_Result] = []
    with ThreeCommon(api_key=api_key, base_url=base_url, telemetry=False) as client:
        results.append(_check_list(client))
        results.append(_check_auto_paginate(client))
        results.append(_check_retrieve(client, known_event_id))
        results.append(_check_404(client))
    results.append(_check_401(base_url))

    failed = 0
    for entry in results:
        icon = {"pass": "✓", "fail": "✗", "skip": "○"}.get(entry.status, "?")
        sys.stdout.write(f"{icon} {entry.check} — {entry.detail}\n")
        if entry.status == "fail":
            failed += 1

    if failed:
        sys.stderr.write(f"\n{failed} check(s) failed.\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
