from __future__ import annotations

import pytest

from threecommon._core.retry import (
    RetryPolicy,
    compute_backoff,
    is_idempotent,
    is_retryable_status,
)
from threecommon.config import RetryDelay


@pytest.mark.parametrize(
    ("method", "has_key", "want"),
    [
        ("GET", False, True),
        ("PATCH", False, True),
        ("PUT", False, True),
        ("POST", False, False),
        ("DELETE", False, False),
        ("POST", True, True),
        ("DELETE", True, True),
        ("get", False, True),  # case-insensitive
    ],
)
def test_is_idempotent(method: str, has_key: bool, want: bool) -> None:  # noqa: FBT001
    assert is_idempotent(method, has_idempotency_key=has_key) is want


@pytest.mark.parametrize("status", [408, 425, 429, 500, 502, 503, 504])
def test_retryable_statuses(status: int) -> None:
    assert is_retryable_status(status)


@pytest.mark.parametrize("status", [200, 301, 400, 401, 404, 422, 501])
def test_non_retryable_statuses(status: int) -> None:
    assert not is_retryable_status(status)


def test_retry_after_takes_precedence() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0.1, max_seconds=2.0, jitter=False)
    assert compute_backoff(attempt=0, retry_after_seconds=0.5, policy=policy) == 0.5


def test_retry_after_capped_at_max() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0.1, max_seconds=1.0, jitter=False)
    assert compute_backoff(attempt=0, retry_after_seconds=10.0, policy=policy) == 1.0


def test_exponential_no_jitter() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0.1, max_seconds=2.0, jitter=False)
    assert compute_backoff(attempt=0, retry_after_seconds=None, policy=policy) == pytest.approx(0.1)
    assert compute_backoff(attempt=1, retry_after_seconds=None, policy=policy) == pytest.approx(0.2)
    assert compute_backoff(attempt=4, retry_after_seconds=None, policy=policy) == pytest.approx(1.6)
    # Capped at max
    assert compute_backoff(attempt=10, retry_after_seconds=None, policy=policy) == pytest.approx(
        2.0
    )


def test_jitter_within_bounds() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0.1, max_seconds=2.0, jitter=True)
    for _ in range(50):
        got = compute_backoff(attempt=2, retry_after_seconds=None, policy=policy)
        assert 0 <= got < 0.4


def test_negative_attempt_clamped() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0.1, max_seconds=2.0, jitter=False)
    assert compute_backoff(attempt=-1, retry_after_seconds=None, policy=policy) == pytest.approx(
        0.1
    )


def test_jitter_zero_returns_zero() -> None:
    policy = RetryPolicy(max_retries=3, initial_seconds=0, max_seconds=0, jitter=True)
    assert compute_backoff(attempt=0, retry_after_seconds=None, policy=policy) == 0.0


def test_from_delay_factory() -> None:
    delay = RetryDelay(initial_seconds=0.5, max_seconds=8.0, jitter=True)
    policy = RetryPolicy.from_delay(3, delay)
    assert policy.max_retries == 3
    assert policy.initial_seconds == 0.5
    assert policy.max_seconds == 8.0
    assert policy.jitter is True
