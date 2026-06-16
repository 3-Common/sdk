"""Sync + async request orchestrators.

Both classes wrap an [httpx][https://www.python-httpx.org/] client and
compose the pure modules in this folder into a complete request lifecycle:
build URL → build headers → send → parse → map errors → retry. The
sync/async split is at the I/O boundary only; everything else is shared.
"""

from __future__ import annotations

import asyncio
import logging
import time
from dataclasses import dataclass, field
from http import HTTPStatus
from typing import Any

import httpx

from threecommon._core.headers import build_headers
from threecommon._core.parse import (
    parse_error_body,
    parse_retry_after,
    parse_success_body,
    request_id_of,
)
from threecommon._core.retry import (
    RetryPolicy,
    compute_backoff,
    is_idempotent,
    is_retryable_status,
)
from threecommon._core.telemetry import Telemetry
from threecommon._core.url import build_url
from threecommon.api_version import API_PATH
from threecommon.errors.base import APIError
from threecommon.errors.classes import (
    AuthError,
    ConflictError,
    ConnectionError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ValidationError,
)
from threecommon.version import VERSION


@dataclass(slots=True)
class Request:
    """One logical SDK call. The HTTP clients fill in the rest."""

    method: str
    path: str
    query: dict[str, str] | None = None
    body: dict[str, Any] | None = None
    idempotency_key: str | None = None
    timeout_seconds: float | None = None
    max_retries: int | None = None  # negative → disable retries for this call


@dataclass(slots=True)
class _Resolved:
    """Pre-resolved per-call values shared by sync + async paths."""

    url: str
    method: str
    body: dict[str, Any] | None
    idempotency_key: str | None
    max_retries: int
    timeout_seconds: float
    is_idempotent: bool


# ────────────────────────────────────────────────────────────────────────────
# Common (pure) helpers
# ────────────────────────────────────────────────────────────────────────────


def _resolve(
    req: Request,
    *,
    base_url: str,
    api_version_header: str,
    default_timeout: float,
    default_max_retries: int,
) -> _Resolved:
    _ = api_version_header
    max_retries = default_max_retries
    if req.max_retries is not None:
        max_retries = 0 if req.max_retries < 0 else req.max_retries
    return _Resolved(
        url=build_url(base_url=base_url, api_path=API_PATH, path=req.path, query=req.query),
        method=req.method.upper(),
        body=req.body,
        idempotency_key=req.idempotency_key,
        max_retries=max_retries,
        timeout_seconds=req.timeout_seconds if req.timeout_seconds is not None else default_timeout,
        is_idempotent=is_idempotent(
            req.method.upper(), has_idempotency_key=req.idempotency_key is not None
        ),
    )


# Status-code -> typed-exception mapping. ValidationError is the catch-all
# for any unmapped 4xx; ServerError covers >= 500.
_STATUS_TO_ERROR: dict[int, type[APIError]] = {
    HTTPStatus.UNAUTHORIZED: AuthError,
    HTTPStatus.FORBIDDEN: PermissionError,
    HTTPStatus.NOT_FOUND: NotFoundError,
    HTTPStatus.CONFLICT: ConflictError,
    HTTPStatus.BAD_REQUEST: ValidationError,
    HTTPStatus.UNPROCESSABLE_ENTITY: ValidationError,
}

# Status-code -> SDK error.code default. Used when the API didn't return a
# parsable error envelope.
_STATUS_TO_CODE: dict[int, str] = {
    HTTPStatus.UNAUTHORIZED: "unauthorized",
    HTTPStatus.FORBIDDEN: "forbidden",
    HTTPStatus.NOT_FOUND: "not_found",
    HTTPStatus.CONFLICT: "conflict",
    HTTPStatus.TOO_MANY_REQUESTS: "rate_limit_exceeded",
}


