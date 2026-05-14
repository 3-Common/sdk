from __future__ import annotations

from datetime import datetime, timedelta, timezone
from email.utils import format_datetime

import httpx
import pytest

from threecommon._core.parse import (
    parse_error_body,
    parse_retry_after,
    parse_success_body,
    request_id_of,
)


def test_parse_success_body_decodes_json() -> None:
    assert parse_success_body('{"id":"evt_1"}') == {"id": "evt_1"}


def test_parse_success_body_empty_returns_none() -> None:
    assert parse_success_body("") is None


def test_parse_success_body_malformed_returns_none() -> None:
    assert parse_success_body("{not-json") is None


def test_parse_error_body_full_envelope() -> None:
    body = '{"error":{"code":"not_found","message":"missing","details":{"id":"evt_1"}}}'
    code, message, details = parse_error_body(body)
    assert code == "not_found"
    assert message == "missing"
    assert details == {"id": "evt_1"}


def test_parse_error_body_empty_returns_zero_values() -> None:
    assert parse_error_body("") == ("", "", None)


def test_parse_error_body_malformed_returns_zero_values() -> None:
    assert parse_error_body("not json at all") == ("", "", None)


def test_parse_error_body_missing_error_key() -> None:
    assert parse_error_body('{"data":[]}') == ("", "", None)


@pytest.mark.parametrize(
    ("header", "want"),
    [
        ("5", 5.0),
        ("1.5", 1.5),
        ("0", 0.0),
        ("-3", 0.0),  # negative clamped
    ],
)
def test_parse_retry_after_delta_seconds(header: str, want: float) -> None:
    assert parse_retry_after(header) == pytest.approx(want)


def test_parse_retry_after_empty_returns_none() -> None:
    assert parse_retry_after("") is None
    assert parse_retry_after(None) is None


def test_parse_retry_after_malformed_returns_none() -> None:
    assert parse_retry_after("not a number") is None


def test_parse_retry_after_http_date_future() -> None:
    future = datetime.now(tz=timezone.utc) + timedelta(seconds=10)
    got = parse_retry_after(format_datetime(future, usegmt=True))
    assert got is not None
    assert 5 <= got <= 11


def test_parse_retry_after_http_date_past_returns_zero() -> None:
    past = datetime.now(tz=timezone.utc) - timedelta(hours=1)
    assert parse_retry_after(format_datetime(past, usegmt=True)) == 0.0


def test_request_id_of() -> None:
    response = httpx.Response(200, headers={"x-request-id": "req-1"})
    assert request_id_of(response) == "req-1"


def test_request_id_of_missing_returns_none() -> None:
    response = httpx.Response(200)
    assert request_id_of(response) is None
