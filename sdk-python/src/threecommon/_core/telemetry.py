"""Threecommon-Client-Telemetry header builder + last-request tracker.

The header carries SDK version, language, and the previous request's
latency snapshot.

both sync and async paths share one [Telemetry][threecommon._core.telemetry.Telemetry]
instance per client.
"""

from __future__ import annotations

import json
import threading
from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class _Snapshot:
    method: str
    path: str
    status: int
    duration_ms: int


class Telemetry:
    """Tracks one previous-request snapshot and emits the header value."""

    __slots__ = ("_enabled", "_last", "_lock")

    def __init__(self, *, enabled: bool) -> None:
        self._enabled = enabled
        self._last: _Snapshot | None = None
        self._lock = threading.Lock()

    @property
    def enabled(self) -> bool:
        with self._lock:
            return self._enabled

    def disable(self) -> None:
        """Turn telemetry off and clear the cached snapshot."""
        with self._lock:
            self._enabled = False
            self._last = None

    def record(self, *, method: str, path: str, status: int, duration_seconds: float) -> None:
        """Store a snapshot of the just-completed request. No-op when disabled."""
        with self._lock:
            if not self._enabled:
                return
            self._last = _Snapshot(
                method=method,
                path=path,
                status=status,
                duration_ms=int(duration_seconds * 1000),
            )

    def header_value(self, *, sdk_version: str, api_version: str) -> str | None:
        """Return the JSON header value, or ``None`` to omit the header."""
        with self._lock:
            if not self._enabled:
                return None
            last = self._last

        payload: dict[str, object] = {
            "lang": "python",
            "sdk": sdk_version,
            "api": api_version,
        }
        if last is not None:
            payload["last"] = {
                "m": last.method,
                "p": last.path,
                "s": last.status,
                "d": last.duration_ms,
            }
        return json.dumps(payload, separators=(",", ":"))
