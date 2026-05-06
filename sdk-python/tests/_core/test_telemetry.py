from __future__ import annotations

import json
import threading

from threecommon._core.telemetry import Telemetry


def test_disabled_returns_no_header() -> None:
    t = Telemetry(enabled=False)
    assert not t.enabled
    assert t.header_value(sdk_version="0.1.0", api_version="2026-04-29") is None


def test_enabled_no_last_emits_baseline_payload() -> None:
    t = Telemetry(enabled=True)
    got = t.header_value(sdk_version="0.1.0", api_version="2026-04-29")
    assert got is not None
    payload = json.loads(got)
    assert payload == {"lang": "python", "sdk": "0.1.0", "api": "2026-04-29"}


def test_record_populates_last() -> None:
    t = Telemetry(enabled=True)
    t.record(method="GET", path="/events", status=200, duration_seconds=0.123)
    got = t.header_value(sdk_version="0.1.0", api_version="2026-04-29")
    assert got is not None
    payload = json.loads(got)
    assert payload["last"] == {"m": "GET", "p": "/events", "s": 200, "d": 123}


def test_disable_clears_state() -> None:
    t = Telemetry(enabled=True)
    t.record(method="GET", path="/events", status=200, duration_seconds=1)
    t.disable()
    assert not t.enabled
    assert t.header_value(sdk_version="0.1.0", api_version="2026-04-29") is None


def test_record_when_disabled_is_noop() -> None:
    t = Telemetry(enabled=False)
    t.record(method="GET", path="/events", status=200, duration_seconds=1)
    assert t.header_value(sdk_version="0.1.0", api_version="2026-04-29") is None


def test_concurrent_record_and_header_value() -> None:
    """Smoke-test thread safety — no assertion beyond not raising."""
    t = Telemetry(enabled=True)

    def writer() -> None:
        for _ in range(100):
            t.record(method="GET", path="/events", status=200, duration_seconds=0.001)

    def reader() -> None:
        for _ in range(100):
            t.header_value(sdk_version="0.1.0", api_version="2026-04-29")

    threads = [threading.Thread(target=writer) for _ in range(4)] + [
        threading.Thread(target=reader) for _ in range(4)
    ]
    for thr in threads:
        thr.start()
    for thr in threads:
        thr.join()
