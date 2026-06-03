"""Entitlements service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    ConflictError,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.entitlements import (
    ConsumeBody,
    GrantBody,
    ListParams,
    LookupParams,
    RetrieveParams,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


_SAMPLE = {
    "id": "ent_123",
    "hostId": "host_1",
    "contactId": "cnt_7",
    "featureKey": "api_calls",
    "balance": 100,
    "grants": [
        {
            "id": "grant_1",
            "source": "manual",
            "amount": 100,
            "remaining": 100,
            "addedAt": "2026-05-01T18:00:00.000Z",
        }
    ],
    "totalGranted": 100,
    "totalConsumed": 0,
    "metadata": {},
    "createdAt": "2026-05-01T18:00:00.000Z",
    "updatedAt": "2026-05-01T18:00:00.000Z",
}


# ────────────────────────────────────────────────────────────────────────────
# Sync entitlements
# ────────────────────────────────────────────────────────────────────────────


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements?featureKey=api_calls&minBalance=1",
        json={"data": [_SAMPLE], "hasMore": False},
    )
    with _make_sync() as c:
        result = c.entitlements.list(ListParams(feature_key="api_calls", min_balance=1))
    assert len(result.data) == 1
    assert result.data[0].id == "ent_123"
    assert result.data[0].balance == 100
    assert result.data[0].grants is not None
    assert result.data[0].grants[0].source == "manual"
    assert result.has_more is False


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.entitlements.list()
    assert result.data == []


def test_list_empty_params_omits_query(httpx_mock: HTTPXMock) -> None:
    # ListParams() with every field None encodes to no query string at all.
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements", json={"data": [], "hasMore": False}
    )
    with _make_sync() as c:
        result = c.entitlements.list(ListParams())
    assert result.data == []


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/ent_123",
        json={"data": _SAMPLE},
    )
    with _make_sync() as c:
        ent = c.entitlements.retrieve("ent_123")
    assert ent.id == "ent_123"
    assert ent.feature_key == "api_calls"


def test_retrieve_passes_fields(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/ent_1?fields=id%2Cbalance",
        json={"data": {"id": "ent_1", "balance": 5}},
    )
    with _make_sync() as c:
        ent = c.entitlements.retrieve("ent_1", RetrieveParams(fields="id,balance"))
    assert ent.balance == 5


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.entitlements.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/ent_missing",
        status_code=404,
        headers={"x-request-id": "req-404"},
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError) as exc:
        c.entitlements.retrieve("ent_missing")
    assert exc.value.request_id == "req-404"


def test_lookup_forwards_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/lookup?contactId=cnt_7&featureKey=api_calls",
        json={"data": _SAMPLE},
    )
    with _make_sync() as c:
        ent = c.entitlements.lookup(LookupParams(contact_id="cnt_7", feature_key="api_calls"))
    assert ent.id == "ent_123"
    assert ent.balance == 100


def test_lookup_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/lookup?contactId=cnt_7&featureKey=unknown",
        status_code=404,
        json={"error": {"code": "not_found", "message": "no record"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.entitlements.lookup(LookupParams(contact_id="cnt_7", feature_key="unknown"))


def test_grant_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/grants",
        method="POST",
        match_json={
            "contactId": "cnt_7",
            "featureKey": "api_calls",
            "amount": 50,
            "grantId": "grant_2",
        },
        json={"data": {**_SAMPLE, "balance": 150}},
    )
    with _make_sync() as c:
        ent = c.entitlements.grant(
            GrantBody(contact_id="cnt_7", feature_key="api_calls", amount=50, grant_id="grant_2")
        )
    assert ent.balance == 150


def test_consume_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/consume",
        method="POST",
        match_json={
            "contactId": "cnt_7",
            "featureKey": "api_calls",
            "amount": 1,
            "reason": "POST /generate",
        },
        json={"data": {**_SAMPLE, "balance": 99}},
    )
    with _make_sync() as c:
        ent = c.entitlements.consume(
            ConsumeBody(
                contact_id="cnt_7", feature_key="api_calls", amount=1, reason="POST /generate"
            )
        )
    assert ent.balance == 99


def test_consume_409_conflict(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/consume",
        method="POST",
        status_code=409,
        json={"error": {"code": "conflict", "message": "insufficient balance"}},
    )
    with _make_sync() as c, pytest.raises(ConflictError):
        c.entitlements.consume(
            ConsumeBody(contact_id="cnt_7", feature_key="api_calls", amount=9999)
        )


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements?page=0&featureKey=api_calls",
        json={"data": [{"id": "ent_a"}, {"id": "ent_b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements?page=1&featureKey=api_calls",
        json={"data": [{"id": "ent_c"}], "hasMore": False},
    )
    with _make_sync() as c:
        ids = [e.id for e in c.entitlements.list_auto_paginate(ListParams(feature_key="api_calls"))]
    assert ids == ["ent_a", "ent_b", "ent_c"]


# ────────────────────────────────────────────────────────────────────────────
# Async entitlements
# ────────────────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements", json={"data": [], "hasMore": False}
    )
    async with _make_async() as c:
        r = await c.entitlements.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_lookup(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/lookup?contactId=cnt_7&featureKey=api_calls",
        json={"data": _SAMPLE},
    )
    async with _make_async() as c:
        ent = await c.entitlements.lookup(LookupParams(contact_id="cnt_7", feature_key="api_calls"))
    assert ent.id == "ent_123"


@pytest.mark.asyncio
async def test_async_grant_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/grants",
        method="POST",
        match_json={
            "contactId": "cnt_7",
            "featureKey": "api_calls",
            "amount": 25,
            "grantId": "grant_3",
            "metadata": {"reason": "comp"},
        },
        json={"data": {**_SAMPLE, "balance": 125}},
    )
    async with _make_async() as c:
        ent = await c.entitlements.grant(
            GrantBody(
                contact_id="cnt_7",
                feature_key="api_calls",
                amount=25,
                grant_id="grant_3",
                metadata={"reason": "comp"},
            )
        )
    assert ent.balance == 125


@pytest.mark.asyncio
async def test_async_consume_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/consume",
        method="POST",
        match_json={"contactId": "cnt_7", "featureKey": "api_calls", "amount": 2},
        json={"data": {**_SAMPLE, "balance": 98}},
    )
    async with _make_async() as c:
        ent = await c.entitlements.consume(
            ConsumeBody(contact_id="cnt_7", feature_key="api_calls", amount=2)
        )
    assert ent.balance == 98


@pytest.mark.asyncio
async def test_async_consume_409(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/consume",
        method="POST",
        status_code=409,
        json={"error": {"code": "conflict", "message": "insufficient balance"}},
    )
    async with _make_async() as c:
        with pytest.raises(ConflictError):
            await c.entitlements.consume(
                ConsumeBody(contact_id="cnt_7", feature_key="api_calls", amount=9999)
            )


@pytest.mark.asyncio
async def test_async_retrieve_requires_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.entitlements.retrieve("")


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements?page=0",
        json={"data": [{"id": "a"}, {"id": "b"}], "hasMore": True},
    )
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements?page=1",
        json={"data": [{"id": "c"}], "hasMore": False},
    )
    async with _make_async() as c:
        ids = [e.id async for e in c.entitlements.list_auto_paginate()]
    assert ids == ["a", "b", "c"]


@pytest.mark.asyncio
async def test_async_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/entitlements/ent_123",
        json={"data": _SAMPLE},
    )
    async with _make_async() as c:
        ent = await c.entitlements.retrieve("ent_123")
    assert ent.id == "ent_123"
    assert ent.total_granted == 100


# ────────────────────────────────────────────────────────────────────────────
# Body guards — passing None violates the type but must raise, not crash.
# ────────────────────────────────────────────────────────────────────────────


def test_grant_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.entitlements.grant(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_consume_requires_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.entitlements.consume(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_grant_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.entitlements.grant(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"


@pytest.mark.asyncio
async def test_async_consume_requires_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.entitlements.consume(None)  # type: ignore[arg-type]
        assert exc.value.code == "missing_body"
