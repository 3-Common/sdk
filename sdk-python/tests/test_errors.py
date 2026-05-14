"""Errors module unit tests."""

from __future__ import annotations

from threecommon import (
    APIError,
    AuthError,
    ConflictError,
    ConnectionError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ValidationError,
)


def test_apierror_format_with_request_id() -> None:
    e = APIError(code="not_found", message="Event evt_1 not found", request_id="req-1")
    assert str(e) == "[not_found] Event evt_1 not found (request_id=req-1)"
    assert e.request_id == "req-1"
    assert e.code == "not_found"


def test_apierror_format_without_request_id() -> None:
    e = APIError(code="request_failed", message="boom")
    assert str(e) == "[request_failed] boom"


def test_apierror_unwrap_via_cause() -> None:
    cause = RuntimeError("dial tcp: timeout")
    e = APIError(code="connection_error", message="boom", cause=cause)
    assert e.__cause__ is cause


def test_apierror_repr_contains_status() -> None:
    e = NotFoundError(code="not_found", message="missing", http_status=404, request_id="req-x")
    text = repr(e)
    assert "NotFoundError" in text
    assert "404" in text
    assert "req-x" in text


def test_typed_subclasses_inherit_from_apierror() -> None:
    for cls in (
        AuthError,
        PermissionError,
        NotFoundError,
        ValidationError,
        ConflictError,
        RateLimitError,
        ServerError,
        ConnectionError,
    ):
        assert issubclass(cls, APIError)


def test_rate_limit_error_carries_retry_after() -> None:
    e = RateLimitError(
        code="rate_limit_exceeded",
        message="slow down",
        http_status=429,
        retry_after_seconds=7.0,
    )
    assert e.retry_after_seconds == 7.0
    assert e.http_status == 429


def test_typed_error_caught_by_except_apierror() -> None:
    try:
        raise NotFoundError(code="not_found", message="x")
    except APIError as e:
        assert isinstance(e, NotFoundError)
