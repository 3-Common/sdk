"""Events service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    NotFoundError,
    ServerError,
    ThreeCommon,
    ValidationError,
)
from threecommon.events import ListParams, RetrieveParams, UpdateBody


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


# ────────────────────────────────────────────────────────────────────────────
# Sync events
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events?pageSize=10&status=open",
        json={"data": [{"id": "evt_a", "name": "A", "status": "open"}], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.events.list(ListParams(status="open", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "evt_a"
    assert result.data[0].status == "open"
    assert result.has_more is False


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", json={"data": [], "hasMore": False})
    with _make_sync() as c:
        result = c.events.list()
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_123",
        json={"data": {"id": "evt_123", "name": "Demo"}},
    )
    with _make_sync() as c:
        ev = c.events.retrieve("evt_123")
    assert ev.id == "evt_123"
    assert ev.name == "Demo"


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_1?fields=id%2Cname",
        json={"data": {"id": "evt_1"}},
    )
    with _make_sync() as c:
        c.events.retrieve("evt_1", RetrieveParams(fields="id,name"))


def test_retrieve_requires_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.events.retrieve("")
        assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_missing",
        status_code=404,
        headers={"x-request-id": "req-404"},
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c:
        with pytest.raises(NotFoundError) as exc:
            c.events.retrieve("evt_missing")
        assert exc.value.request_id == "req-404"


def test_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_1",
        method="PATCH",
        match_json={"name": "Renamed"},
        json={"data": {"id": "evt_1", "name": "Renamed"}},
    )
    with _make_sync() as c:
        ev = c.events.update("evt_1", UpdateBody(name="Renamed"))
    assert ev.name == "Renamed"


def test_update_validates_id() -> None:
    with _make_sync() as c:
        with pytest.raises(ValidationError) as exc:
            c.events.update("", UpdateBody(name="x"))
        assert exc.value.code == "missing_id"


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events?page=0&status=open",
        json={"data": [{"id": "evt_1"}, {"id": "evt_2"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/events?page=1&status=open",
        json={"data": [{"id": "evt_3"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [ev.id for ev in c.events.list_auto_paginate(ListParams(status="open"))]
    assert ids == ["evt_1", "evt_2", "evt_3"]


def test_list_auto_paginate_surfaces_error(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events?page=0",
        status_code=500,
        json={"error": {"code": "internal_error", "message": "boom"}},
    )
    with _make_sync() as c:
        iter_ = c.events.list_auto_paginate()
        with pytest.raises(ServerError):
            next(iter_)


# ────────────────────────────────────────────────────────────────────────────
# Async events
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(url="http://test.local/v1/events", json={"data": [], "hasMore": False})
    async with _make_async() as c:
        r = await c.events.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_retrieve_404(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    async with _make_async() as c:
        with pytest.raises(NotFoundError):
            await c.events.retrieve("evt_missing")


@pytest.mark.asyncio
async def test_async_update_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events/evt_1",
        method="PATCH",
        match_json={"name": "X"},
        json={"data": {"id": "evt_1", "name": "X"}},
    )
    async with _make_async() as c:
        ev = await c.events.update("evt_1", UpdateBody(name="X"))
    assert ev.name == "X"


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/events?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/events?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [ev.id async for ev in c.events.list_auto_paginate()]
    assert ids == ["a", "b", "c"]


@pytest.mark.asyncio
async def test_async_update_validates() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.events.update("", UpdateBody(name="x"))
