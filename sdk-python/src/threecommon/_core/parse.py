"""Response parsing helpers. Pure functions over text bodies."""

from __future__ import annotations

import json
from datetime import datetime, timezone
from email.utils import parsedate_to_datetime
from typing import Any

import httpx


def parse_success_body(body_text: str) -> Any:
    """Decode a 2xx body. Empty or non-JSON resolves to ``None``."""
    if not body_text:
        return None
    try:
        return json.loads(body_text)
    except json.JSONDecodeError:
        return None


def parse_error_body(body_text: str) -> tuple[str, str, dict[str, Any] | None]:
    """Trying to parse the API's standard ``{"error": {...}}`` envelope.

    Returns ``("", "", None)`` when the body is missing or malformed; callers
    fall back to status-based defaults.
    """
    if not body_text:
        return ("", "", None)
    try:
        envelope = json.loads(body_text)
    except json.JSONDecodeError:
        return ("", "", None)
    if not isinstance(envelope, dict):
        return ("", "", None)
    err = envelope.get("error")
    if not isinstance(err, dict):
        return ("", "", None)
    code = err.get("code", "") if isinstance(err.get("code"), str) else ""
    message = err.get("message", "") if isinstance(err.get("message"), str) else ""
    details = err.get("details") if isinstance(err.get("details"), dict) else None
    return (code, message, details)


def parse_retry_after(header: str | None) -> float | None:
    """Parse a ``Retry-After`` header into seconds.

    Accepts either a delta-seconds value or an HTTP-date. Returns ``None``
    when the header is missing or malformed; ``0`` when the date is in the
    past.
    """
    if not header:
        return None
    try:
        seconds = float(header)
    except ValueError:
        pass
    else:
        return max(seconds, 0.0)

    try:
        target = parsedate_to_datetime(header)
    except (TypeError, ValueError):
        return None

    now = datetime.now(tz=timezone.utc)
    if target.tzinfo is None:
        target = target.replace(tzinfo=timezone.utc)
    delta = (target - now).total_seconds()
    return max(delta, 0.0)


def request_id_of(response: httpx.Response) -> str | None:
    """Return the ``X-Request-Id`` header value, or ``None``."""
    value: str | None = response.headers.get("x-request-id")
    return value