def _map_error_response(response: httpx.Response, retry_after: float | None) -> APIError:
    """Turn a non-2xx response into the typed exception subclass."""
    code, message, details = parse_error_body(response.text)
    if not code:
        code = _default_code_for_status(response.status_code)
    if not message:
        message = f"Request failed with status {response.status_code}"

    base_kwargs: dict[str, Any] = {
        "code": code,
        "message": message,
        "http_status": response.status_code,
        "request_id": request_id_of(response),
        "details": details,
        "raw_response": response.text or None,
    }

    status = response.status_code
    if status == HTTPStatus.TOO_MANY_REQUESTS:
        return RateLimitError(**base_kwargs, retry_after_seconds=retry_after)
    if status in _STATUS_TO_ERROR:
        return _STATUS_TO_ERROR[status](**base_kwargs)
    if status >= HTTPStatus.INTERNAL_SERVER_ERROR:
        return ServerError(**base_kwargs)
    return ValidationError(**base_kwargs)


def _wrap_connection(message: str, cause: BaseException) -> ConnectionError:
    return ConnectionError(code="connection_error", message=message, cause=cause)


def _default_code_for_status(status: int) -> str:
    code = _STATUS_TO_CODE.get(status)
    if code is not None:
        return code
    if status >= HTTPStatus.INTERNAL_SERVER_ERROR:
        return "internal_error"
    return "request_failed"


def _build_request_headers(
    *,
    api_key: str,
    api_version: str,
    telemetry: Telemetry,
    idempotency_key: str | None,
    user_agent_extra: str | None,
    has_body: bool,
) -> dict[str, str]:
    return build_headers(
        api_key=api_key,
        api_version=api_version,
        sdk_version=VERSION,
        user_agent_extra=user_agent_extra,
        telemetry_header=telemetry.header_value(sdk_version=VERSION, api_version=api_version),
        idempotency_key=idempotency_key,
        has_body=has_body,
    )


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


@dataclass(slots=True)
class HTTPClientOptions:
    """Configuration accepted by :class:`HTTPClient` / :class:`AsyncHTTPClient`."""

    api_key: str
    base_url: str
    api_version: str
    timeout_seconds: float
    retry: RetryPolicy
    telemetry: Telemetry
    logger: logging.Logger | None = None
    user_agent_extra: str | None = None
    httpx_client: httpx.Client | None = field(default=None, repr=False)
    async_httpx_client: httpx.AsyncClient | None = field(default=None, repr=False)


class HTTPClient:
    """Sync request orchestrator. One instance per [ThreeCommon] client."""

    __slots__ = ("_opts", "_owns_httpx", "httpx")

    def __init__(self, opts: HTTPClientOptions) -> None:
        self._opts = opts
        if opts.httpx_client is not None:
            self.httpx = opts.httpx_client
            self._owns_httpx = False
        else:
            self.httpx = httpx.Client(timeout=opts.timeout_seconds)
            self._owns_httpx = True

    def close(self) -> None:
        """Close the underlying httpx client if we created it."""
        if self._owns_httpx:
            self.httpx.close()

    def request(self, req: Request) -> Any:
        """Send a request honoring the client's retry policy.

        Returns the decoded JSON body for 2xx responses, or raises a
        [APIError][threecommon.APIError] subclass.
        """
        resolved = _resolve(
            req,
            base_url=self._opts.base_url,
            api_version_header=self._opts.api_version,
            default_timeout=self._opts.timeout_seconds,
            default_max_retries=self._opts.retry.max_retries,
        )

        attempt = 0
        while True:
            headers = _build_request_headers(
                api_key=self._opts.api_key,
                api_version=self._opts.api_version,
                telemetry=self._opts.telemetry,
                idempotency_key=resolved.idempotency_key,
                user_agent_extra=self._opts.user_agent_extra,
                has_body=resolved.body is not None,
            )

            start = time.monotonic()
            try:
                response = self.httpx.request(
                    method=resolved.method,
                    url=resolved.url,
                    headers=headers,
                    json=resolved.body,
                    timeout=resolved.timeout_seconds,
                )
            except (httpx.TimeoutException, httpx.NetworkError, httpx.ProtocolError) as exc:
                if resolved.is_idempotent and attempt < resolved.max_retries:
                    time.sleep(
                        compute_backoff(
                            attempt=attempt, retry_after_seconds=None, policy=self._opts.retry
                        )
                    )
                    attempt += 1
                    continue
                raise _wrap_connection(str(exc) or "network error", exc) from exc

            duration = time.monotonic() - start
            self._opts.telemetry.record(
                method=resolved.method,
                path=req.path,
                status=response.status_code,
                duration_seconds=duration,
            )
            if self._opts.logger is not None:
                self._opts.logger.debug(
                    "threecommon:request",
                    extra={
                        "method": resolved.method,
                        "path": req.path,
                        "status": response.status_code,
                        "duration_ms": int(duration * 1000),
                        "request_id": request_id_of(response),
                        "attempt": attempt,
                    },
                )

            if response.is_success:
                return parse_success_body(response.text)

            retry_after = parse_retry_after(response.headers.get("retry-after"))
            if (
                resolved.is_idempotent
                and attempt < resolved.max_retries
                and is_retryable_status(response.status_code)
            ):
                time.sleep(
                    compute_backoff(
                        attempt=attempt, retry_after_seconds=retry_after, policy=self._opts.retry
                    )
                )
                attempt += 1
                continue

            raise _map_error_response(response, retry_after)


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


