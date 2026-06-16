"""Header builder. Pure function over already-resolved values."""

from __future__ import annotations

import platform
import sys


def user_agent_suffix(extra: str | None = None) -> str:
    """Runtime + OS portion of the ``User-Agent`` header."""
    parts = [
        f"Python/{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        f"{platform.system()}-{platform.machine()}",
    ]
    if extra:
        parts.append(extra)
    return "; ".join(parts)


def build_headers(
    *,
    api_key: str,
    api_version: str,
    sdk_version: str,
    user_agent_extra: str | None = None,
    telemetry_header: str | None = None,
    idempotency_key: str | None = None,
    has_body: bool = True,
) -> dict[str, str]:
    """Return a fresh header dict populated with every header the SDK sends."""
    headers: dict[str, str] = {
        "Authorization": f"Bearer {api_key}",
        "Threecommon-Version": api_version,
        "User-Agent": f"ThreeCommonPython/{sdk_version} ({user_agent_suffix(user_agent_extra)})",
        "Accept": "application/json",
    }
    # Bodyless requests (DELETE, action-style POSTs like finalize/auto_charge)
    # must not advertise a JSON body: a server enforcing Content-Type against
    # an empty body rejects them with a 400.
    if has_body:
        headers["Content-Type"] = "application/json"
    if telemetry_header:
        headers["Threecommon-Client-Telemetry"] = telemetry_header
    if idempotency_key:
        headers["Idempotency-Key"] = idempotency_key
    return headers
