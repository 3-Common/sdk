"""Retry policy + backoff math.

Pure module — no I/O, no timing. The sync/async HTTP clients call
[compute_backoff][threecommon._core.retry.compute_backoff] and pass the
result to ``time.sleep`` / ``asyncio.sleep`` themselves.
"""

from __future__ import annotations

import secrets
from dataclasses import dataclass

from threecommon.config import RetryDelay

#: HTTP status codes the SDK considers retryable for idempotent methods.
RETRYABLE_STATUSES: frozenset[int] = frozenset({408, 425, 429, 500, 502, 503, 504})

#: HTTP methods the SDK retries automatically. ``POST`` and ``DELETE`` opt
#: in via an explicit idempotency key.
IDEMPOTENT_METHODS: frozenset[str] = frozenset({"GET", "PATCH", "PUT"})


@dataclass(frozen=True, slots=True)
class RetryPolicy:
    """Bundled retry configuration for the HTTP client."""

    max_retries: int
    initial_seconds: float
    max_seconds: float
    jitter: bool

    @classmethod
    def from_delay(cls, max_retries: int, delay: RetryDelay) -> RetryPolicy:
        """Build a [RetryPolicy] from the public [RetryDelay]."""
        return cls(
            max_retries=max_retries,
            initial_seconds=delay.initial_seconds,
            max_seconds=delay.max_seconds,
            jitter=delay.jitter,
        )


def is_idempotent(method: str, *, has_idempotency_key: bool) -> bool:
    """Whether the SDK may safely retry a request with this method."""
    return has_idempotency_key or method.upper() in IDEMPOTENT_METHODS


def is_retryable_status(status: int) -> bool:
    """Whether ``status`` is in the retry set (alongside method idempotency)."""
    return status in RETRYABLE_STATUSES


def compute_backoff(
    *,
    attempt: int,
    retry_after_seconds: float | None,
    policy: RetryPolicy,
) -> float:
    """Return the next sleep duration in seconds.

    When ``retry_after_seconds`` is provided (e.g. parsed from a
    ``Retry-After`` header) it takes precedence, capped at
    ``policy.max_seconds``. Otherwise: exponential ``2**attempt * initial``,
    capped, with optional full-jitter randomization.
    """
    if retry_after_seconds is not None and retry_after_seconds >= 0:
        return min(retry_after_seconds, policy.max_seconds)

    safe_attempt = max(attempt, 0)
    exp: float = policy.initial_seconds * (2**safe_attempt)
    capped: float = min(exp, policy.max_seconds)

    if not policy.jitter:
        return capped
    if capped <= 0:
        return 0.0
    # Full jitter: pick uniformly in [0, capped). secrets.randbits avoids
    # the math.random global-state lock and isn't security-sensitive here.
    bits = secrets.randbits(32)
    return float(capped * (bits / (1 << 32)))
