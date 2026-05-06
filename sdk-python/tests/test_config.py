"""Config validation tests."""

from __future__ import annotations

import pytest

from threecommon import (
    DEFAULT_BASE_URL,
    DEFAULT_MAX_RETRIES,
    DEFAULT_RETRY_DELAY,
    DEFAULT_TIMEOUT_SECONDS,
    RetryDelay,
    ValidationError,
)
from threecommon.config import resolve_config


def test_resolve_requires_api_key(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("THREECOMMON_API_KEY", raising=False)
    with pytest.raises(ValidationError) as exc_info:
        resolve_config(
            api_key=None,
            base_url=None,
            api_version=None,
            timeout_seconds=None,
            max_retries=None,
            retry_delay=None,
        )
    assert exc_info.value.code == "missing_api_key"


def test_resolve_picks_up_env_var(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("THREECOMMON_API_KEY", "3co_from_env")
    cfg = resolve_config(
        api_key=None,
        base_url=None,
        api_version=None,
        timeout_seconds=None,
        max_retries=None,
        retry_delay=None,
    )
    assert cfg.api_key == "3co_from_env"


def test_resolve_applies_defaults() -> None:
    cfg = resolve_config(
        api_key="k",
        base_url=None,
        api_version=None,
        timeout_seconds=None,
        max_retries=None,
        retry_delay=None,
    )
    assert cfg.base_url == DEFAULT_BASE_URL
    assert cfg.timeout_seconds == DEFAULT_TIMEOUT_SECONDS
    assert cfg.max_retries == DEFAULT_MAX_RETRIES
    assert cfg.retry_delay == DEFAULT_RETRY_DELAY
    assert cfg.telemetry is True


def test_resolve_rejects_invalid_base_url() -> None:
    with pytest.raises(ValidationError) as exc_info:
        resolve_config(
            api_key="k",
            base_url="not-a-url",
            api_version=None,
            timeout_seconds=None,
            max_retries=None,
            retry_delay=None,
        )
    assert exc_info.value.code == "invalid_base_url"


def test_resolve_rejects_negative_timeout() -> None:
    with pytest.raises(ValidationError) as exc_info:
        resolve_config(
            api_key="k",
            base_url=None,
            api_version=None,
            timeout_seconds=-1.0,
            max_retries=None,
            retry_delay=None,
        )
    assert exc_info.value.code == "invalid_timeout"


def test_resolve_rejects_negative_max_retries() -> None:
    with pytest.raises(ValidationError) as exc_info:
        resolve_config(
            api_key="k",
            base_url=None,
            api_version=None,
            timeout_seconds=None,
            max_retries=-1,
            retry_delay=None,
        )
    assert exc_info.value.code == "invalid_max_retries"


def test_resolve_explicit_zero_retries_allowed() -> None:
    cfg = resolve_config(
        api_key="k",
        base_url=None,
        api_version=None,
        timeout_seconds=None,
        max_retries=0,
        retry_delay=None,
    )
    assert cfg.max_retries == 0


def test_resolve_trims_trailing_slash() -> None:
    cfg = resolve_config(
        api_key="k",
        base_url="https://api.3common.com//",
        api_version=None,
        timeout_seconds=None,
        max_retries=None,
        retry_delay=None,
    )
    assert cfg.base_url == "https://api.3common.com"


def test_resolve_respects_custom_retry_delay() -> None:
    custom = RetryDelay(initial_seconds=0.1, max_seconds=2.0, jitter=False)
    cfg = resolve_config(
        api_key="k",
        base_url=None,
        api_version=None,
        timeout_seconds=None,
        max_retries=None,
        retry_delay=custom,
    )
    assert cfg.retry_delay == custom
