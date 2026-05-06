from __future__ import annotations

from threecommon._core.url import build_url


def test_trims_trailing_slash_and_adds_leading_slash() -> None:
    assert (
        build_url(base_url="https://api.3common.com//", api_path="/v1", path="events")
        == "https://api.3common.com/v1/events"
    )


def test_stable_query_ordering() -> None:
    got = build_url(
        base_url="https://api.3common.com",
        api_path="/v1",
        path="/events",
        query={"status": "open", "page": "0", "pageSize": "50"},
    )
    assert got == "https://api.3common.com/v1/events?page=0&pageSize=50&status=open"


def test_omits_empty_values() -> None:
    got = build_url(
        base_url="https://api.3common.com",
        api_path="/v1",
        path="/events",
        query={"status": "", "page": "0"},
    )
    assert got == "https://api.3common.com/v1/events?page=0"


def test_no_query_when_empty() -> None:
    assert (
        build_url(base_url="https://api.3common.com", api_path="/v1", path="/events", query=None)
        == "https://api.3common.com/v1/events"
    )
    assert (
        build_url(
            base_url="https://api.3common.com", api_path="/v1", path="/events", query={"a": ""}
        )
        == "https://api.3common.com/v1/events"
    )


def test_encodes_special_characters() -> None:
    got = build_url(
        base_url="https://api.3common.com",
        api_path="/v1",
        path="/events",
        query={"search": "a b&c"},
    )
    assert got == "https://api.3common.com/v1/events?search=a+b%26c"