class AsyncHTTPClient:
    """Async request orchestrator. One instance per [AsyncThreeCommon] client."""

    __slots__ = ("_opts", "_owns_httpx", "httpx")

    def __init__(self, opts: HTTPClientOptions) -> None:
        self._opts = opts
        if opts.async_httpx_client is not None:
            self.httpx = opts.async_httpx_client
            self._owns_httpx = False
        else:
            self.httpx = httpx.AsyncClient(timeout=opts.timeout_seconds)
            self._owns_httpx = True

    async def aclose(self) -> None:
        """Close the underlying httpx client if we created it."""
        if self._owns_httpx:
            await self.httpx.aclose()

    async def request(self, req: Request) -> Any:
        """Send a request honoring the client's retry policy.

        Async variant of [HTTPClient.request].
        """
        resolved = _resolve(
            req,
            base_url=self._opts.base_url,
            api_version_header=self._opts.api_version,
            default_timeout=self._opts.timeout_seconds,
            default_max_retries=self._opts.retry.max_retries,
        )

        attempt = 0
        while True:
            headers = _build_request_headers(
                api_key=self._opts.api_key,
                api_version=self._opts.api_version,
                telemetry=self._opts.telemetry,
                idempotency_key=resolved.idempotency_key,
                user_agent_extra=self._opts.user_agent_extra,
                has_body=resolved.body is not None,
            )

            start = time.monotonic()
            try:
                response = await self.httpx.request(
                    method=resolved.method,
                    url=resolved.url,
                    headers=headers,
                    json=resolved.body,
                    timeout=resolved.timeout_seconds,
                )
            except (httpx.TimeoutException, httpx.NetworkError, httpx.ProtocolError) as exc:
                if resolved.is_idempotent and attempt < resolved.max_retries:
                    await asyncio.sleep(
                        compute_backoff(
                            attempt=attempt, retry_after_seconds=None, policy=self._opts.retry
                        ),
                    )
                    attempt += 1
                    continue
                raise _wrap_connection(str(exc) or "network error", exc) from exc

            duration = time.monotonic() - start
            self._opts.telemetry.record(
                method=resolved.method,
                path=req.path,
                status=response.status_code,
                duration_seconds=duration,
            )
            if self._opts.logger is not None:
                self._opts.logger.debug(
                    "threecommon:request",
                    extra={
                        "method": resolved.method,
                        "path": req.path,
                        "status": response.status_code,
                        "duration_ms": int(duration * 1000),
                        "request_id": request_id_of(response),
                        "attempt": attempt,
                    },
                )

            if response.is_success:
                return parse_success_body(response.text)

            retry_after = parse_retry_after(response.headers.get("retry-after"))
            if (
                resolved.is_idempotent
                and attempt < resolved.max_retries
                and is_retryable_status(response.status_code)
            ):
                await asyncio.sleep(
                    compute_backoff(
                        attempt=attempt,
                        retry_after_seconds=retry_after,
                        policy=self._opts.retry,
                    ),
                )
                attempt += 1
                continue

            raise _map_error_response(response, retry_after)


__all__ = (
    "AsyncHTTPClient",
    "HTTPClient",
    "HTTPClientOptions",
    "Request",
)
