"""HTTPClient (sync) + AsyncHTTPClient integration tests via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AuthError,
    ConflictError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ValidationError,
)
from threecommon._core.http_client import (
    AsyncHTTPClient,
    HTTPClient,
    HTTPClientOptions,
    Request,
)
from threecommon._core.retry import RetryPolicy
from threecommon._core.telemetry import Telemetry


def _opts(*, max_retries: int = 0) -> HTTPClientOptions:
    return HTTPClientOptions(
        api_key="3co_test",
        base_url="http://test.local",
        api_version="2026-04-29",
        timeout_seconds=5.0,
        retry=RetryPolicy(
            max_retries=max_retries, initial_seconds=0.0, max_seconds=0.0, jitter=False
        ),
        telemetry=Telemetry(enabled=False),
    )


# ────────────────────────────────────────────────────────────────────────────
# Sync
# ────────────────────────────────────────────────────────────────────────────


def test_sync_decodes_success_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        json={"data": [{"id": "evt_1"}], "hasMore": False},
    )
    c = HTTPClient(_opts())
    body = c.request(Request(method="GET", path="/events"))
    assert body == {"data": [{"id": "evt_1"}], "hasMore": False}
    c.close()


@pytest.mark.parametrize(
    ("status", "exc"),
    [
        (401, AuthError),
        (403, PermissionError),
        (404, NotFoundError),
        (409, ConflictError),
        (400, ValidationError),
        (422, ValidationError),
    ],
)
def test_sync_maps_typed_errors(
    httpx_mock: HTTPXMock,
    status: int,
    exc: type[Exception],
) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        status_code=status,
        json={"error": {"code": "x", "message": "boom"}},
    )
    c = HTTPClient(_opts())
    with pytest.raises(exc):
        c.request(Request(method="GET", path="/events"))
    c.close()


def test_sync_unknown_4xx_falls_back_to_validation_error(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", status_code=418, text="")
    c = HTTPClient(_opts())
    with pytest.raises(ValidationError):
        c.request(Request(method="GET", path="/events"))
    c.close()


def test_sync_500_is_server_error_when_retries_disabled(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        status_code=500,
        json={"error": {"code": "internal_error", "message": "boom"}},
    )
    c = HTTPClient(_opts())
    with pytest.raises(ServerError):
        c.request(Request(method="GET", path="/events"))
    c.close()


def test_sync_429_carries_retry_after(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        status_code=429,
        headers={"retry-after": "7"},
        json={"error": {"code": "rate_limit_exceeded", "message": "slow"}},
    )
    c = HTTPClient(_opts())
    with pytest.raises(RateLimitError) as exc_info:
        c.request(Request(method="GET", path="/events"))
    assert exc_info.value.retry_after_seconds == 7.0
    c.close()


def test_sync_retries_idempotent_on_500(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        status_code=500,
        json={"error": {"code": "internal_error", "message": "first"}},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        json={"data": [], "hasMore": False},
    )
    c = HTTPClient(_opts(max_retries=1))
    body = c.request(Request(method="GET", path="/events"))
    assert body == {"data": [], "hasMore": False}
    c.close()


def test_sync_does_not_retry_post_without_idempotency_key(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events",
        method="POST",
        status_code=500,
        json={"error": {"code": "internal_error", "message": "boom"}},
    )
    c = HTTPClient(_opts(max_retries=3))
    with pytest.raises(ServerError):
        c.request(Request(method="POST", path="/events", body={"x": 1}))
    c.close()


def test_sync_retries_post_with_idempotency_key(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", method="POST", status_code=502)
    httpx_mock.add_response(url="http://test.local/v1/events", method="POST", json={"ok": True})
    c = HTTPClient(_opts(max_retries=1))
    body = c.request(Request(method="POST", path="/events", body={"x": 1}, idempotency_key="key-1"))
    assert body == {"ok": True}
    c.close()


def test_sync_per_request_max_retries_override(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", status_code=500)
    c = HTTPClient(_opts(max_retries=5))
    with pytest.raises(ServerError):
        c.request(Request(method="GET", path="/events", max_retries=-1))
    c.close()


# ────────────────────────────────────────────────────────────────────────────
# Async
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_decodes_success_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", json={"data": [], "hasMore": False})
    c = AsyncHTTPClient(_opts())
    body = await c.request(Request(method="GET", path="/events"))
    assert body == {"data": [], "hasMore": False}
    await c.aclose()


@pytest.mark.asyncio
async def test_async_maps_404_to_not_found_error(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    c = AsyncHTTPClient(_opts())
    with pytest.raises(NotFoundError) as exc_info:
        await c.request(Request(method="GET", path="/events/evt_missing"))
    assert exc_info.value.code == "not_found"
    await c.aclose()


@pytest.mark.asyncio
async def test_async_retries_idempotent_on_503(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", status_code=503)
    httpx_mock.add_response(url="http://test.local/v1/events", json={"data": [], "hasMore": False})
    c = AsyncHTTPClient(_opts(max_retries=1))
    body = await c.request(Request(method="GET", path="/events"))
    assert body == {"data": [], "hasMore": False}
    await c.aclose()
