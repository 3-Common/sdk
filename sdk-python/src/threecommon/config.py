"""Resolved client configuration.

User do not construct [ClientConfig][threecommon.config.ClientConfig]
directly — they pass keyword arguments to [ThreeCommon][threecommon.ThreeCommon]
or [AsyncThreeCommon][threecommon.AsyncThreeCommon], which build a
``ClientConfig`` internally after validation.

This module also exposes the named defaults so the public docs can reference
them.
"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from typing import TYPE_CHECKING

from threecommon.api_version import API_VERSION
from threecommon.errors.classes import ValidationError

if TYPE_CHECKING:  # pragma: no cover
    import logging

    import httpx


#: Default API base URL.
DEFAULT_BASE_URL = "https://api.3common.com"

#: Default per-request deadline.
DEFAULT_TIMEOUT_SECONDS = 30.0

#: Default number of retry attempts for idempotent requests on retryable
#: failures (408, 425, 429, 5xx, network errors).
DEFAULT_MAX_RETRIES = 3

#: Environment variable consulted when ``api_key`` is not passed explicitly.
ENV_VAR_API_KEY = "THREECOMMON_API_KEY"


@dataclass(frozen=True, slots=True)
class RetryDelay:
    """Exponential-backoff schedule.

    Backoff doubles each attempt, capped at :attr:`max_seconds`. When
    :attr:`jitter` is true the SDK picks a random value in
    ``[0, capped]`` for the actual sleep.
    """

    initial_seconds: float = 0.5
    max_seconds: float = 8.0
    jitter: bool = True


#: Default backoff schedule applied when ``retry_delay`` isn't passed.
DEFAULT_RETRY_DELAY = RetryDelay()


@dataclass(frozen=True, slots=True)
class ClientConfig:
    """Internal frozen view of the resolved client configuration.

    Construct via :func:`resolve_config` rather than directly.
    """

    api_key: str
    base_url: str
    api_version: str
    timeout_seconds: float
    max_retries: int
    retry_delay: RetryDelay
    http_client: httpx.Client | None = None
    async_http_client: httpx.AsyncClient | None = None
    logger: logging.Logger | None = None
    telemetry: bool = True
    user_agent_extra: str | None = field(default=None, repr=False)


def resolve_config(
    *,
    api_key: str | None,
    base_url: str | None,
    api_version: str | None,
    timeout_seconds: float | None,
    max_retries: int | None,
    retry_delay: RetryDelay | None,
    http_client: httpx.Client | None = None,
    async_http_client: httpx.AsyncClient | None = None,
    logger: logging.Logger | None = None,
    telemetry: bool | None = None,
) -> ClientConfig:
    """Validate constructor kwargs and return a frozen [ClientConfig].

    Raises [ValidationError][threecommon.ValidationError] for missing API
    key or invalid numeric ranges; the message names the exact field so
    customers don't have to read a stack trace.
    """
    resolved_key = api_key or os.environ.get(ENV_VAR_API_KEY, "")
    if not resolved_key:
        raise ValidationError(
            code="missing_api_key",
            message=(
                "An API key is required. Pass `api_key` to the ThreeCommon "
                f"constructor, or set the {ENV_VAR_API_KEY} environment variable."
            ),
        )

    resolved_base = (base_url or DEFAULT_BASE_URL).rstrip("/")
    if not resolved_base.startswith(("http://", "https://")):
        raise ValidationError(
            code="invalid_base_url",
            message=f"base_url must start with http:// or https://; got {base_url!r}.",
        )

    resolved_timeout = timeout_seconds if timeout_seconds is not None else DEFAULT_TIMEOUT_SECONDS
    if resolved_timeout <= 0:
        raise ValidationError(
            code="invalid_timeout",
            message=f"timeout_seconds must be positive; got {resolved_timeout!r}.",
        )

    resolved_retries = max_retries if max_retries is not None else DEFAULT_MAX_RETRIES
    if resolved_retries < 0:
        raise ValidationError(
            code="invalid_max_retries",
            message=f"max_retries must be non-negative; got {resolved_retries!r}.",
        )

    return ClientConfig(
        api_key=resolved_key,
        base_url=resolved_base,
        api_version=api_version or API_VERSION,
        timeout_seconds=resolved_timeout,
        max_retries=resolved_retries,
        retry_delay=retry_delay or DEFAULT_RETRY_DELAY,
        http_client=http_client,
        async_http_client=async_http_client,
        logger=logger,
        telemetry=True if telemetry is None else telemetry,
    )
